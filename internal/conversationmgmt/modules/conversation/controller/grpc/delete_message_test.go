package grpc

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_service "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/core/port/service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationModifierGRPC_DeleteMessage(t *testing.T) {
	t.Parallel()

	mockPortService := mock_service.NewConversationModifierService(t)

	svc := &ConversationModifierGRPC{
		ConversationModifierServicePort: mockPortService,
	}

	conversationID := "convo-id"
	messageID := "msg-id"
	userID := "user-id"

	testCases := []struct {
		Name  string
		Err   error
		Req   *cpb.DeleteMessageRequest
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Req: &cpb.DeleteMessageRequest{
				ConversationId:  conversationID,
				VendorMessageId: messageID,
			},
			Setup: func(ctx context.Context) {
				msg := &domain.Message{
					ConversationID:  conversationID,
					VendorMessageID: messageID,
					IsDeleted:       true,
				}

				mockPortService.On("DeleteMessage", mock.Anything, msg).Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx = interceptors.ContextWithUserID(ctx, userID)
			testCase.Setup(ctx)
			_, err := svc.DeleteMessage(ctx, testCase.Req)

			assert.Nil(t, err)
		})
	}
}
