package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/puddle"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/infrastructure/postgres"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
)

func TestConversationReaderService_GetConversationsDetail(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}

	mockConversationRepo := &mock_repositories.MockConversationRepo{}
	mockConversationMemberRepo := &mock_repositories.MockConversationMemberRepo{}
	mockChatVendorUserRepo := &mock_repositories.MockAgoraUserRepo{}

	svc := &conversationReaderServiceImpl{
		DB:                     mockDB,
		Environment:            "local",
		ConversationRepo:       mockConversationRepo,
		ConversationMemberRepo: mockConversationMemberRepo,
		ChatVendorUserRepo:     mockChatVendorUserRepo,
	}

	conversationIDs := []string{"conversation-id-1", "conversation-id-2"}
	userID := "user-id-1"

	ctx := context.Background()
	ctx = interceptors.ContextWithUserID(ctx, userID)

	t.Run("happy case", func(t *testing.T) {
		conversationMemberDomains := []*domain.ConversationMember{
			{
				ID:             "conversation-member-1",
				ConversationID: "conversation-id-1",
				User: domain.ChatVendorUser{
					UserID:       "user-id-1",
					VendorUserID: "vendor-user-id-1",
				},
				Status:    common.ConversationMemberStatusActive,
				SeenAt:    time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:             "conversation-member-2",
				ConversationID: "conversation-id-1",
				User: domain.ChatVendorUser{
					UserID:       "user-id-2",
					VendorUserID: "vendor-user-id-2",
				},
				Status:    common.ConversationMemberStatusActive,
				SeenAt:    time.Now(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		conversationDomains := []*domain.Conversation{
			{
				ID:                    "conversation-id-1",
				Name:                  "conversation-name-1",
				LatestMessage:         nil,
				LatestMessageSentTime: nil,
				OptionalConfig:        []byte("optional-config"),
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			},
			{
				ID:                    "conversation-id-2",
				Name:                  "conversation-name-2",
				LatestMessage:         nil,
				LatestMessageSentTime: nil,
				OptionalConfig:        []byte("optional-config"),
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
			},
		}
		chatVendorUsers := []*domain.ChatVendorUser{
			{
				UserID:       "user-id-1",
				VendorUserID: "vendor-user-id-1",
			},
			{
				UserID:       "user-id-2",
				VendorUserID: "vendor-user-id-2",
			},
		}
		conversationWithMembers := mapMembersToConversations(conversationDomains, conversationMemberDomains, chatVendorUsers)
		mockConversationRepo.On("FindByIDsAndUserID", ctx, mockDB, userID, conversationIDs).Once().Return(conversationDomains, nil)
		mockConversationMemberRepo.On("GetConversationMembersByUserID", ctx, mockDB, userID, conversationIDs).Once().Return(conversationMemberDomains, nil)
		mockChatVendorUserRepo.On("GetByUserIDs", ctx, mockDB, []string{"user-id-1", "user-id-2"}).Once().Return(chatVendorUsers, nil)

		res, err := svc.GetConversationsDetail(ctx, conversationIDs)
		assert.Nil(t, err)
		assert.Equal(t, conversationWithMembers, res)
	})

	t.Run("get conversations error", func(t *testing.T) {
		mockConversationRepo.On("FindByIDsAndUserID", ctx, mockDB, userID, conversationIDs).Once().Return(nil, puddle.ErrClosedPool)

		_, err := svc.GetConversationsDetail(ctx, conversationIDs)
		assert.Equal(t, fmt.Errorf("error when ConversationRepo.GetConversations: [%v]", puddle.ErrClosedPool), err)
	})
}
