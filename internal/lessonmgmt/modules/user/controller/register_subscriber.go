package controller

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/application/consumers"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/user/infrastructure"

	"go.uber.org/zap"
)

func RegisterStudentCourseSlotInfoSubscriptionHandler(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	db database.Ext,
	userRepo infrastructure.UserRepo,
	studentSubRepo infrastructure.StudentSubscriptionRepo,
	studentSubAccessPathRepo infrastructure.StudentSubscriptionAccessPathRepo,
) error {
	handler := &consumers.StudentCourseSlotInfoHandler{
		Logger:                            logger,
		DB:                                db,
		JSM:                               jsm,
		UserRepo:                          userRepo,
		StudentSubscriptionRepo:           studentSubRepo,
		StudentSubscriptionAccessPathRepo: studentSubAccessPathRepo,
	}

	subscriber := &StudentCourseSlotInfoSubscriber{
		JSM:               jsm,
		Logger:            logger,
		SubscriberHandler: handler,
	}

	return subscriber.Subscribe()
}
