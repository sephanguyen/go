package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
)

// Internal gRPC
func (svc *conversationModifierServiceImpl) UpdateConversationInfo(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error) {
	err := svc.ConversationRepo.UpdateConversationInfo(ctx, svc.DB, conversation)
	if err != nil {
		return nil, fmt.Errorf("error when ConversationRepo.UpdateConversationInfo: [%v]", err)
	}
	return conversation, nil
}
