package grpc

import (
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/port/service"
)

type ConversationModifierGRPC struct {
	ConversationModifierServicePort service.ConversationModifierService
}

func NewNotificationModifierGRPC(service service.ConversationModifierService) *ConversationModifierGRPC {
	return &ConversationModifierGRPC{
		ConversationModifierServicePort: service,
	}
}
