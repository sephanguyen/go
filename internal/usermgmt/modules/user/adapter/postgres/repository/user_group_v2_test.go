package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func userGroupV2RepoMock() (*UserGroupV2Repo, *testutil.MockDB) {
	repo := &UserGroupV2Repo{}
	return repo, testutil.NewMockDB()
}

func TestUserGroupV2_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	UserGroupV2Repo := &UserGroupV2Repo{}
	UserGroupV2 := &entity.UserGroupV2{}
	_, UserGroupV2Values := UserGroupV2.FieldMap()
	argsUserGroup := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(UserGroupV2Values))...,
	)

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				cmtTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmtTag, nil)
			},
		},
		{
			name:      "can not insert usergroup",
			expectErr: fmt.Errorf("err insert usergroup: %w", fmt.Errorf("error")),
			setup: func() {
				cmtTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmtTag, fmt.Errorf("error"))
			},
		},
		{
			name:      "cannot upsert usergroup",
			expectErr: fmt.Errorf("cannot upsert usergroup"),
			setup: func() {
				cmtTag := pgconn.CommandTag([]byte(`0`))
				mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmtTag, nil)
			},
		},
	}

	for index, testcase := range testCases {
		testName := fmt.Sprintf("%s-%d", testcase.name, index)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			err := UserGroupV2Repo.Create(ctx, mockDB.DB, UserGroupV2)
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			}
		})

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	}
}

func TestUserGroupV2Repo_Update(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userGroupID := pgtype.Text{}
	userGroupID.Set(idutil.ULIDNow())
	userGroup := &entity.UserGroupV2{
		UserGroupID: userGroupID,
	}
	_, userGroupValues := (&entity.UserGroupV2{}).FieldMap()
	argsUserGroup := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(userGroupValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := userGroupV2RepoMock()
		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, userGroup)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update user group fail", func(t *testing.T) {
		repo, mockDB := userGroupV2RepoMock()
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(nil, puddle.ErrClosedPool)

		err := repo.Update(ctx, mockDB.DB, userGroup)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update user fail: rows affect not equal", func(t *testing.T) {
		repo, mockDB := userGroupV2RepoMock()
		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, userGroup)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUserGroupV2Repo_Find(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	userGroupID := pgtype.Text{}
	userGroupID.Set(idutil.ULIDNow())
	_, userGroupValues := (&entity.UserGroupV2{}).FieldMap()
	argsUserGroup := append([]interface{}{}, genSliceMock(len(userGroupValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := userGroupV2RepoMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &userGroupID).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsUserGroup...).Once().Return(nil)
		userGroup, err := repo.Find(ctx, mockDB.DB, userGroupID)
		assert.Nil(t, err)
		assert.NotNil(t, userGroup)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := userGroupV2RepoMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &userGroupID).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsUserGroup...).Once().Return(puddle.ErrClosedPool)
		userGroup, err := repo.Find(ctx, mockDB.DB, userGroupID)
		assert.Nil(t, userGroup)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUserGroupV2Repo_FindByIDs(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	r := new(UserGroupV2Repo)
	userGroupIDs := []string{idutil.ULIDNow(), idutil.ULIDNow()}
	userGroup := &entity.UserGroupV2{}
	userGroupFields := database.GetFieldNames(userGroup)
	scanFields := database.GetScanFields(userGroup, userGroupFields)

	tests := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "error when query",
			expectErr: fmt.Errorf("error when query"),
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userGroupIDs).Once().Return(mockDB.Rows, fmt.Errorf("error when query"))
			},
		},
		{
			name:      "closed pool",
			expectErr: puddle.ErrClosedPool,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userGroupIDs).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Times(len(userGroupIDs)).Return(true)
				mockDB.Rows.On("Scan", scanFields...).Times(len(userGroupIDs)).Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(puddle.ErrClosedPool)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "tx is closed",
			expectErr: pgx.ErrTxClosed,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userGroupIDs).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userGroupIDs).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Times(len(userGroupIDs)).Return(true)
				mockDB.Rows.On("Scan", scanFields...).Times(len(userGroupIDs)).Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.setup()
			userGroups, err := r.FindByIDs(ctx, mockDB.DB, userGroupIDs)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.NotNil(t, userGroups)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
		})
	}
}

func TestUserGroupV2Repo_FindUserGroupAndRoleByUserID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	r := new(UserGroupV2Repo)
	grantedRoleIDs := []string{idutil.ULIDNow(), idutil.ULIDNow()}
	userID := database.Text(idutil.ULIDNow())
	userGroup := entity.UserGroupV2{}
	userGroupFields := database.GetFieldNames(&userGroup)
	role := entity.Role{}
	roleFields := database.GetFieldNames(&role)
	scanFields := append(database.GetScanFields(&userGroup, userGroupFields), database.GetScanFields(&role, roleFields)...)

	tests := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Times(len(grantedRoleIDs)).Return(true)
				mockDB.Rows.On("Scan", scanFields...).Times(len(grantedRoleIDs)).Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "error when query",
			expectErr: fmt.Errorf("database.Select: error when query"),
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, fmt.Errorf("error when query"))
			},
		},
		{
			name:      "tx is closed",
			expectErr: pgx.ErrTxClosed,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.setup()
			userGroups, err := r.FindUserGroupAndRoleByUserID(ctx, mockDB.DB, userID)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.NotNil(t, userGroups)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
		})
	}
}

func TestUserRepo_FindUserGroupByRoleName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	roleName := constant.RoleStudent
	_, userValues := (&entity.UserGroupV2{}).FieldMap()
	argsUser := append([]interface{}{}, genSliceMock(len(userValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := userGroupV2RepoMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), roleName).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsUser...).Once().Return(nil)
		students, err := repo.FindUserGroupByRoleName(ctx, mockDB.DB, roleName)
		assert.Nil(t, err)
		assert.NotNil(t, students)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := userGroupV2RepoMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), roleName).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsUser...).Once().Return(puddle.ErrClosedPool)
		student, err := repo.FindUserGroupByRoleName(ctx, mockDB.DB, roleName)
		assert.Nil(t, student)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestUserGroupV2Repo_FindAndMapUserGroupAndRolesByUserID(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	r := new(UserGroupV2Repo)
	grantedRoleIDs := []string{idutil.ULIDNow(), idutil.ULIDNow()}
	userID := database.Text(idutil.ULIDNow())
	userGroup := entity.UserGroupV2{}
	userGroupFields := database.GetFieldNames(&userGroup)
	role := entity.Role{}
	roleFields := database.GetFieldNames(&role)
	scanFields := append(database.GetScanFields(&userGroup, userGroupFields), database.GetScanFields(&role, roleFields)...)

	tests := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Times(len(grantedRoleIDs)).Return(true)
				mockDB.Rows.On("Scan", scanFields...).Times(len(grantedRoleIDs)).Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "error when query",
			expectErr: fmt.Errorf("database.Select: error when query"),
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, fmt.Errorf("error when query"))
			},
		},
		{
			name:      "tx is closed",
			expectErr: pgx.ErrTxClosed,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", scanFields...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.setup()
			userGroups, err := r.FindAndMapUserGroupAndRolesByUserID(ctx, mockDB.DB, userID)
			if tt.expectErr != nil {
				assert.EqualError(t, err, tt.expectErr.Error())
			} else {
				assert.NotNil(t, userGroups)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
		})
	}
}
