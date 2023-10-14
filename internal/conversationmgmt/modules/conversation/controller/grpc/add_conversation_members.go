package grpc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *ConversationModifierGRPC) AddConversationMembers(ctx context.Context, req *cpb.AddConversationMembersRequest) (*cpb.AddConversationMembersResponse, error) {
	err := validateAddConversationMemberRequest(req)
	if err != nil {
		return nil, err
	}

	conversationMembers := toConversationMemberDomain(req.GetMemberIds(),
		domain.WithConversationID(req.GetConversationId()),
		domain.WithStatus(common.ConversationMemberStatusActive),
	)

	err = svc.ConversationModifierServicePort.AddConversationMembers(ctx, conversationMembers)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("AddConversationMembers: %v", err))
	}

	return &cpb.AddConversationMembersResponse{}, nil
}

func validateAddConversationMemberRequest(req *cpb.AddConversationMembersRequest) error {
	if req.GetConversationId() == "" {
		return status.Errorf(codes.InvalidArgument, "Conversation ID is empty")
	}

	if len(req.GetMemberIds()) == 0 {
		return status.Errorf(codes.InvalidArgument, "Member list is empty")
	}

	if len(req.GetMemberIds()) > common.MaxMembersInRequest {
		return status.Errorf(codes.InvalidArgument, "Too many members in request")
	}

	return nil
}

func toConversationMemberDomain(memberIDs []string, opts ...domain.ConversationMemberOpt) []domain.ConversationMember {
	conversationMembers := make([]domain.ConversationMember, 0, len(memberIDs))
	for _, memberID := range memberIDs {
		cm := domain.ConversationMember{
			User: domain.ChatVendorUser{
				UserID: memberID,
			},
		}

		for _, opt := range opts {
			opt(&cm)
		}

		conversationMembers = append(conversationMembers, cm)
	}

	return conversationMembers
}
