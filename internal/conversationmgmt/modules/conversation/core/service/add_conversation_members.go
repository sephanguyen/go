package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	chatvendor_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
)

// Internal gRPC
func (svc *conversationModifierServiceImpl) AddConversationMembers(ctx context.Context, conversationMembers []domain.ConversationMember) error {
	conversationID := conversationMembers[0].ConversationID
	// check if conversation exist
	conversations, err := svc.ConversationRepo.FindByIDs(ctx, svc.DB, []string{conversationID})
	if err != nil {
		return fmt.Errorf("failed FindByIds: %+v", err)
	}

	if len(conversations) == 0 {
		return fmt.Errorf("conversation not found")
	}

	memberIDs := []string{}
	for _, convMember := range conversationMembers {
		memberIDs = append(memberIDs, convMember.User.UserID)
	}

	// check members existed in agora
	chatVendorUsers, err := svc.ChatVendorUserRepo.GetByUserIDs(ctx, svc.DB, memberIDs)
	if err != nil {
		return fmt.Errorf("failed GetByUserIDs: %+v", err)
	}

	if len(chatVendorUsers) != len(memberIDs) {
		return fmt.Errorf("some users do not exist")
	}

	// check members have existed in the current conversation
	existedMemberIDs, err := svc.ConversationMemberRepo.CheckMembersExistInConversation(ctx, svc.DB, conversationID, memberIDs)
	if err != nil {
		return fmt.Errorf("failed CheckMembersExistInConversation: %+v", err)
	}

	if len(existedMemberIDs) > 0 {
		return fmt.Errorf("some members already existed in conversation")
	}

	vendorUserIDs := []string{}
	for _, vendorUser := range chatVendorUsers {
		vendorUserIDs = append(vendorUserIDs, vendorUser.VendorUserID)
	}

	request := &chatvendor_dto.AddConversationMembersRequest{
		ConversationID:  conversationID,
		MemberVendorIDs: vendorUserIDs,
	}

	_, err = svc.ChatVendor.AddConversationMembers(request)
	if err != nil {
		return fmt.Errorf("failed AddConversationMembers: %+v", err)
	}

	// upsert conversation members
	err = database.ExecInTxWithRetry(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := svc.ConversationMemberRepo.BulkUpsert(ctx, tx, conversationMembers)
		if err != nil {
			return fmt.Errorf("failed BulkUpsert: %+v", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
