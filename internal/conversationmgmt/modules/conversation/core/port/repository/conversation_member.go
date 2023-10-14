package repository

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type ConversationMemberRepo interface {
	// Command
	BulkUpsert(ctx context.Context, db database.Ext, convoMembers []domain.ConversationMember) error

	// Query
	GetConversationMembersByUserID(ctx context.Context, db database.Ext, userID string, conversationIDs []string) ([]*domain.ConversationMember, error)
	CheckMembersExistInConversation(ctx context.Context, db database.Ext, conversationID string, conversationMemberIDs []string) ([]string, error)
}
