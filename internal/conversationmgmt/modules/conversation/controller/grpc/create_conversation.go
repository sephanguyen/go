package grpc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *ConversationModifierGRPC) CreateConversation(ctx context.Context, req *cpb.CreateConversationRequest) (*cpb.CreateConversationResponse, error) {
	conversation := svc.toConversationDomain(req)
	conversation, err := svc.ConversationModifierServicePort.CreateConversation(ctx, conversation)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("CreateConversation: %v", err))
	}

	return svc.toPbCreateConversationResponse(conversation), nil
}

func (svc *ConversationModifierGRPC) toConversationDomain(req *cpb.CreateConversationRequest) *domain.Conversation {
	conversationMembers := make([]domain.ChatVendorUser, 0)
	for _, memberID := range req.MemberIds {
		conversationMembers = append(conversationMembers, domain.ChatVendorUser{
			UserID: memberID,
		})
	}
	conversation := domain.NewConversation("", req.Name, conversationMembers, req.OptionalConfig)

	return &conversation
}

func (svc *ConversationModifierGRPC) toPbCreateConversationResponse(conversation *domain.Conversation) *cpb.CreateConversationResponse {
	members := make([]*cpb.ChatVendorUser, 0)
	for _, convoMember := range conversation.Members {
		members = append(members, &cpb.ChatVendorUser{
			UserId:       convoMember.User.VendorUserID,
			VendorUserId: convoMember.User.UserID,
		})
	}

	return &cpb.CreateConversationResponse{
		ConversationId: conversation.ID,
		Name:           conversation.Name,
		Members:        members,
		OptionalConfig: conversation.OptionalConfig,
	}
}
