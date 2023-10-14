package controller

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/virtualclassroom/configurations"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/consumers"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/infrastructure"

	"go.uber.org/zap"
)

func RegisterLessonDeletedSubscriber(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	wrapperConnection *support.WrapperDBConnection,
	recordRepo infrastructure.RecordedVideoRepo,
	fireStore infrastructure.FileStore,
	mediaModulePort infrastructure.MediaModulePort,
	cfg configurations.Config,
) error {
	l := LessonDeletedSubscription{
		Logger: logger,
		JSM:    jsm,
		LessonDeletedHandler: consumers.LessonDeletedHandler{
			Logger:            logger,
			WrapperConnection: wrapperConnection,
			JSM:               jsm,
			FileStore:         fireStore,
			RecordedVideoRepo: recordRepo,
			MediaModulePort:   mediaModulePort,
			Cfg:               cfg,
		},
	}
	return l.Subscribe()
}

func RegisterLessonDefaultChatStateSubscriber(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	wrapperConnection *support.WrapperDBConnection,
	lessonMemberRepo infrastructure.LessonMemberRepo,
) error {
	l := LessonDefaultChatStateSubscriber{
		Logger: logger,
		JSM:    jsm,
		SubscriberHandler: &consumers.LessonDefaultChatStateHandler{
			Logger:            logger,
			WrapperConnection: wrapperConnection,
			JSM:               jsm,
			LessonMemberRepo:  lessonMemberRepo,
		},
	}
	return l.Subscribe()
}

func RegisterCreateLiveLessonRoomSubscriber(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	wrapperConnection *support.WrapperDBConnection,
	lessonRepo infrastructure.VirtualLessonRepo,
	whiteboardService infrastructure.WhiteboardPort,
) error {
	c := CreateLiveLessonRoomSubscriber{
		Logger: logger,
		JSM:    jsm,
		SubscriberHandler: &consumers.CreateLiveLessonRoomHandler{
			Logger:            logger,
			WrapperConnection: wrapperConnection,
			JSM:               jsm,
			WhiteboardService: whiteboardService,
			LessonRepo:        lessonRepo,
		},
	}
	return c.Subscribe()
}

func RegisterUpcomingLiveLessonNotificationSubscriptionHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.Ext,
	wrapConnection *support.WrapperDBConnection,
	virtualLessonRepo infrastructure.VirtualLessonRepo,
	liveLessonSentNotificationRepo infrastructure.LiveLessonSentNotificationRepo,
	lessonMemberRepo infrastructure.LessonMemberRepo,
	studentParentRepo infrastructure.StudentParentRepo,
	userRepo infrastructure.UserRepo,
	env string,
	unleashClientIns unleashclient.ClientInstance,
) error {
	unleashClientIns.WaitForUnleashReady()
	isUnleashToggled, err := unleashClientIns.IsFeatureEnabled("BACKEND_Lesson_UpcomingLiveLessonNotification", env)
	if err != nil {
		logger.Error("VirtualClassroomService.UnleashClient cannot check for Unleash toggle:", zap.Error(err))
		isUnleashToggled = false
	}
	if isUnleashToggled {
		s := &UpcomingLiveLessonNotificationSubscription{
			Logger: logger,
			JSM:    jsm,
			SubscriberHandler: &consumers.UpcomingLiveLessonNotificationHandler{
				Logger:                         logger,
				JSM:                            jsm,
				BobDB:                          db,
				WrapperConnection:              wrapConnection,
				VirtualLessonRepo:              virtualLessonRepo,
				LiveLessonSentNotificationRepo: liveLessonSentNotificationRepo,
				LessonMemberRepo:               lessonMemberRepo,
				StudentParentRepo:              studentParentRepo,
				UserRepo:                       userRepo,
			},
		}
		return s.Subscribe()
	}
	return nil
}

func RegisterLessonUpdatedSubscriptionHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	wrapperConnection *support.WrapperDBConnection,
	liveLessonSentNotificationRepo infrastructure.LiveLessonSentNotificationRepo,
) error {
	s := &LessonUpdatedSubscription{
		Logger: logger,
		JSM:    jsm,
		SubscriberHandler: &consumers.LessonUpdatedHandler{
			JSM:                            jsm,
			WrapperConnection:              wrapperConnection,
			LiveLessonSentNotificationRepo: liveLessonSentNotificationRepo,
		},
	}
	return s.Subscribe()
}
