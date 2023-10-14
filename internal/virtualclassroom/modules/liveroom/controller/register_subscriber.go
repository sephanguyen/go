package controller

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/application/consumers"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/liveroom/infrastructure"

	"go.uber.org/zap"
)

func RegisterLiveRoomSubscriber(
	jsm nats.JetStreamManagement,
	logger *zap.Logger,
	lessonmgmtDB database.Ext,
	liveRoomMemberStateRepo infrastructure.LiveRoomMemberStateRepo,
) error {
	l := LiveRoomSubscriber{
		Logger: logger,
		JSM:    jsm,
		SubscriberHandler: &consumers.LiveRoomHandler{
			Logger:                  logger,
			LessonmgmtDB:            lessonmgmtDB,
			JSM:                     jsm,
			LiveRoomMemberStateRepo: liveRoomMemberStateRepo,
		},
	}
	return l.Subscribe()
}
