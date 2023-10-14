package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
)

// External gRPC (based on conversation member)
func (svc *conversationReaderServiceImpl) GetConversationsDetail(ctx context.Context, conversationIDs []string) ([]*domain.Conversation, error) {
	userID := interceptors.UserIDFromContext(ctx)
	if userID == "" {
		return nil, fmt.Errorf("error user_id doesn't exist in request")
	}

	conversationDomains, err := svc.ConversationRepo.FindByIDsAndUserID(ctx, svc.DB, userID, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("error when ConversationRepo.GetConversations: [%v]", err)
	}

	conversationMemberDomains, err := svc.ConversationMemberRepo.GetConversationMembersByUserID(ctx, svc.DB, userID, conversationIDs)
	if err != nil {
		return nil, fmt.Errorf("error when ConversationMemberRepo.GetConversationMembers: [%v]", err)
	}
	memberIDs := make([]string, 0)
	for _, conversationMemberDomain := range conversationMemberDomains {
		memberIDs = append(memberIDs, conversationMemberDomain.User.UserID)
	}
	uniqueMemberIDs := sliceutils.RemoveDuplicates(memberIDs)

	chatVendorUsers, err := svc.ChatVendorUserRepo.GetByUserIDs(ctx, svc.DB, uniqueMemberIDs)
	if err != nil {
		return nil, fmt.Errorf("error when ChatVendorUserRepo.GetByUserIDs: [%v]", err)
	}

	if len(chatVendorUsers) != len(uniqueMemberIDs) {
		return nil, fmt.Errorf("some users do not exist")
	}

	conversationWithMembers := mapMembersToConversations(conversationDomains, conversationMemberDomains, chatVendorUsers)
	return conversationWithMembers, nil
}

func mapMembersToConversations(conversationDomains []*domain.Conversation, conversationMemberDomains []*domain.ConversationMember, chatVendorUsers []*domain.ChatVendorUser) []*domain.Conversation {
	mapChatVendorUser := make(map[string]string)
	for _, chatVendorUser := range chatVendorUsers {
		mapChatVendorUser[chatVendorUser.UserID] = chatVendorUser.VendorUserID
	}

	mapConversationMemberDomain := make(map[string][]domain.ConversationMember)
	for _, conversationMemberDomain := range conversationMemberDomains {
		conversationMemberDomain.User.VendorUserID = mapChatVendorUser[conversationMemberDomain.User.UserID]
		mapConversationMemberDomain[conversationMemberDomain.ConversationID] = append(mapConversationMemberDomain[conversationMemberDomain.ConversationID], *conversationMemberDomain)
	}
	for _, conversationDomain := range conversationDomains {
		conversationDomain.Members = mapConversationMemberDomain[conversationDomain.ID]
	}
	return conversationDomains
}
