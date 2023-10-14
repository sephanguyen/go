package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ConversationRepo interface {
	// Command
	UpsertConversation(ctx context.Context, db database.Ext, conversation *domain.Conversation) error
	UpdateLatestMessage(ctx context.Context, db database.Ext, message *domain.Message) error
	UpdateConversationInfo(ctx context.Context, db database.Ext, conversation *domain.Conversation) error

	// Query
	FindByIDsAndUserID(ctx context.Context, db database.Ext, userID string, conversationIDs []string) ([]*domain.Conversation, error)
	FindByIDs(ctx context.Context, db database.Ext, conversationIDs []string) ([]*domain.Conversation, error)
	FindByID(ctx context.Context, db database.Ext, conversationID string) (*domain.Conversation, error)
}
