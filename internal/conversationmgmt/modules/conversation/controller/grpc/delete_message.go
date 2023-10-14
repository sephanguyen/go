package grpc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"
)

func (svc *ConversationModifierGRPC) DeleteMessage(ctx context.Context, req *cpb.DeleteMessageRequest) (*cpb.DeleteMessageResponse, error) {
	messageDTO := &domain.Message{
		ConversationID:  req.ConversationId,
		VendorMessageID: req.VendorMessageId,
		IsDeleted:       true,
	}

	err := svc.ConversationModifierServicePort.DeleteMessage(ctx, messageDTO)
	if err != nil {
		return nil, fmt.Errorf("cannot delete message [%s]: [%+v]", messageDTO.VendorMessageID, err)
	}

	return &cpb.DeleteMessageResponse{}, nil
}
