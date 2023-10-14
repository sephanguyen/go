package service

import (
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/port/repository"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/port/service"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/infrastructure/postgres"
	"github.com/manabie-com/backend/internal/golibs/chatvendor"
	"github.com/manabie-com/backend/internal/golibs/database"

	"go.uber.org/zap"
)

type conversationReaderServiceImpl struct {
	Logger      *zap.Logger
	DB          database.Ext
	Environment string
	ChatVendor  chatvendor.ChatVendorClient

	ConversationRepo       repository.ConversationRepo
	ConversationMemberRepo repository.ConversationMemberRepo
	ChatVendorUserRepo     repository.ChatVendorUserRepo
}

func NewConversationReaderService(db database.Ext, logger *zap.Logger, env string, chatVendor chatvendor.ChatVendorClient) service.ConversationReaderService {
	return &conversationReaderServiceImpl{
		DB:          db,
		Logger:      logger,
		ChatVendor:  chatVendor,
		Environment: env,

		ConversationRepo:       &postgres.ConversationRepo{},
		ConversationMemberRepo: &postgres.ConversationMemberRepo{},
		ChatVendorUserRepo:     &postgres.AgoraUserRepo{},
	}
}
