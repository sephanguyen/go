package bob

import (
	"fmt"

	"github.com/manabie-com/backend/internal/bob/configurations"
	"github.com/manabie-com/backend/internal/bob/repositories"
	"github.com/manabie-com/backend/internal/bob/subscriptions"
	"github.com/manabie-com/backend/internal/bob/support"
	"github.com/manabie-com/backend/internal/golibs/cloudconvert"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/golibs/whiteboard"
	lesson_allocation_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/allocation/infrastructure/repo"
	cls_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/infrastructure/repo"
	user_repo "github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure/repo"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

func NewTelemetry(c *configurations.Config) (*interceptors.PrometheusProvider, *tracesdk.TracerProvider, error) {
	pp, tp, err := interceptors.NewTelemetry(c.Common.RemoteTrace.OtelCollectorReceiver, c.Common.Name, 1)
	if !c.Common.StatsEnabled {
		pp = nil
	}

	if !c.Common.RemoteTrace.Enabled || len(c.Common.RemoteTrace.OtelCollectorReceiver) == 0 {
		tp = nil
		err = nil
	}

	return pp, tp, err
}

func initCloudConvertJobEventsSubscription(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.Ext,
	mediaRepo *repositories.MediaRepo,
	conversionTaskRepo *repositories.ConversionTaskRepo,
	conversionSvc *cloudconvert.Service,
) error {
	s := &subscriptions.CloudConvert{
		JSM:                jsm,
		Logger:             logger,
		DB:                 db,
		MediaRepo:          mediaRepo,
		ConversionTaskRepo: conversionTaskRepo,
		ConversionSvc:      conversionSvc,
	}
	err := s.Subscribe()
	if err != nil {
		return fmt.Errorf("subscriptions.CloudConvert.Subscribe: %v", err)
	}
	return err
}

func initInternalLiveLesson(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.Ext,
	lessonRepo *repositories.LessonRepo,
	whiteboard *whiteboard.Service,
) error {
	internalLiveLesson := &subscriptions.InternalLiveLessonCreated{
		JSM:           jsm,
		Logger:        logger,
		DB:            db,
		WhiteboardSvc: whiteboard,
		LessonRepo:    lessonRepo,
	}
	err := internalLiveLesson.Subscribe()
	if err != nil {
		return fmt.Errorf("internalLiveLesson.Subscribe: %w", err)
	}

	return nil
}

func initStudentCourserSubscription(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	wrapperConnection *support.WrapperDBConnection,
	studentSubscriptionRepo *repositories.StudentSubscriptionRepo,
	studentSubscriptionAccessPathRepo *repositories.StudentSubscriptionAccessPathRepo,
	userRepo *repositories.UserRepo,
	courseLocationScheduleRepo *cls_repo.CourseLocationScheduleRepo,
	lessonAllocationRepo *lesson_allocation_repo.LessonAllocationRepo,
	studentCourseRepo *user_repo.StudentCourseRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
) error {
	s := &subscriptions.StudentSubscription{
		JSM:                               jsm,
		Logger:                            logger,
		WrapperConnection:                 wrapperConnection,
		StudentSubscriptionRepo:           studentSubscriptionRepo,
		StudentSubscriptionAccessPathRepo: studentSubscriptionAccessPathRepo,
		CourseLocationScheduleRepo:        courseLocationScheduleRepo,
		LessonAllocationRepo:              lessonAllocationRepo,
		StudentCourseRepo:                 studentCourseRepo,
		Env:                               env,
		UnleashClientIns:                  unleashClientIns,
	}
	err := s.Subscribe()
	if err != nil {
		return fmt.Errorf("subscriptions.StudentSubs.Subscribe: %v", err)
	}
	return err
}
