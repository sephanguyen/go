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

func Test_AddConversationMembers(t *testing.T) {
	t.Parallel()

	mockPortService := mock_service.NewConversationModifierService(t)

	svc := &ConversationModifierGRPC{
		ConversationModifierServicePort: mockPortService,
	}

	testCases := []struct {
		Name     string
		Setup    func(ctx context.Context)
		Request  *cpb.AddConversationMembersRequest
		Response *cpb.AddConversationMembersResponse
		Err      error
	}{
		{
			Name: "too many members in request",
			Request: func() *cpb.AddConversationMembersRequest {
				req := &cpb.AddConversationMembersRequest{
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
			Request: &cpb.AddConversationMembersRequest{
				ConversationId: "conv-1",
				MemberIds:      []string{"member-1", "member-2"},
			},
			Response: &cpb.AddConversationMembersResponse{},
			Err:      nil,
			Setup: func(ctx context.Context) {
				expectConversationMembers := []domain.ConversationMember{
					{
						ConversationID: "conv-1",
						User: domain.ChatVendorUser{
							UserID: "member-1",
						},
						Status: common.ConversationMemberStatusActive,
					},
					{
						ConversationID: "conv-1",
						User: domain.ChatVendorUser{
							UserID: "member-2",
						},
						Status: common.ConversationMemberStatusActive,
					},
				}
				mockPortService.On("AddConversationMembers", ctx, mock.MatchedBy(func(in []domain.ConversationMember) bool {
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
			res, err := svc.AddConversationMembers(ctx, tc.Request)
			assert.Equal(t, tc.Err, err)
			assert.Equal(t, tc.Response, res)
		})
	}
}

func Test_toConversationMemberDomain(t *testing.T) {
	t.Parallel()
	memberIDs := []string{"member-1", "member-2"}
	convID := "conv-id"
	testCases := []struct {
		Name                            string
		MemberIDs                       []string
		Opts                            []domain.ConversationMemberOpt
		ExpectDomainConversationMembers []domain.ConversationMember
	}{
		{
			Name:      "should no error with status active",
			MemberIDs: memberIDs,
			Opts: []domain.ConversationMemberOpt{
				domain.WithConversationID(convID),
				domain.WithStatus(common.ConversationMemberStatusActive),
			},
			ExpectDomainConversationMembers: []domain.ConversationMember{
				{
					ConversationID: convID,
					User: domain.ChatVendorUser{
						UserID: memberIDs[0],
					},
					Status: common.ConversationMemberStatusActive,
				},
				{
					ConversationID: convID,
					User: domain.ChatVendorUser{
						UserID: memberIDs[1],
					},
					Status: common.ConversationMemberStatusActive,
				},
			},
		},
		{
			Name:      "should no error with status inactive",
			MemberIDs: memberIDs,
			Opts: []domain.ConversationMemberOpt{
				domain.WithConversationID(convID),
				domain.WithStatus(common.ConversationMemberStatusInActive),
			},
			ExpectDomainConversationMembers: []domain.ConversationMember{
				{
					ConversationID: convID,
					User: domain.ChatVendorUser{
						UserID: memberIDs[0],
					},
					Status: common.ConversationMemberStatusInActive,
				},
				{
					ConversationID: convID,
					User: domain.ChatVendorUser{
						UserID: memberIDs[1],
					},
					Status: common.ConversationMemberStatusInActive,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			dConversationMembers := toConversationMemberDomain(tc.MemberIDs, tc.Opts...)

			assert.Equal(t, tc.ExpectDomainConversationMembers, dConversationMembers)
		})
	}
}
