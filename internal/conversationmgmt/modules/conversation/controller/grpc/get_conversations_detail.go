package grpc

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (svc *ConversationReaderGRPC) GetConversationsDetail(ctx context.Context, req *cpb.GetConversationsDetailRequest) (*cpb.GetConversationsDetailResponse, error) {
	if len(req.GetConversationIds()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Conversation IDs is empty")
	}
	conversationDomains, err := svc.ConversationReaderServicePort.GetConversationsDetail(ctx, req.GetConversationIds())
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("ConversationReaderServicePort.GetConversationsDetail: %v", err))
	}

	res, err := svc.mapConversationDomainsToPbGetConversationsDetailResponse(conversationDomains)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("mapConversationDomainsToPbGetConversationsDetailResponse: %v", err))
	}

	return res, nil
}

func (svc *ConversationReaderGRPC) mapConversationDomainsToPbGetConversationsDetailResponse(conversationDomains []*domain.Conversation) (*cpb.GetConversationsDetailResponse, error) {
	conversations := make([]*cpb.Conversation, 0, len(conversationDomains))
	for _, conversationDomain := range conversationDomains {
		conversationMembers := svc.mapConversationMemberDomainsToPbConversationMember(conversationDomain.Members)

		latestMessage, err := conversationDomain.LatestMessage.ToBytes()
		if err != nil {
			return nil, err
		}

		conversation := &cpb.Conversation{
			ConversationId: conversationDomain.ID,
			Name:           conversationDomain.Name,
			LatestMessage:  latestMessage,
			OptionalConfig: conversationDomain.OptionalConfig,
			CreatedAt:      timestamppb.New(conversationDomain.CreatedAt),
			UpdatedAt:      timestamppb.New(conversationDomain.UpdatedAt),
			Members:        conversationMembers,
		}

		if conversationDomain.LatestMessageSentTime != nil {
			conversation.LatestMessageSentTime = timestamppb.New(*conversationDomain.LatestMessageSentTime)
		}

		conversations = append(conversations, conversation)
	}
	return &cpb.GetConversationsDetailResponse{
		Conversations: conversations,
	}, nil
}

func (svc *ConversationReaderGRPC) mapConversationMemberDomainsToPbConversationMember(conversationMemberDomains []domain.ConversationMember) []*cpb.ConversationMember {
	conversationMembers := make([]*cpb.ConversationMember, 0, len(conversationMemberDomains))
	for _, conversationMemberDomain := range conversationMemberDomains {
		conversationMember := &cpb.ConversationMember{
			ConversationMemberId: conversationMemberDomain.ID,
			ConversationId:       conversationMemberDomain.ConversationID,
			User: &cpb.ChatVendorUser{
				UserId:       conversationMemberDomain.User.UserID,
				VendorUserId: conversationMemberDomain.User.VendorUserID,
			},
			Status:    cpb.ConversationMemberStatus(cpb.ConversationMemberStatus_value[string(conversationMemberDomain.Status)]),
			SeenAt:    timestamppb.New(conversationMemberDomain.SeenAt),
			CreatedAt: timestamppb.New(conversationMemberDomain.CreatedAt),
			UpdatedAt: timestamppb.New(conversationMemberDomain.UpdatedAt),
		}
		conversationMembers = append(conversationMembers, conversationMember)
	}
	return conversationMembers
}
