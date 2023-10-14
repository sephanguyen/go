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

func (svc *ConversationModifierGRPC) RemoveConversationMembers(ctx context.Context, req *cpb.RemoveConversationMembersRequest) (*cpb.RemoveConversationMembersResponse, error) {
	err := validateRemoveConversationMembersRequest(req)
	if err != nil {
		return nil, err
	}

	conversationMembers := toConversationMemberDomain(req.GetMemberIds(),
		domain.WithConversationID(req.GetConversationId()),
		domain.WithStatus(common.ConversationMemberStatusInActive),
	)

	err = svc.ConversationModifierServicePort.RemoveConversationMembers(ctx, conversationMembers)

	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("RemoveConversationMembers: %v", err))
	}

	return &cpb.RemoveConversationMembersResponse{}, nil
}

func validateRemoveConversationMembersRequest(req *cpb.RemoveConversationMembersRequest) error {
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
