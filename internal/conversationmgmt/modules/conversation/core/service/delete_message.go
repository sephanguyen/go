package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

// Internal gRPC
func (svc *conversationModifierServiceImpl) DeleteMessage(ctx context.Context, message *domain.Message) error {
	userID := interceptors.UserIDFromContext(ctx)

	conversations, err := svc.ConversationRepo.FindByIDsAndUserID(ctx, svc.DB, userID, []string{message.ConversationID})
	if err != nil {
		return fmt.Errorf("error ConversationRepo.FindByID: [%+v]", err)
	}
	if len(conversations) == 0 {
		return fmt.Errorf("not found conversation with current user or user is not a member of this conversation")
	}

	// Get chatvendor user
	vendorUser, err := svc.ChatVendorUserRepo.GetByUserID(ctx, svc.DB, userID)
	if err != nil {
		return fmt.Errorf("error ChatVendorUserRepo.GetByUserID: [%+v]", err)
	}

	_, err = svc.ChatVendor.DeleteMessage(&dto.DeleteMessageRequest{
		ConversationID:  message.ConversationID,
		VendorMessageID: message.VendorMessageID,
		VendorUserID:    vendorUser.VendorUserID,
	})
	if err != nil {
		return fmt.Errorf("error ChatVendor.DeleteMessage: %+v", err)
	}

	conversation := conversations[0]
	// Not found latest message
	if conversation.LatestMessage == nil {
		return nil
	}

	latestMessage := conversation.LatestMessage
	if latestMessage.VendorMessageID == message.VendorMessageID {
		latestMessage.IsDeleted = true
		err := svc.ConversationRepo.UpdateLatestMessage(ctx, svc.DB, latestMessage)
		if err != nil {
			return fmt.Errorf("ConversationRepo.UpdateLatestMessage: [%+v]", err)
		}
	}

	return nil
}
