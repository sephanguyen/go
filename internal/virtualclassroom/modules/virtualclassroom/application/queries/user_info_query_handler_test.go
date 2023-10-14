package queries

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/support"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	mock_unleash_client "github.com/manabie-com/backend/mock/golibs/unleashclient"
	mock_virtual_repo "github.com/manabie-com/backend/mock/virtualclassroom/virtualclassroom/repositories"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUserInfoQuery_GetUserInfosByIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := &mock_database.Ext{}
	mockUnleashClient := new(mock_unleash_client.UnleashClientInstance)
	wrapperConnection := support.InitWrapperDBConnector(db, db, mockUnleashClient, "local")
	userBasicInfoRepo := &mock_virtual_repo.MockUserBasicInfoRepo{}

	userIDs := []string{"user-id1", "user-id2", "user-id3"}
	userInfos := []domain.UserBasicInfo{
		{
			UserID: "user-id1",
			Name:   "user name 1",
		},
		{
			UserID: "user-id2",
			Name:   "user name 2",
		},
		{
			UserID: "user-id3",
			Name:   "user name 3",
		},
	}

	testCases := []struct {
		name     string
		setup    func(ctx context.Context)
		result   []domain.UserBasicInfo
		hasError bool
	}{
		{
			name: "success",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userBasicInfoRepo.On("GetUserInfosByIDs", ctx, db, userIDs).
					Return(userInfos, nil).Once()
			},
			result: userInfos,
		},
		{
			name: "failed",
			setup: func(ctx context.Context) {
				mockUnleashClient.
					On("IsFeatureEnabledOnOrganization", mock.Anything, mock.Anything, mock.Anything).
					Return(false, nil).Once()
				userBasicInfoRepo.On("GetUserInfosByIDs", ctx, db, userIDs).
					Return(nil, errors.New("error")).Once()
			},
			hasError: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setup(ctx)

			query := UserInfoQuery{
				WrapperDBConnection: wrapperConnection,
				UserBasicInfoRepo:   userBasicInfoRepo,
			}
			res, err := query.GetUserInfosByIDs(ctx, userIDs)

			if tc.hasError {
				require.Error(t, err)
				require.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, res, tc.result)
			}
			mock.AssertExpectationsForObjects(t, mockUnleashClient)
		})
	}

}
