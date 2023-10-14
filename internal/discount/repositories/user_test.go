package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/discount/utils"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	userconstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserRepoWithSqlMock() (*UserRepo, *testutil.MockDB) {
	userRepo := &UserRepo{}
	return userRepo, testutil.NewMockDB()
}

func TestUserRepo_GetUserIDsByRoleNamesAndLocationID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		userRepoWithSqlMock *UserRepo
		mockDB              *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when query",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				[]string{
					userconstant.RoleCentreManager,
					userconstant.RoleCentreStaff,
				},
				constant.LocationID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Rows, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when scan",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				[]string{
					userconstant.RoleCentreManager,
					userconstant.RoleCentreStaff,
				},
				constant.LocationID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything).Once().Return(constant.ErrDefault)
				mockDB.Rows.On("Close").Once()
			},
		},
		{
			Name: constant.HappyCase,
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				[]string{
					userconstant.RoleCentreManager,
					userconstant.RoleCentreStaff,
				},
				constant.LocationID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", mock.Anything).Once().Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Close").Once()
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			userRepoWithSqlMock, mockDB = UserRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			roleNames := testCase.Req.([]interface{})[0].([]string)
			locationID := testCase.Req.([]interface{})[1].(string)

			_, err := userRepoWithSqlMock.GetUserIDsByRoleNamesAndLocationID(testCase.Ctx, mockDB.DB, roleNames, locationID)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestUserRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepo, mockDB := UserRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.User{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		user, err := userRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, user)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.User{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		user, err := userRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, user)

	})
}
