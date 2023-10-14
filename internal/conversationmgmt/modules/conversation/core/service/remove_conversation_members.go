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
func (svc *conversationModifierServiceImpl) RemoveConversationMembers(ctx context.Context, conversationMembers []domain.ConversationMember) error {
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

	existedMemberIDs, err := svc.ConversationMemberRepo.CheckMembersExistInConversation(ctx, svc.DB, conversationID, memberIDs)
	if err != nil {
		return fmt.Errorf("failed CheckMembersExistInConversation: %+v", err)
	}

	if len(existedMemberIDs) != len(memberIDs) {
		return fmt.Errorf("some users do not belong to conversation")
	}

	vendorUserIDs := []string{}
	for _, vendorUser := range chatVendorUsers {
		vendorUserIDs = append(vendorUserIDs, vendorUser.VendorUserID)
	}

	request := &chatvendor_dto.RemoveConversationMembersRequest{
		ConversationID:  conversationID,
		MemberVendorIDs: vendorUserIDs,
	}

	chatVendorResponse, err := svc.ChatVendor.RemoveConversationMembers(request)
	if err != nil {
		return fmt.Errorf("failed RemoveConversationMembers: %+v", err)
	}

	if len(chatVendorResponse.FailedMembers) > 0 {
		// TODO: log warn here
		svc.Logger.Warn(fmt.Sprintf("failed to remove %d user from conversation %s", len(chatVendorResponse.FailedMembers), conversationID))
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
