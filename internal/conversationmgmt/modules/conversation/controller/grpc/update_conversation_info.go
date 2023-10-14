package grpc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *ConversationModifierGRPC) UpdateConversationInfo(ctx context.Context, req *cpb.UpdateConversationInfoRequest) (*cpb.UpdateConversationInfoResponse, error) {
	conversationDomain := svc.mapReqToConversationDomain(req)

	_, err := svc.ConversationModifierServicePort.UpdateConversationInfo(ctx, conversationDomain)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ConversationModifierServicePort.UpdateConversationInfo: %v", err))
	}

	return &cpb.UpdateConversationInfoResponse{
		ConversationId: conversationDomain.ID,
		Name:           conversationDomain.Name,
		OptionalConfig: conversationDomain.OptionalConfig,
	}, nil
}

func (svc *ConversationModifierGRPC) mapReqToConversationDomain(req *cpb.UpdateConversationInfoRequest) *domain.Conversation {
	conversation := &domain.Conversation{
		ID:             req.ConversationId,
		Name:           req.Name,
		OptionalConfig: req.OptionalConfig,
	}
	return conversation
}
