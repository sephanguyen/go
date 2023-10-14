package service

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
)

type ConversationReaderService interface {
	GetConversationsDetail(ctx context.Context, conversationIDs []string) ([]*domain.Conversation, error)
}
