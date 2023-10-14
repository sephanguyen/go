package service

import (
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/port/service"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/zap"
)

type notificationHandlerServiceImpl struct {
	Logger      *zap.Logger
	DB          database.Ext
	Environment string
}

func NewNotificationHandlerService(db database.Ext, logger *zap.Logger, env string) service.NotificationHandler {
	return &notificationHandlerServiceImpl{
		DB:          db,
		Logger:      logger,
		Environment: env,
	}
}
