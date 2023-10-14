package grpc

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	mock_service "github.com/manabie-com/backend/mock/conversationmgmt/modules/conversation/core/port/service"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_RemoveConversationMembers(t *testing.T) {
	t.Parallel()

	mockPortService := mock_service.NewConversationModifierService(t)

	svc := &ConversationModifierGRPC{
		ConversationModifierServicePort: mockPortService,
	}

	testCases := []struct {
		Name     string
		Setup    func(ctx context.Context)
		Request  *cpb.RemoveConversationMembersRequest
		Response *cpb.RemoveConversationMembersResponse
		Err      error
	}{
		{
			Name: "too many members in request",
			Request: func() *cpb.RemoveConversationMembersRequest {
				req := &cpb.RemoveConversationMembersRequest{
					ConversationId: "conv-1",
				}
				for i := 0; i < common.MaxMembersInRequest+1; i++ {
					req.MemberIds = append(req.MemberIds, idutil.ULIDNow())
				}
				return req
			}(),
			Response: nil,
			Err:      status.Errorf(codes.InvalidArgument, "Too many members in request"),
			Setup:    func(ctx context.Context) {},
		},
		{
			Name: "should success",
			Request: &cpb.RemoveConversationMembersRequest{
				ConversationId: "conv-1",
				MemberIds:      []string{"member-1", "member-2"},
			},
			Response: &cpb.RemoveConversationMembersResponse{},
			Err:      nil,
			Setup: func(ctx context.Context) {
				expectConversationMembers := []domain.ConversationMember{
					{
						ConversationID: "conv-1",
						User: domain.ChatVendorUser{
							UserID: "member-1",
						},
						Status: common.ConversationMemberStatusInActive,
					},
					{
						ConversationID: "conv-1",
						User: domain.ChatVendorUser{
							UserID: "member-2",
						},
						Status: common.ConversationMemberStatusInActive,
					},
				}
				mockPortService.On("RemoveConversationMembers", ctx, mock.MatchedBy(func(in []domain.ConversationMember) bool {
					if len(expectConversationMembers) != len(in) {
						return false
					}

					for i, member := range expectConversationMembers {
						if member.ConversationID != in[i].ConversationID {
							return false
						}
						if member.User.UserID != in[i].User.UserID {
							return false
						}
						if member.Status != in[i].Status {
							return false
						}
					}
					return true
				})).Once().Return(nil)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			tc.Setup(ctx)
			res, err := svc.RemoveConversationMembers(ctx, tc.Request)
			assert.Equal(t, tc.Err, err)
			assert.Equal(t, tc.Response, res)
		})
	}
}
