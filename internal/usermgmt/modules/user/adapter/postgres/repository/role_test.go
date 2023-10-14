package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	successTag = []byte(`1`)
	failedTag  = []byte(`0`)
)

func RoleRepoWithSqlMock() (*RoleRepo, *testutil.MockDB) {
	repo := &RoleRepo{}
	return repo, testutil.NewMockDB()
}

func TestRoleRepo_GetRolesByRoleIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	roleRepo := &RoleRepo{}
	role := &entity.Role{}
	roleP := new(entity.Role)
	fields := database.GetFieldNames(role)
	mockDB := testutil.NewMockDB()
	rows := &mock_database.Rows{}

	IDs := []string{idutil.ULIDNow(), idutil.ULIDNow()}
	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case query not include deleted",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(IDs) - 1).Return(true)
				rows.On("Scan", database.GetScanFields(roleP, fields)...).Times(len(IDs) - 1).Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(nil)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "error when querying",
			expectErr: fmt.Errorf("failed to get roles: %w", fmt.Errorf("error when querying")),
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error when querying"))
			},
		},
		{
			name:      "error occurred row",
			expectErr: errors.Wrap(fmt.Errorf("error when scanning"), "rows.Scan"),
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", database.GetScanFields(roleP, fields)...).Once().Return(fmt.Errorf("error when scanning"))
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "error occurred row",
			expectErr: errors.Wrap(fmt.Errorf("error occurred row"), "rows.Err"),
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(IDs) - 1).Return(true)
				rows.On("Scan", database.GetScanFields(roleP, fields)...).Times(len(IDs) - 1).Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Err").Once().Return(fmt.Errorf("error occurred row"))
				rows.On("Close").Once().Return(nil)
			},
		},
	}

	for index, testcase := range testCases {
		testName := fmt.Sprintf("%s-%d", testcase.name, index)
		t.Run(testName, func(t *testing.T) {
			fmt.Println(testName)
			testcase.setup()
			role, err := roleRepo.GetRolesByRoleIDs(ctx, mockDB.DB, database.TextArray(IDs))
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			} else {
				assert.NotNil(t, role)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, rows)
		})
	}
}

func TestRoleRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	roleRepo := &RoleRepo{}
	roleEntity := &entity.Role{}
	_, roleEntityValues := roleEntity.FieldMap()
	argsRole := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(roleEntityValues))...,
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
				cmtTag := pgconn.CommandTag([]byte(successTag))
				mockDB.DB.On("Exec", argsRole...).Once().Return(cmtTag, nil)
			},
		},
		{
			name:      "can not insert role",
			expectErr: fmt.Errorf("err insert role: %w", fmt.Errorf("error")),
			setup: func() {
				cmtTag := pgconn.CommandTag([]byte(successTag))
				mockDB.DB.On("Exec", argsRole...).Once().Return(cmtTag, fmt.Errorf("error"))
			},
		},
		{
			name:      "inserted but not rown effected",
			expectErr: fmt.Errorf("no row effected"),
			setup: func() {
				cmtTag := pgconn.CommandTag([]byte(failedTag))
				mockDB.DB.On("Exec", argsRole...).Once().Return(cmtTag, nil)
			},
		},
	}

	for index, testcase := range testCases {
		testName := fmt.Sprintf("%s-%d", testcase.name, index)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			err := roleRepo.Create(ctx, mockDB.DB, roleEntity)
			assert.Equal(t, err, testcase.expectErr)
		})

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	}
}

