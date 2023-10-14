package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	mock_service "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/core/port/service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConversationReaderGRPC_GetConversationsDetail(t *testing.T) {
	t.Parallel()

	mockPortService := mock_service.NewConversationReaderService(t)

	svc := &ConversationReaderGRPC{
		ConversationReaderServicePort: mockPortService,
	}

	ctx := context.Background()
	conversationIDs := []string{"conversation-id-1", "conversation-id-2"}

	t.Run("happy case", func(t *testing.T) {
		conversationMemberDomains := []domain.ConversationMember{
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
				Members:               conversationMemberDomains,
			},
			{
				ID:                    "conversation-id-2",
				Name:                  "conversation-name-2",
				LatestMessage:         nil,
				LatestMessageSentTime: nil,
				OptionalConfig:        []byte("optional-config"),
				CreatedAt:             time.Now(),
				UpdatedAt:             time.Now(),
				Members:               conversationMemberDomains,
			},
		}
		conversationResponses, err := svc.mapConversationDomainsToPbGetConversationsDetailResponse(conversationDomains)
		assert.Nil(t, err)
		mockPortService.On("GetConversationsDetail", ctx, conversationIDs).Once().Return(conversationDomains, nil)
		res, err := svc.GetConversationsDetail(ctx, &cpb.GetConversationsDetailRequest{ConversationIds: conversationIDs})

		assert.Nil(t, err)
		assert.Equal(t, conversationResponses, res)
	})

	t.Run("get conversation error", func(t *testing.T) {
		mockPortService.On("GetConversationsDetail", ctx, conversationIDs).Once().Return(nil, puddle.ErrClosedPool)
		_, err := svc.GetConversationsDetail(ctx, &cpb.GetConversationsDetailRequest{ConversationIds: conversationIDs})
		assert.Equal(t, status.Error(codes.Internal, fmt.Sprintf("ConversationReaderServicePort.GetConversationsDetail: %v", puddle.ErrClosedPool)), err)
	})
}
