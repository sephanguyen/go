package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/infrastructure/postgres"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestConversationModifierService_DeleteMessage(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	mockConversationRepo := &mock_repositories.MockConversationRepo{}
	mockChatVendorUserRepo := &mock_repositories.MockAgoraUserRepo{}

	svc := &conversationModifierServiceImpl{
		DB:                 mockDB.DB,
		Environment:        "local",
		Logger:             zap.NewExample(),
		ConversationRepo:   mockConversationRepo,
		ChatVendorUserRepo: mockChatVendorUserRepo,
	}

	conversatiionID := "convo-id"
	messageID := "message-id"
	vendorUserID := "vendor-user-id"
	userID := "user-id"
	message := &domain.Message{
		ConversationID:  conversatiionID,
		VendorMessageID: messageID,
		UserID:          userID,
		IsDeleted:       true,
	}

	testCases := []struct {
		Name  string
		Err   error
		Req   *domain.Message
		Setup func(ctx context.Context, req *domain.Message)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Req:  message,
			Setup: func(ctx context.Context, req *domain.Message) {
				mockConversationRepo.On("FindByIDsAndUserID", mock.Anything, mock.Anything, userID, []string{conversatiionID}).Once().Return([]*domain.Conversation{
					{
						LatestMessage: req,
					},
				})
				mockChatVendorUserRepo.On("GetByUserID", mock.Anything, mock.Anything, userID).Once().Return(&domain.ChatVendorUser{UserID: vendorUserID})
				mockConversationRepo.On("UpdateLatestMessage", mock.Anything, mock.Anything, req).Once().Return(nil)
			},
		},
	}

	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(ctx, testCase.Req)
			ctx = interceptors.ContextWithUserID(ctx, userID)
			err := svc.UpdateLatestMessage(ctx, testCase.Req)
			assert.Nil(t, testCase.Err)
			if testCase.Err != nil {
				assert.Equal(t, testCase.Err, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