func TestRoleRepo_UpsertPermission(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	roleRepo := &RoleRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.PermissionRole{
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: upsert multiple granted_role",
			req: []*entity.PermissionRole{
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "success: resource path status null",
			req: []*entity.PermissionRole{
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "success: skip role id status not equal present",
			req: []*entity.PermissionRole{
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Null},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.PermissionRole{
				{
					PermissionID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:       pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := roleRepo.UpsertPermission(ctx, db, testCase.req.([]*entity.PermissionRole))
		assert.Equal(t, err, testCase.expectedErr)
	}
}

func TestRoleRepo_GetByName(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	roleName := database.Text(fmt.Sprintf("role-%s", idutil.ULIDNow()))
	_, schoolValues := (&entity.School{}).FieldMap()
	argsSchool := append([]interface{}{}, genSliceMock(len(schoolValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := RoleRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &roleName, mock.Anything).Once().Return(mockDB.Row, nil)
		mockDB.Row.On("Scan", argsSchool...).Once().Return(nil)
		school, err := repo.GetByName(ctx, mockDB.DB, roleName)
		assert.Nil(t, err)
		assert.NotNil(t, school)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("query row error", func(t *testing.T) {
		repo, mockDB := RoleRepoWithSqlMock()
		mockDB.DB.On("QueryRow", mock.Anything, mock.AnythingOfType("string"), &roleName, mock.Anything).Once().Return(mockDB.Row)
		mockDB.Row.On("Scan", argsSchool...).Once().Return(puddle.ErrClosedPool)
		school, err := repo.GetByName(ctx, mockDB.DB, roleName)
		assert.Nil(t, school)
		assert.Error(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestRoleRepo_GetRolesByUserGroupIDs(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()
	roleRepo := new(RoleRepo)
	userIDs := database.TextArray([]string{idutil.ULIDNow()})
	userGroupID := ""
	role := entity.Role{}
	roleFields := database.GetFieldNames(&role)
	scanFields := append(database.GetScanFields(&role, roleFields), &userGroupID)

	tests := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userIDs).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Times(len(userIDs.Elements)).Return(true)
				mockDB.Rows.On("Scan", scanFields...).Times(len(userIDs.Elements)).Return(nil)
				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "error when query",
			expectErr: fmt.Errorf("database.Select: %w", fmt.Errorf("error")),
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userIDs).Once().Return(mockDB.Rows, fmt.Errorf("error"))
			},
		},
		{
			name:      "tx is closed",
			expectErr: pgx.ErrTxClosed,
			setup: func() {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userIDs).Once().Return(mockDB.Rows, nil)
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
			userGroups, err := roleRepo.GetRolesByUserGroupIDs(ctx, mockDB.DB, userIDs)
			assert.Equal(t, err, tt.expectErr)
			if tt.expectErr == nil {
				assert.NotNil(t, userGroups)
			} else {
				assert.Nil(t, userGroups)
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
		})
	}
}
func TestRoleRepo_FindBelongedRoles(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo := &RoleRepo{}
	id := database.Text(idutil.ULIDNow())
	role := &entity.Role{}
	fields, values := role.FieldMap()

	tests := []struct {
		name         string
		ctx          context.Context
		expectedErr  error
		expectedResp bool
		setup        func(context.Context) *mock_database.Ext
	}{
		{
			name:        "happy case",
			ctx:         ctx,
			expectedErr: nil,
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, nil, mock.Anything, mock.AnythingOfType("string"), &id)
				mockDB.MockScanArray(nil, fields, [][]interface{}{values})
				return mockDB.DB
			},
		},
		{
			name:        "error when execute query",
			ctx:         ctx,
			expectedErr: fmt.Errorf("database.Select: %w", fmt.Errorf("err db.Query: %w", puddle.ErrClosedPool)),
			setup: func(ctx context.Context) *mock_database.Ext {
				mockDB := testutil.NewMockDB()
				mockDB.MockQueryArgs(t, puddle.ErrClosedPool, mock.Anything, mock.AnythingOfType("string"), &id)
				return mockDB.DB
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			db := testCase.setup(testCase.ctx)
			userGroups, err := repo.FindBelongedRoles(testCase.ctx, db, id)

			if testCase.expectedErr != nil {
				assert.Equal(t, testCase.expectedErr.Error(), err.Error())
			} else {
				assert.NotNil(t, userGroups)
			}
		})
	}
}
