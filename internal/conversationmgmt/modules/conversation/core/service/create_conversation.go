package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	chatvendor_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4"
)

// Internal gRPC
func (svc *conversationModifierServiceImpl) CreateConversation(ctx context.Context, conversation *domain.Conversation) (*domain.Conversation, error) {
	createConversationReq, err := svc.makeChatVendorCreateConversationReq(ctx, conversation)
	if err != nil {
		return nil, fmt.Errorf("error when makeChatVendorCreateConversationReq: [%v]", err)
	}

	chatVendorCreateConversationRes, err := svc.ChatVendor.CreateConversation(createConversationReq)
	if err != nil {
		return nil, fmt.Errorf("error when ChatVendor.CreateConversation: [%v]", err)
	}

	if chatVendorCreateConversationRes.ConversationID == "" {
		return nil, fmt.Errorf("canot create conversation by chatvendor")
	}

	// Fill conversation_id
	conversationID := chatVendorCreateConversationRes.ConversationID
	conversation.ID = conversationID
	for i := 0; i < len(conversation.Members); i++ {
		conversation.Members[i].ConversationID = conversationID
	}

	err = database.ExecInTxWithRetry(ctx, svc.DB, func(ctx context.Context, tx pgx.Tx) error {
		err := svc.ConversationRepo.UpsertConversation(ctx, tx, conversation)
		if err != nil {
			return fmt.Errorf("error when ConversationRepo.UpsertConversation: [%v]", err)
		}

		err = svc.ConversationMemberRepo.BulkUpsert(ctx, tx, conversation.Members)
		if err != nil {
			return fmt.Errorf("error when ConversationMemberRepo.BulkUpsert: [%v]", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("err when database.ExecInTxWithRetry: [%v]", err)
	}

	return conversation, nil
}

func (svc *conversationModifierServiceImpl) makeChatVendorCreateConversationReq(ctx context.Context, conversation *domain.Conversation) (*chatvendor_dto.CreateConversationRequest, error) {
	memberIDs := make([]string, 0)
	for _, member := range conversation.Members {
		memberIDs = append(memberIDs, member.User.UserID)
	}

	chatVendorUsers, err := svc.ChatVendorUserRepo.GetByUserIDs(ctx, svc.DB, memberIDs)
	if err != nil {
		return nil, fmt.Errorf("error when ChatVendorUserRepo.GetByUserIDs: [%v]", err)
	}
	if len(chatVendorUsers) != len(memberIDs) {
		return nil, fmt.Errorf("error when ChatVendorUserRepo.GetByUserIDs: [some user not exist (on agora or manabie)]")
	}

	// Support testing on Local
	ownerID := ""
	if svc.Environment != common.LocalEnv {
		internalAdminUser, err := svc.InternalAdminUserRepo.GetOne(ctx, svc.DB)
		if err != nil || internalAdminUser.VendorUserID == "" {
			return nil, fmt.Errorf("error when InternalAdminUserRepo.GetOne: [%v]", err)
		}

		ownerID = internalAdminUser.VendorUserID
	} else {
		ownerID = common.OwnerChatGroupUserOnLocal
	}

	memberVendorIDs := make([]string, 0)
	for _, memberVendor := range chatVendorUsers {
		memberVendorIDs = append(memberVendorIDs, memberVendor.VendorUserID)
	}

	return &chatvendor_dto.CreateConversationRequest{
		OwnerVendorID:   ownerID,
		MemberVendorIDs: memberVendorIDs,
	}, nil
}
