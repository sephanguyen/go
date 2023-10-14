package grpc

import "github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/port/service"

type ConversationReaderGRPC struct {
	ConversationReaderServicePort service.ConversationReaderService
}

func NewConversationReaderGRPC(service service.ConversationReaderService) *ConversationReaderGRPC {
	return &ConversationReaderGRPC{
		ConversationReaderServicePort: service,
	}
}
