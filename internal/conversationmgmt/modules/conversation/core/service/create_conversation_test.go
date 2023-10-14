package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	chatvendor_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/infrastructure/postgres"
	mock_chatvendor "github.com/manabie-com/backend/mock/golibs/chatvendor"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationModifierService_CreateConversation(t *testing.T) {
	t.Parallel()
	mockDB := &mock_database.Ext{}
	mockTx := &mock_database.Tx{}

	mockChatVendor := mock_chatvendor.NewChatVendorClient(t)
	mockChatVendorUserRepo := &mock_repositories.MockAgoraUserRepo{}
	mockInternalAdminUserRepo := &mock_repositories.MockInternalAdminUserRepo{}
	mockConversationRepo := &mock_repositories.MockConversationRepo{}
	mockConversationMemberRepo := &mock_repositories.MockConversationMemberRepo{}

	svc := &conversationModifierServiceImpl{
		DB:                     mockDB,
		Environment:            "local",
		ChatVendor:             mockChatVendor,
		ConversationRepo:       mockConversationRepo,
		ConversationMemberRepo: mockConversationMemberRepo,
		ChatVendorUserRepo:     mockChatVendorUserRepo,
		InternalAdminUserRepo:  mockInternalAdminUserRepo,
	}

	conversationID := "convo-id"
	memberIDs := []string{"user-member-1"}
	vendorMemberIDs := []string{"user-vendor-member-1"}

	testCases := []struct {
		Name  string
		Err   error
		Req   *domain.Conversation
		Setup func(ctx context.Context, req *domain.Conversation)
	}{
		{
			Name: "happy case local",
			Err:  nil,
			Req: &domain.Conversation{
				ID:   "",
				Name: "unit-test-conversation",
				Members: []domain.ConversationMember{
					{
						ID:             "",
						ConversationID: "",
						User: domain.ChatVendorUser{
							UserID:       memberIDs[0],
							VendorUserID: vendorMemberIDs[0],
						},
					},
				},
				OptionalConfig: []byte(`{"unit_test_field": "unit_test_value"}`),
			},
			Setup: func(ctx context.Context, req *domain.Conversation) {
				mockDB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)

				mockChatVendorUserRepo.On("GetByUserIDs", mock.Anything, mockDB, memberIDs).Once().Return([]*domain.ChatVendorUser{
					{
						UserID:       memberIDs[0],
						VendorUserID: vendorMemberIDs[0],
					},
				}, nil)

				chatVendorCreateConvoReq := &chatvendor_dto.CreateConversationRequest{
					OwnerVendorID:   common.OwnerChatGroupUserOnLocal,
					MemberVendorIDs: vendorMemberIDs,
				}
				mockChatVendor.On("CreateConversation", chatVendorCreateConvoReq).Once().Return(&chatvendor_dto.CreateConversationResponse{
					ConversationID: conversationID,
				}, nil)

				// Fill conversationID
				req.ID = conversationID
				for i := 0; i < len(req.Members); i++ {
					req.Members[i].ConversationID = conversationID
				}

				mockConversationRepo.On("UpsertConversation", mock.Anything, mockTx, req).Once().Return(nil)
				mockConversationMemberRepo.On("BulkUpsert", mock.Anything, mockTx, req.Members).Once().Return(nil)
			},
		},
	}

	userID := memberIDs[0]
	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx = interceptors.ContextWithUserID(ctx, userID)
			testCase.Setup(ctx, testCase.Req)
			_, err := svc.CreateConversation(ctx, testCase.Req)
			assert.Nil(t, testCase.Err)
			if testCase.Err != nil {
				assert.Equal(t, testCase.Err, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
