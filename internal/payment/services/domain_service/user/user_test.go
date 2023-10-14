package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserService_GetUserIDsForLoaNotification(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		mockUserRepo *mockRepositories.MockUserRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get user ids by role names and location id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: []interface{}{
				constant.LocationID,
			},
			Setup: func(ctx context.Context) {
				mockUserRepo.On("GetUserIDsByRoleNamesAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				constant.LocationID,
			},
			Setup: func(ctx context.Context) {
				mockUserRepo.On("GetUserIDsByRoleNamesAndLocationID", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]string{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			mockUserRepo = &mockRepositories.MockUserRepo{}

			s := &UserService{
				userRepo: mockUserRepo,
			}
			testCase.Setup(testCase.Ctx)

			location := testCase.Req.([]interface{})[0].(string)
			_, err := s.GetUserIDsForLoaNotification(testCase.Ctx, db, location)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, mockUserRepo, db)
		})
	}
}
