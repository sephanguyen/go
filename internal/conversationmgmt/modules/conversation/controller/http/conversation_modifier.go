package http

import (
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/port/repository"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/port/service"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/zap"
)

type ConversationModifierHTTP struct {
	Logger                          *zap.Logger
	DB                              database.Ext
	ConversationModifierServicePort service.ConversationModifierService
	NotificationHandlerServicePort  service.NotificationHandler

	ConversationRepo   repository.ConversationRepo
	ChatVendorUserRepo repository.ChatVendorUserRepo
}

func NewNotificationModifierGTTP(db database.Ext, logger *zap.Logger, service service.ConversationModifierService, notificationSvc service.NotificationHandler) *ConversationModifierHTTP {
	return &ConversationModifierHTTP{
		DB:                              db,
		Logger:                          logger,
		ConversationModifierServicePort: service,
		NotificationHandlerServicePort:  notificationSvc,

		ConversationRepo:   &postgres.ConversationRepo{},
		ChatVendorUserRepo: &postgres.AgoraUserRepo{},
	}
}
