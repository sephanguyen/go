package subscriptions

import (
	"context"
	"fmt"
	"math"
	"time"

	enigma_entities "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/services"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type JprepSyncUserCourse struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement

	CourseService interface {
		SyncStudentLesson(ctx context.Context, req []*npb.EventSyncUserCourse_StudentLesson) error
		CourseIDsByClass(ctx context.Context, classID []int32) (mapByClass map[int32][]string, err error)
	}

	PartnerSyncDataLogService interface {
		UpdateLogStatus(ctx context.Context, id, status string) error
		GetLogBySignature(ctx context.Context, signature string) (*enigma_entities.PartnerSyncDataLog, error)
	}

	ConfigService interface {
		UpsertConfig(ctx context.Context, upsertReq *services.UpsertConfig) error
	}
}

func (j *JprepSyncUserCourse) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
	}

	optionStudentLesson := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncUserCourse, constants.DurableJPREPSyncUserCourseNatsJS),
			nats.DeliverSubject(constants.DeliverSyncUserCourse)),
	}
	_, err := j.JSM.QueueSubscribe(constants.SubjectJPREPSyncUserCourseNatsJS,
		constants.QueueJPREPSyncUserCourseNatsJS, optionStudentLesson, j.syncStudentLesson)
	if err != nil {
		return fmt.Errorf("syncStudentLesson.Subscribe: %w", err)
	}

	optionSyncStudentPackage := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncUserRegistration, constants.DurableSyncStudentPackage),
			nats.DeliverSubject(constants.DeliverSyncUserRegistrationStudentPackage)),
	}
	_, err = j.JSM.QueueSubscribe(constants.SubjectUserRegistrationNatsJS,
		constants.QueueSyncStudentPackage, optionSyncStudentPackage, j.syncStudentPackage)
	if err != nil {
		return fmt.Errorf("syncStudentPackage.Subscribe: %w", err)
	}

	return nil
}

func (j *JprepSyncUserCourse) syncStudentLesson(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	/* No need for now
	if err := j.ConfigService.UpsertConfig(ctx, &services.UpsertConfig{
		Key:     "JprepSyncUserCourseLastTime",
		Group:   "yasuo",
		Country: "COUNTRY_MASTER",
		Value:   time.Now().GoString(),
	}); err != nil {
		j.Logger.Info("ConfigService.UpsertConfig upsert last time",
			zap.String("error", err.Error()),
		)
	}
	*/
	var req npb.EventSyncUserCourse
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("err SyncStudentLesson proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepSyncUserCourse.syncStudentLesson",
		zap.String("signature", req.Signature),
	)
	partnerSyncDataLog, err := j.PartnerSyncDataLogService.GetLogBySignature(ctx, req.Signature)
	if err != nil {
		return true, fmt.Errorf("PartnerSyncDataLogService.GetLogBySignature err: %w", err)
	}
	// Cheat: Only receive messages last 1 hour (LT-21620)
	if math.Abs((time.Since(partnerSyncDataLog.UpdatedAt.Time)).Minutes()) > 60 {
		j.Logger.Info("Nats receive old message over 1 hour ago",
			zap.String("signature", req.Signature),
		)
		return false, nil
	}

	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepSyncUserCourse.syncStudentLesson update log status to processing: %w", err)
	}
	if err := nats.ChunkHandler(len(req.StudentLessons), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.CourseService.SyncStudentLesson(ctx, req.StudentLessons[start:end])
	}); err != nil {
		return true, fmt.Errorf("err SyncStudentLesson: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepSyncUserCourse.syncStudentLesson update log status to success: %w", err)
	}

	return false, nil
}

func (j *JprepSyncUserCourse) syncStudentPackage(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventUserRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncStudentPackage proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepSyncUserCourse.syncStudentPackage",
		zap.String("signature", req.Signature),
	)
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepSyncUserCourse.syncStudentPackage update log status to processing: %w", err)
	}
	if len(req.Students) == 0 {
		return false, nil
	}
	err := nats.ChunkHandler(len(req.Students), constants.MaxRecordProcessPertime, func(start, end int) error {
		students := req.Students[start:end]
		classIDs := []int32{}
		for _, s := range students {
			for _, p := range s.Packages {
				classIDs = append(classIDs, int32(p.ClassId))
			}
		}

		mapByClass, err := j.CourseService.CourseIDsByClass(ctx, classIDs)
		if err != nil {
			return fmt.Errorf("err j.CourseService.CourseIDsByClass: %w", err)
		}

		event := &npb.EventSyncStudentPackage{}
		for _, s := range students {
			packages := []*npb.EventSyncStudentPackage_Package{}
			for _, p := range s.Packages {
				courseIDs, ok := mapByClass[int32(p.ClassId)]
				if !ok {
					continue
				}

				packages = append(packages, &npb.EventSyncStudentPackage_Package{
					CourseIds: courseIDs,
					StartDate: p.StartDate,
					EndDate:   p.EndDate,
				})
			}

			event.StudentPackages = append(event.StudentPackages, &npb.EventSyncStudentPackage_StudentPackage{
				ActionKind: s.ActionKind,
				StudentId:  s.StudentId,
				Packages:   packages,
			})
		}

		data, _ := proto.Marshal(event)
		_, err = j.JSM.PublishAsyncContext(ctx, constants.SubjectSyncStudentPackage, data)
		if err != nil {
			return fmt.Errorf("err PublishAsync: %w", err)
		}

		if len(classIDs) != len(mapByClass) {
			classNotFound := []int32{}
			for _, id := range classIDs {
				_, ok := mapByClass[id]
				if !ok {
					classNotFound = append(classNotFound, id)
				}
			}

			if len(classNotFound) > 0 {
				return fmt.Errorf("class has not create yet: %v", classNotFound)
			}
		}

		return nil
	})
	if err != nil {
		return true, fmt.Errorf("err SyncStudentPackage: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepSyncUserCourse.syncStudentPackage update log status to success: %w", err)
	}

	return false, nil
}
