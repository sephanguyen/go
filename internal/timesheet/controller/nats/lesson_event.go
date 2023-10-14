package nats

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/timesheet/domain/constant"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type LessonEventSubscription struct {
	Logger        *zap.Logger
	JSM           nats.JetStreamManagement
	UnleashClient unleashclient.ClientInstance
	Env           string

	LessonNatsService interface {
		HandleEventLessonUpdate(ctx context.Context, msg *bpb.EvtLesson_UpdateLesson) error
		HandleEventLessonCreate(ctx context.Context, msg *bpb.EvtLesson_CreateLessons) error
		HandleEventLessonDelete(ctx context.Context, msg *bpb.EvtLesson_DeletedLessons) error
	}

	MastermgmtConfigurationService interface {
		CheckPartnerTimesheetServiceIsOnWithoutToken(ctx context.Context) (bool, error)
	}
}

func (s *LessonEventSubscription) Subscribe() error {
	s.UnleashClient.WaitForUnleashReady()
	isAutocreateEnable, err := s.UnleashClient.IsFeatureEnabled(constant.FeatureToggleAutoCreate, s.Env)

	if err != nil {
		s.Logger.Error("err Subscribe UnleashClient.IsFeatureEnabled failed", zap.Error(err))
		return fmt.Errorf("%s unleashClient.IsFeatureEnabled: %w", constant.FeatureToggleAutoCreate, err)
	}
	log.Printf("unleashClient.IsFeatureEnabled: %v", isAutocreateEnable)

	if isAutocreateEnable {
		optsLessonSub := nats.Option{
			JetStreamOptions: []nats.JSSubOption{
				nats.ManualAck(),
				nats.Bind(constants.StreamLesson, constants.DurableTimesheetLesson),
				nats.DeliverSubject(constants.DeliverTimesheetLessonEvent),
				nats.MaxDeliver(10),
				nats.AckWait(30 * time.Second),
			},
		}

		_, err = s.JSM.QueueSubscribe(constants.SubjectLesson, constants.QueueTimesheetLesson, optsLessonSub, s.HandlerNatsMessageLesson)
		if err != nil {
			s.Logger.Error("err Subscribe JSM.QueueSubscribe failed", zap.Error(err))
			return fmt.Errorf("subLesson.QueueSubscribe SubjectLesson: %w", err)
		}
	}

	return nil
}

func (s *LessonEventSubscription) HandlerNatsMessageLesson(ctx context.Context, data []byte) (bool, error) {

	serviceStatus, err := s.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOnWithoutToken(ctx)
	if err != nil {
		return true, status.Errorf(codes.Internal, err.Error())
	}
	if !serviceStatus {
		return false, status.Errorf(codes.PermissionDenied, "don't have permission to modify timesheet")
	}

	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()

	req := &bpb.EvtLesson{}
	err = proto.Unmarshal(data, req)
	if err != nil {
		s.Logger.Error(err.Error())
		return true, err
	}
	switch req.Message.(type) {
	case *bpb.EvtLesson_UpdateLesson_:
		msg := req.GetUpdateLesson()
		ctx, span := interceptors.StartSpan(ctx, "HandlerNatsMessageLessonUpdate")
		defer span.End()
		err := s.LessonNatsService.HandleEventLessonUpdate(ctx, msg)
		if err != nil {
			s.Logger.Error("err HandlerNatsMessageLesson HandleEventLessonUpdate failed", zap.Error(err))
			return true, err
		}
	case *bpb.EvtLesson_CreateLessons_:
		msg := req.GetCreateLessons()
		ctx, span := interceptors.StartSpan(ctx, "HandlerNatsMessageLessonCreate")
		defer span.End()
		err := s.LessonNatsService.HandleEventLessonCreate(ctx, msg)
		if err != nil {
			s.Logger.Error("err HandlerNatsMessageLesson HandleEventLessonCreate", zap.Error(err))
			return true, err
		}
	case *bpb.EvtLesson_DeletedLessons_:
		msg := req.GetDeletedLessons()
		ctx, span := interceptors.StartSpan(ctx, "HandlerNatsMessageLessonDelete")
		defer span.End()
		err := s.LessonNatsService.HandleEventLessonDelete(ctx, msg)
		if err != nil {
			s.Logger.Error("err HandlerNatsMessageLesson HandleEventLessonDelete", zap.Error(err))
			return true, err
		}
	default:
		s.Logger.Warn("HandlerNatsMessageLesson Unknown message lesson type", zap.Error(err))
		return false, nil
	}
	return false, nil
}
