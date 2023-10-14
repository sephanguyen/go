package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	userconstant "github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/mock/testutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func UserRepoWithSqlMock() (*UserRepo, *testutil.MockDB) {
	userRepo := &UserRepo{}
	return userRepo, testutil.NewMockDB()
}

func TestUserRepo_GetStudentByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userRepoWithSqlMock, mockDB := UserRepoWithSqlMock()

	var userID = "1"
	t.Run("success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			userID,
			cpb.UserGroup_USER_GROUP_STUDENT.String(),
		)
		e := &entities.User{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		user, err := userRepoWithSqlMock.GetStudentByIDForUpdate(ctx, mockDB.DB, userID)
		assert.Nil(t, err)
		assert.NotNil(t, user)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			userID,
			cpb.UserGroup_USER_GROUP_STUDENT.String(),
		)
		e := &entities.User{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		user, err := userRepoWithSqlMock.GetStudentByIDForUpdate(ctx, mockDB.DB, userID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, user)

	})
}

func TestUserRepo_GetStudentsByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userIDs := []string{"1", "2", "3", "4"}
	t.Run("success", func(t *testing.T) {
		userRepoWithSqlMock, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.Anything,
			database.TextArray(userIDs),
			cpb.UserGroup_USER_GROUP_STUDENT.String(),
		)
		e := &entities.User{}
		fields, values := e.FieldMap()

		var dst [][]interface{}
		dst = append(dst, values)
		mockDB.MockScanArray(nil, fields, dst)
		users, err := userRepoWithSqlMock.GetStudentsByIDs(ctx, mockDB.DB, userIDs)
		assert.Nil(t, err)
		assert.NotNil(t, users)

	})
	t.Run("no rows", func(t *testing.T) {
		expectedErr := pgx.ErrNoRows
		userRepoWithSqlMock, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.Anything,
			database.TextArray(userIDs),
			cpb.UserGroup_USER_GROUP_STUDENT.String(),
		)
		e := &entities.User{}
		fields, values := e.FieldMap()

		var dst [][]interface{}
		dst = append(dst, values)
		mockDB.MockScanArray(expectedErr, fields, dst)
		users, err := userRepoWithSqlMock.GetStudentsByIDs(ctx, mockDB.DB, userIDs)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, users)
	})
	t.Run("err query", func(t *testing.T) {
		expectedErr := errors.New("invalid connection")
		r, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, expectedErr,
			mock.Anything,
			mock.Anything,
			database.TextArray(userIDs),
			cpb.UserGroup_USER_GROUP_STUDENT.String(),
		)
		e := &entities.User{}
		fields, values := e.FieldMap()

		var dst [][]interface{}
		dst = append(dst, values)
		mockDB.MockScanArray(nil, fields, dst)
		users, err := r.GetStudentsByIDs(ctx, mockDB.DB, userIDs)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, users)
	})
	t.Run("err scan", func(t *testing.T) {
		expectedErr := errors.New("invalid scan")
		userRepoWithSqlMock, mockDB := UserRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil,
			mock.Anything,
			mock.Anything,
			database.TextArray(userIDs),
			cpb.UserGroup_USER_GROUP_STUDENT.String(),
		)
		e := &entities.User{}
		fields, values := e.FieldMap()

		var dst [][]interface{}
		dst = append(dst, values)
		mockDB.MockScanArray(expectedErr, fields, dst)
		users, err := userRepoWithSqlMock.GetStudentsByIDs(ctx, mockDB.DB, userIDs)
		assert.True(t, errors.Is(err, expectedErr))
		assert.Nil(t, users)
	})
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
