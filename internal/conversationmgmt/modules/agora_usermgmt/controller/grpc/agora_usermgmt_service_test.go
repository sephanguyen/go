package grpc

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/domain/models"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	mock_repositories "github.com/manabie-com/backend/mock/conversationmgmt/modules/agora_usermgmt/infrastructure/repositories"
	mock_chatvendor "github.com/manabie-com/backend/mock/golibs/chatvendor"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_GetAppInfo(t *testing.T) {
	t.Parallel()
	mockDB := testutil.NewMockDB()
	mockUserBasicInfoRepo := &mock_repositories.MockUserBasicInfoRepo{}
	mockChatVendorClient := mock_chatvendor.NewChatVendorClient(t)
	userID := "user-id"
	svc := &AgoraUserMgmtService{
		DB:                mockDB.DB,
		ChatVendorClient:  mockChatVendorClient,
		UserBasicInfoRepo: mockUserBasicInfoRepo,
	}

	userToken := "user-token"
	expiredTime := uint64(100)
	appKey := "app-key"

	testCases := []struct {
		Name     string
		Response *cpb.GetAppInfoResponse
		Err      error
		Setup    func(ctx context.Context) context.Context
	}{
		{
			Name:     "user not exist",
			Response: nil,
			Err:      status.Errorf(codes.InvalidArgument, "user does not exist"),
			Setup: func(ctx context.Context) context.Context {
				ctx = interceptors.ContextWithUserID(ctx, userID)
				userInfos := []*models.UserBasicInfo{}
				mockUserBasicInfoRepo.On("GetUsers", ctx, mock.Anything, []string{userID}).Once().Return(
					userInfos,
					nil,
				)
				return ctx
			},
		},
		{
			Name: "get token success",
			Response: &cpb.GetAppInfoResponse{
				AppKey:           appKey,
				CurrentUserToken: userToken,
				TokenExpiredAt:   expiredTime,
			},
			Err: nil,
			Setup: func(ctx context.Context) context.Context {
				ctx = interceptors.ContextWithUserID(ctx, userID)
				userInfos := []*models.UserBasicInfo{
					{
						UserID: database.Text(userID),
					},
				}
				mockUserBasicInfoRepo.On("GetUsers", ctx, mock.Anything, []string{userID}).Once().Return(
					userInfos,
					nil,
				)

				mockChatVendorClient.On("GetUserToken", userID).Once().Return(
					userToken,
					expiredTime,
					nil,
				)

				mockChatVendorClient.On("GetAppKey").Once().Return(appKey)
				return ctx
			},
		},
	}

	for _, tc := range testCases {
		ctx := context.Background()
		ctx = tc.Setup(ctx)
		resp, err := svc.GetAppInfo(ctx, &cpb.GetAppInfoRequest{})
		assert.Equal(t, tc.Err, err)
		assert.Equal(t, tc.Response, resp)
	}
}
