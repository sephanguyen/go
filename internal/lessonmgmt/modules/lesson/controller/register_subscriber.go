package controller

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/commands"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/consumers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/infrastructure"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/support"
	infrastructureClass "github.com/manabie-com/backend/internal/mastermgmt/modules/class/infrastructure"

	"go.uber.org/zap"
)

func RegisterLessonStudentSubscriptionHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	wrapperConnection *support.WrapperDBConnection,
	env string,
	unleashClientIns unleashclient.ClientInstance,
	lessonMemberRepo infrastructure.LessonMemberRepo,
	lessonReportRepo infrastructure.LessonReportRepo,
	reallocationRepo infrastructure.ReallocationRepo,
	lessonRepo infrastructure.LessonRepo,
	lessonCommandHandler commands.LessonCommandHandler,
) error {
	s := &LessonStudentSubscription{
		JSM:                  jsm,
		Logger:               logger,
		wrapperConnection:    wrapperConnection,
		Env:                  env,
		UnleashClientIns:     unleashClientIns,
		LessonMemberRepo:     lessonMemberRepo,
		LessonReportRepo:     lessonReportRepo,
		ReallocationRepo:     reallocationRepo,
		LessonRepo:           lessonRepo,
		LessonCommandHandler: lessonCommandHandler,
	}

	return s.Subscribe()
}

func RegisterLockLessonSubscriptionHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo infrastructure.LessonRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
) error {
	unleashClientIns.WaitForUnleashReady()
	isUnleashToggled, err := unleashClientIns.IsFeatureEnabled("BACKEND_Lesson_LockLessonSubscription", env)
	if err != nil {
		logger.Error("LessonManagementService.UnleashClient cannot check for Unleash toggle:", zap.Error(err))
		isUnleashToggled = false
	}
	if isUnleashToggled {
		s := &LockLessonSubscription{
			wrapperConnection: wrapperConnection,
			Logger:            logger,
			JSM:               jsm,
			LessonRepo:        lessonRepo,
		}
		return s.Subscribe()
	}
	return nil
}

func RegisterStudentClassSubscriptionHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.Ext,
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo infrastructure.LessonRepo,
	lessonMemberRepo infrastructure.LessonMemberRepo,
	classMemberRepo infrastructureClass.ClassMemberRepo,
	lessonReportRepo infrastructure.LessonReportRepo,
) error {
	subscriberHandler := &consumers.StudentChangeClassHandler{
		Logger:            logger,
		DB:                db,
		WrapperConnection: wrapperConnection,
		JSM:               jsm,
		LessonRepo:        lessonRepo,
		LessonMemberRepo:  lessonMemberRepo,
		ClassMemberRepo:   classMemberRepo,
		LessonReportRepo:  lessonReportRepo,
	}
	s := &StudentChangeClassSubscriber{
		Logger:            logger,
		JSM:               jsm,
		SubscriberHandler: subscriberHandler,
	}
	return s.Subscribe()
}

func RegisterReserveClassHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.Ext,
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo infrastructure.LessonRepo,
	lessonMemberRepo infrastructure.LessonMemberRepo,
	classMemberRepo infrastructureClass.ClassMemberRepo,
	lessonReportRepo infrastructure.LessonReportRepo,
) error {
	subscriberHandler := &consumers.ScheduleClassHandler{
		Logger:            logger,
		BobDB:             db,
		WrapperConnection: wrapperConnection,
		JSM:               jsm,
		LessonRepo:        lessonRepo,
		LessonMemberRepo:  lessonMemberRepo,
		ClassMemberRepo:   classMemberRepo,
		LessonReportRepo:  lessonReportRepo,
	}
	s := &ReserveClassSubscriber{
		Logger:            logger,
		JSM:               jsm,
		SubscriberHandler: subscriberHandler,
	}
	return s.Subscribe()
}
