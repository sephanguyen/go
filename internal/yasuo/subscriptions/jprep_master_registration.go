package subscriptions

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/bob/repositories"
	enigma_entities "github.com/manabie-com/backend/internal/enigma/entities"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/yasuo/services"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	nats_org "github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type JprepMasterRegistration struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	SubsJS []nats_org.Subscription

	CourseService interface {
		SyncCourse(ctx context.Context, req []*npb.EventMasterRegistration_Course) error
		SyncLiveLesson(ctx context.Context, req []*npb.EventMasterRegistration_Lesson) error
		SyncAcademicYear(ctx context.Context, req []*npb.EventMasterRegistration_AcademicYear) error
		UpdateAcademicYear(ctx context.Context, req []*repositories.UpdateAcademicYearOpts) error
	}

	ClassService interface {
		SyncClass(ctx context.Context, req []*npb.EventMasterRegistration_Class) error
	}

	PartnerSyncDataLogService interface {
		UpdateLogStatus(ctx context.Context, id, status string) error
	}

	ConfigService interface {
		UpsertConfig(ctx context.Context, upsertReq *services.UpsertConfig) error
	}
}

func (j *JprepMasterRegistration) Subscribe() error {
	option := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
			nats.DeliverNew(),
		},
	}

	optionLiveLesson := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncMasterRegistration, constants.DurableSyncLiveLesson),
			nats.DeliverSubject(constants.DeliverSyncMasterRegistrationLiveLesson)),
	}
	_, err := j.JSM.QueueSubscribe(constants.SubjectSyncMasterRegistration,
		constants.QueueSyncLiveLesson, optionLiveLesson, j.syncLiveLesson)
	if err != nil {
		return fmt.Errorf("syncLiveLesson.Subscribe: %w", err)
	}

	optionClass := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncMasterRegistration, constants.DurableSyncClass),
			nats.DeliverSubject(constants.DeliverSyncMasterRegistrationClass)),
	}
	_, err = j.JSM.QueueSubscribe(constants.SubjectSyncMasterRegistration,
		constants.QueueSyncClass, optionClass, j.syncClassHandler)
	if err != nil {
		return fmt.Errorf("syncClassSub.Subscribe: %w", err)
	}

	optionCourseAcademic := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncMasterRegistration, constants.DurableSyncCourseAcademic),
			nats.DeliverSubject(constants.DeliverSyncMasterRegistrationCourseAcademic)),
	}
	_, err = j.JSM.QueueSubscribe(constants.SubjectSyncMasterRegistration,
		constants.QueueSyncCourseAcademic, optionCourseAcademic, j.syncCourseAcademicHandler)
	if err != nil {
		return fmt.Errorf("syncClassSub.Subscribe: %w", err)
	}

	optionCourse := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncMasterRegistration, constants.DurableSyncCourse),
			nats.DeliverSubject(constants.DeliverSyncMasterRegistrationCourse)),
	}
	_, err = j.JSM.QueueSubscribe(constants.SubjectSyncMasterRegistration,
		constants.QueueSyncCourse, optionCourse, j.syncCourseHandler)
	if err != nil {
		return fmt.Errorf("syncCourseSub.Subscribe: %w", err)
	}

	optionAcademic := nats.Option{
		JetStreamOptions: append(option.JetStreamOptions,
			nats.Bind(constants.StreamSyncMasterRegistration, constants.DurableSyncAcademicYear),
			nats.DeliverSubject(constants.DeliverSyncMasterRegistrationAcademicYear)),
	}
	_, err = j.JSM.QueueSubscribe(constants.SubjectSyncMasterRegistration,
		constants.QueueSyncAcademicYear, optionAcademic, j.syncAcademicYear)
	if err != nil {
		return fmt.Errorf("syncAcademicYear.Subscribe: %w", err)
	}

	return nil
}

func (j *JprepMasterRegistration) syncClassHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventMasterRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncClassHandler proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepMasterRegistration.syncClassHandler",
		zap.String("signature", req.Signature),
	)
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncClassHandler update log status to processing: %w", err)
	}
	if len(req.Classes) == 0 {
		return false, fmt.Errorf("syncClassHandler length of classes = 0")
	}
	if err := nats.ChunkHandler(len(req.Classes), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.ClassService.SyncClass(ctx, req.Classes[start:end])
	}); err != nil {
		return true, fmt.Errorf("syncClassHandler err SyncClass: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncClassHandler update log status to success: %w", err)
	}
	return false, nil
}

