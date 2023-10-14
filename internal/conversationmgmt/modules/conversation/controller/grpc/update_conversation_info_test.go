package grpc

import (
	"context"
	"fmt"
	"testing"

	mock_service "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/core/port/service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestConversationModifierGRPC_UpdateConversationInfo(t *testing.T) {
	mockPortService := mock_service.NewConversationModifierService(t)

	svc := &ConversationModifierGRPC{
		ConversationModifierServicePort: mockPortService,
	}

	ctx := context.Background()

	t.Run("happy case", func(t *testing.T) {
		mockPortService.On("UpdateConversationInfo", ctx, mock.Anything).Once().Return(nil, nil)

		req := &cpb.UpdateConversationInfoRequest{
			ConversationId: "conversation-id-1",
			Name:           "conversation-name",
			OptionalConfig: []byte("optional-config"),
		}
		_, err := svc.UpdateConversationInfo(ctx, req)
		assert.Nil(t, err)
	})

	t.Run("update conversation info error", func(t *testing.T) {
		mockPortService.On("UpdateConversationInfo", ctx, mock.Anything).Once().Return(nil, puddle.ErrClosedPool)

		req := &cpb.UpdateConversationInfoRequest{
			ConversationId: "conversation-id-1",
			Name:           "conversation-name",
			OptionalConfig: []byte("optional-config"),
		}
		_, err := svc.UpdateConversationInfo(ctx, req)
		assert.Equal(t, status.Error(codes.Internal, fmt.Sprintf("ConversationModifierServicePort.UpdateConversationInfo: %v", puddle.ErrClosedPool)), err)
	})
}
