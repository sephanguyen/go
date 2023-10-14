package service

import (
	"context"
	"errors"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	chatvendor_dto "github.com/manabie-com/backend/internal/golibs/chatvendor/dto"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/infrastructure/postgres"
	mock_chatvendor "github.com/manabie-com/backend/mock/golibs/chatvendor"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func Test_RemoveConversationMembers(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	mockTx := &mock_database.Tx{}
	mockChatVendor := mock_chatvendor.NewChatVendorClient(t)
	mockChatVendorUserRepo := &mock_repositories.MockAgoraUserRepo{}
	mockInternalAdminUserRepo := &mock_repositories.MockInternalAdminUserRepo{}
	mockConversationRepo := &mock_repositories.MockConversationRepo{}
	mockConversationMemberRepo := &mock_repositories.MockConversationMemberRepo{}

	conversationID := "conv-id"
	user1 := "user-1"
	user2 := "user-2"
	vendorUser1 := "vendor-user-1"
	vendorUser2 := "vendor-user-2"
	svc := &conversationModifierServiceImpl{
		Logger:                 zap.NewNop(),
		DB:                     mockDB.DB,
		Environment:            "local",
		ChatVendor:             mockChatVendor,
		ConversationRepo:       mockConversationRepo,
		ConversationMemberRepo: mockConversationMemberRepo,
		ChatVendorUserRepo:     mockChatVendorUserRepo,
		InternalAdminUserRepo:  mockInternalAdminUserRepo,
	}
	testCases := []struct {
		Name                string
		ConversationMembers []domain.ConversationMember
		Err                 error
		Setup               func(ctx context.Context)
	}{
		{
			Name: "conversation not exist",
			ConversationMembers: []domain.ConversationMember{
				{
					ConversationID: conversationID,
				},
			},
			Err: errors.New("conversation not found"),
			Setup: func(ctx context.Context) {
				mockConversationRepo.On("FindByIDs", ctx, mockDB.DB, []string{conversationID}).Once().Return([]*domain.Conversation{}, nil)
			},
		},
		{
			Name: "members not exist",
			ConversationMembers: []domain.ConversationMember{
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user1,
					},
				},
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user2,
					},
				},
			},
			Err: errors.New("some users do not exist"),
			Setup: func(ctx context.Context) {
				mockConversationRepo.On("FindByIDs", ctx, mockDB.DB, []string{conversationID}).Once().
					Return([]*domain.Conversation{{ID: conversationID}}, nil)
				mockChatVendorUserRepo.On("GetByUserIDs", ctx, mockDB.DB, mock.MatchedBy(func(in []string) bool {
					if !stringutil.SliceEqual([]string{user1, user2}, in) {
						return false
					}
					return true
				})).Once().
					Return([]*domain.ChatVendorUser{}, nil)
			},
		},
		{
			Name: "members not belong to conversation",
			ConversationMembers: []domain.ConversationMember{
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user1,
					},
				},
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user2,
					},
				},
			},
			Err: errors.New("some users do not belong to conversation"),
			Setup: func(ctx context.Context) {
				mockConversationRepo.On("FindByIDs", ctx, mockDB.DB, []string{conversationID}).Once().
					Return([]*domain.Conversation{{ID: conversationID}}, nil)
				mockChatVendorUserRepo.On("GetByUserIDs", ctx, mockDB.DB, mock.MatchedBy(func(in []string) bool {
					if !stringutil.SliceEqual([]string{user1, user2}, in) {
						return false
					}
					return true
				})).Once().
					Return([]*domain.ChatVendorUser{
						{
							UserID: user1,
						},
						{
							UserID: user2,
						},
					}, nil)
				mockConversationMemberRepo.On("CheckMembersExistInConversation", ctx, mockDB.DB, conversationID, []string{user1, user2}).Once().
					Return([]string{user1}, nil)
			},
		},
		{
			Name: "should no error",
			ConversationMembers: []domain.ConversationMember{
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user1,
					},
				},
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user2,
					},
				},
			},
			Err: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockConversationRepo.On("FindByIDs", ctx, mockDB.DB, []string{conversationID}).Once().
					Return([]*domain.Conversation{{ID: conversationID}}, nil)
				mockChatVendorUserRepo.On("GetByUserIDs", ctx, mockDB.DB, mock.MatchedBy(func(in []string) bool {
					if !stringutil.SliceEqual([]string{user1, user2}, in) {
						return false
					}
					return true
				})).Once().
					Return([]*domain.ChatVendorUser{
						{
							UserID:       user1,
							VendorUserID: vendorUser1,
						},
						{
							UserID:       user2,
							VendorUserID: vendorUser2,
						},
					}, nil)
				mockConversationMemberRepo.On("CheckMembersExistInConversation", ctx, mockDB.DB, conversationID, []string{user1, user2}).Once().
					Return([]string{user1, user2}, nil)
				mockChatVendor.On("RemoveConversationMembers", mock.MatchedBy(func(in *chatvendor_dto.RemoveConversationMembersRequest) bool {
					if in.ConversationID != conversationID || !stringutil.SliceEqual(in.MemberVendorIDs, []string{vendorUser1, vendorUser2}) {
						return false
					}
					return true
				})).Once().Return(&chatvendor_dto.RemoveConversationMembersResponse{}, nil)

				mockConversationMemberRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
		{
			Name: "success with some failed remove members",
			ConversationMembers: []domain.ConversationMember{
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user1,
					},
				},
				{
					ConversationID: conversationID,
					User: domain.ChatVendorUser{
						UserID: user2,
					},
				},
			},
			Err: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(mockTx, nil)
				mockTx.On("Commit", mock.Anything).Return(nil)
				mockConversationRepo.On("FindByIDs", ctx, mockDB.DB, []string{conversationID}).Once().
					Return([]*domain.Conversation{{ID: conversationID}}, nil)
				mockChatVendorUserRepo.On("GetByUserIDs", ctx, mockDB.DB, mock.MatchedBy(func(in []string) bool {
					if !stringutil.SliceEqual([]string{user1, user2}, in) {
						return false
					}
					return true
				})).Once().
					Return([]*domain.ChatVendorUser{
						{
							UserID:       user1,
							VendorUserID: vendorUser1,
						},
						{
							UserID:       user2,
							VendorUserID: vendorUser2,
						},
					}, nil)
				mockConversationMemberRepo.On("CheckMembersExistInConversation", ctx, mockDB.DB, conversationID, []string{user1, user2}).Once().
					Return([]string{user1, user2}, nil)
				mockChatVendor.On("RemoveConversationMembers", mock.MatchedBy(func(in *chatvendor_dto.RemoveConversationMembersRequest) bool {
					if in.ConversationID != conversationID || !stringutil.SliceEqual(in.MemberVendorIDs, []string{vendorUser1, vendorUser2}) {
						return false
					}
					return true
				})).Once().Return(&chatvendor_dto.RemoveConversationMembersResponse{
					ConversationID: conversationID,
					FailedMembers: []chatvendor_dto.FailedRemoveMember{
						{
							MemberVendorID: vendorUser2,
							Reason:         "some reason",
						},
					},
				}, nil)

				mockConversationMemberRepo.On("BulkUpsert", ctx, mockTx, mock.Anything).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx)
			err := svc.RemoveConversationMembers(ctx, tc.ConversationMembers)
			assert.Equal(t, tc.Err, err)
		})
	}
}