func (j *JprepMasterRegistration) syncCourseHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventMasterRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncCourseHandler proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepMasterRegistration.syncCourseHandler",
		zap.String("signature", req.Signature),
	)
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncCourseHandler update log status to processing: %w", err)
	}
	if len(req.Courses) == 0 {
		return false, fmt.Errorf("syncCourseHandler length of courses = 0")
	}
	if err := nats.ChunkHandler(len(req.Courses), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.CourseService.SyncCourse(ctx, req.Courses[start:end])
	}); err != nil {
		return true, fmt.Errorf("syncCourseHandler err SyncCourse: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncCourseHandler update log status to success: %w", err)
	}

	return false, nil
}

func (j *JprepMasterRegistration) syncLiveLesson(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	if err := j.ConfigService.UpsertConfig(ctx, &services.UpsertConfig{
		Key:     "JprepMasterRegistrationLastTime",
		Group:   "yasuo",
		Country: "COUNTRY_MASTER",
		Value:   time.Now().GoString(),
	}); err != nil {
		j.Logger.Info("ConfigService.UpsertConfig upsert last time",
			zap.String("error", err.Error()),
		)
	}
	var req npb.EventMasterRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncLiveLesson proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepMasterRegistration.syncLiveLesson",
		zap.String("signature", req.Signature),
	)
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncLiveLesson update log status to processing: %w", err)
	}
	if err := nats.ChunkHandler(len(req.Lessons), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.CourseService.SyncLiveLesson(ctx, req.Lessons[start:end])
	}); err != nil {
		return true, fmt.Errorf("syncLiveLesson err SyncLiveLesson: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepUserRegistration.syncLiveLesson update log status to success: %w", err)
	}

	return false, nil
}

func (j *JprepMasterRegistration) syncAcademicYear(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventMasterRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncAcademicYear proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepMasterRegistration.syncAcademicYear",
		zap.String("signature", req.Signature),
	)
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncLiveLesson update log status to processing: %w", err)
	}
	if err := nats.ChunkHandler(len(req.AcademicYears), constants.MaxRecordProcessPertime, func(start, end int) error {
		return j.CourseService.SyncAcademicYear(ctx, req.AcademicYears[start:end])
	}); err != nil {
		return true, fmt.Errorf("syncAcademicYear err SyncAcademicYear: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncAcademicYear: %w", err)
	}

	return false, nil
}

func (j *JprepMasterRegistration) syncCourseAcademicHandler(ctx context.Context, data []byte) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var req npb.EventMasterRegistration
	if err := proto.Unmarshal(data, &req); err != nil {
		return false, fmt.Errorf("syncCourseAcademicHandler proto.Unmarshal: %w", err)
	}
	j.Logger.Info("JprepMasterRegistration.syncCourseAcademicHandler",
		zap.String("signature", req.Signature),
	)
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusProcessing)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncCourseAcademicHandler update log status to processing: %w", err)
	}
	if len(req.Classes) == 0 {
		return false, fmt.Errorf("syncCourseAcademicHandler length of classes = 0")
	}
	if err := nats.ChunkHandler(len(req.Classes), constants.MaxRecordProcessPertime, func(start, end int) error {
		opts := []*repositories.UpdateAcademicYearOpts{}
		for _, c := range req.Classes[start:end] {
			opts = append(opts, &repositories.UpdateAcademicYearOpts{
				CourseID:       c.CourseId,
				AcademicYearID: c.AcademicYearId,
			})
		}

		return j.CourseService.UpdateAcademicYear(ctx, opts)
	}); err != nil {
		return true, fmt.Errorf("syncCourseAcademicHandler err SyncAcademicYear: %w", err)
	}
	if err := j.PartnerSyncDataLogService.UpdateLogStatus(ctx, req.LogId, string(enigma_entities.StatusSuccess)); err != nil {
		return true, fmt.Errorf("JprepMasterRegistration.syncCourseAcademicHandler update log status to success: %w", err)
	}

	return false, nil
}
