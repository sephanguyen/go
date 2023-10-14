package service

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
)

// Webhook handle when sending a new message
func (svc *conversationModifierServiceImpl) UpdateLatestMessage(ctx context.Context, message *domain.Message) error {
	return svc.ConversationRepo.UpdateLatestMessage(ctx, svc.DB, message)
}
