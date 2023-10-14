package service

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
)

type ConversationModifierService interface {
	CreateConversation(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error)
	AddConversationMembers(ctx context.Context, conversationMember []domain.ConversationMember) error
	UpdateConversationInfo(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error)
	RemoveConversationMembers(ctx context.Context, conversationMembers []domain.ConversationMember) error
	UpdateLatestMessage(ctx context.Context, message *domain.Message) error
	DeleteMessage(ctx context.Context, message *domain.Message) error
}
