package grpc

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_service "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/core/port/service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConversationModifierGRPC_CreateConversation(t *testing.T) {
	t.Parallel()

	mockPortService := mock_service.NewConversationModifierService(t)

	svc := &ConversationModifierGRPC{
		ConversationModifierServicePort: mockPortService,
	}

	conversationID := "convo-id"
	conversationName := "unit-test-conversation"
	conversationMemberID := "convo-mem-id"
	memberIDs := []string{"user-member-1"}

	testCases := []struct {
		Name  string
		Err   error
		Req   *cpb.CreateConversationRequest
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Req: &cpb.CreateConversationRequest{
				Name:           conversationName,
				MemberIds:      memberIDs,
				OptionalConfig: []byte(`{"unit_test_field": "unit_test_value"}`),
			},
			Setup: func(ctx context.Context) {
				convoDomain := &domain.Conversation{
					ID:   "",
					Name: conversationName,
					Members: []domain.ConversationMember{
						{
							ID:             "",
							ConversationID: "",
							Status:         common.ConversationMemberStatusActive,
							User: domain.ChatVendorUser{
								UserID: memberIDs[0],
							},
						},
					},
					OptionalConfig: []byte(`{"unit_test_field": "unit_test_value"}`),
				}

				mockPortService.On("CreateConversation", mock.Anything, convoDomain).Once().Return(&domain.Conversation{
					ID:   conversationID,
					Name: "unit-test-conversation",
					Members: []domain.ConversationMember{
						{
							ID:             conversationMemberID,
							ConversationID: conversationID,
							Status:         common.ConversationMemberStatusActive,
							User: domain.ChatVendorUser{
								UserID: memberIDs[0],
							},
						},
					},
					OptionalConfig: []byte(`{"unit_test_field": "unit_test_value"}`),
				}, nil)
			},
		},
	}

	userID := memberIDs[0]
	ctx := context.Background()
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx = interceptors.ContextWithUserID(ctx, userID)
			testCase.Setup(ctx)
			res, err := svc.CreateConversation(ctx, testCase.Req)
			if testCase.Err != nil {
				assert.Equal(t, testCase.Err, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, conversationID, res.ConversationId)
				assert.Equal(t, conversationName, res.Name)
			}
		})
	}
}
