package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/tom/configurations"
	bobproto "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/zap"
)

type LessonEventSubscription struct {
	Config             *configurations.Config
	Logger             *zap.Logger
	JSM                nats.JetStreamManagement
	LessonChatModifier interface {
		HandleEventCreateLesson(ctx context.Context, msg *bobproto.EvtLesson_CreateLessons) error
		HandleEventJoinLesson(ctx context.Context, msg *bobproto.EvtLesson_JoinLesson) (bool, error)
		HandleEventLeaveLesson(ctx context.Context, msg *bobproto.EvtLesson_LeaveLesson) (bool, error)
		HandleEventEndLiveLesson(ctx context.Context, msg *bobproto.EvtLesson_EndLiveLesson) error
		HandleEventUpdateLesson(ctx context.Context, msg *bobproto.EvtLesson_UpdateLesson) error
	}
}

func (rcv *LessonEventSubscription) Subscribe() error {
	optsLessonSub := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLesson, constants.DurableLesson),
			nats.DeliverSubject(constants.DeliverLessonEvent),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}

	optsLessonChatSub := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLessonChat, constants.DurableLessonChat),
			nats.DeliverSubject(constants.DeliverSyncLessonChat),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "rcv.HandlerNatsMessageLesson",
	}
	_, err := rcv.JSM.QueueSubscribe(constants.SubjectLesson, constants.QueueLesson, optsLessonSub, rcv.HandlerNatsMessageLesson)
	if err != nil {
		return fmt.Errorf("subLesson.QueueSubscribe: %w", err)
	}
	_, err = rcv.JSM.QueueSubscribe(constants.SubjectLessonChatSynced, constants.QueueLessonChat, optsLessonChatSub, rcv.HandlerNatsMessageLesson)
	if err != nil {
		return fmt.Errorf("subSyncLessonConversation.QueueSubscribe: %w", err)
	}

	return nil
}

func (rcv *LessonEventSubscription) HandlerNatsMessageLesson(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &bobproto.EvtLesson{}
	err := req.Unmarshal(data)
	if err != nil {
		rcv.Logger.Error(err.Error())
		return true, err
	}
	switch req.Message.(type) {
	case *bobproto.EvtLesson_CreateLessons_:
		msg := req.GetCreateLessons()
		err := rcv.LessonChatModifier.HandleEventCreateLesson(ctx, msg)
		if err != nil {
			rcv.Logger.Error("err rcv.handleEventCreateLesson", zap.Error(err))
			return true, err
		}
	case *bobproto.EvtLesson_UpdateLesson_:
		msg := req.GetUpdateLesson()

		ctx, span := interceptors.StartSpan(ctx, "HandleEventUpdateLesson")
		defer span.End()
		err := rcv.LessonChatModifier.HandleEventUpdateLesson(ctx, msg)
		if err != nil {
			rcv.Logger.Error("err rcv.handleEventUpdateLesson", zap.Error(err))
			return true, err
		}
	case *bobproto.EvtLesson_JoinLesson_:
		msg := req.GetJoinLesson()
		retry, err := rcv.LessonChatModifier.HandleEventJoinLesson(ctx, msg)
		if err != nil {
			rcv.Logger.Error("err rcv.handleEventJoinLesson", zap.Error(err))
			return retry, err
		}
	case *bobproto.EvtLesson_LeaveLesson_:
		msg := req.GetLeaveLesson()
		retry, err := rcv.LessonChatModifier.HandleEventLeaveLesson(ctx, msg)
		if err != nil {
			rcv.Logger.Error("err rcv.handleEventJoinLesson", zap.Error(err))
			return retry, err
		}
	case *bobproto.EvtLesson_EndLiveLesson_:
		msg := req.GetEndLiveLesson()
		err := rcv.LessonChatModifier.HandleEventEndLiveLesson(ctx, msg)
		if err != nil {
			rcv.Logger.Error("rcv.LessonChatModifier.HandleEventEndLiveLesson", zap.Error(err))
			return true, err
		}
	}
	return false, nil
}
