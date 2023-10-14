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

func grantedRoleRepoWithSqlMock() (*GrantedRoleRepo, *testutil.MockDB) {
	repo := &GrantedRoleRepo{}
	return repo, testutil.NewMockDB()
}

func TestGrantedRole_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	grantedRoleRepo := &GrantedRoleRepo{}
	grantedRole := &entity.GrantedRole{}
	_, grantedRoleValues := grantedRole.FieldMap()
	argsUserGroup := append(
		[]interface{}{mock.Anything, mock.Anything},
		genSliceMock(len(grantedRoleValues))...,
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
			name:      "can not create granted role",
			expectErr: fmt.Errorf("err insert grantedrole: %w", fmt.Errorf("error")),
			setup: func() {
				cmtTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmtTag, fmt.Errorf("error"))
			},
		},
		{
			name:      "cannot upsert grantedrole",
			expectErr: fmt.Errorf("cannot upsert grantedrole"),
			setup: func() {
				cmtTag := pgconn.CommandTag([]byte(`0`))
				mockDB.DB.On("Exec", argsUserGroup...).Once().Return(cmtTag, nil)
			},
		},
	}

	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			err := grantedRoleRepo.Create(ctx, mockDB.DB, grantedRole)
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			}
		})
	}
}

func TestGrantedRole_LinkGrantedRoleToAccessPath(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDB := testutil.NewMockDB()
	batchResults := &mock_database.BatchResults{}
	grantedRoleRepo := &GrantedRoleRepo{}
	uid := idutil.ULIDNow()
	grantedRole := &entity.GrantedRole{
		GrantedRoleID: database.Text("granted-role-id" + uid),
		UserGroupID:   database.Text("user-group-id" + uid),
		RoleID:        database.Text("role-id" + uid),
		CreatedAt:     database.Timestamptz(time.Now()),
		UpdatedAt:     database.Timestamptz(time.Now()),
		DeletedAt:     database.Timestamptz(time.Time{}),
		ResourcePath:  database.Text(""),
	}

	testCases := []struct {
		name      string
		setup     func()
		expectErr error
	}{
		{
			name:      "happy case",
			expectErr: nil,
			setup: func() {
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults, nil)
				batchResults.On("Exec").Once().Return(pgconn.CommandTag([]byte(`1`)), nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "cannot exec Link Granted Role To AccessPath",
			expectErr: errors.Wrap(fmt.Errorf("error"), "batchResults.Exec"),
			setup: func() {
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults, nil)
				batchResults.On("Exec").Once().Return(nil, fmt.Errorf("error"))
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:      "cannot upsert grantedRoleAccessPath",
			expectErr: fmt.Errorf("cannot upsert grantedRoleAccessPath"),
			setup: func() {
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Return(batchResults, nil)
				batchResults.On("Exec").Once().Return(pgconn.CommandTag([]byte(`0`)), nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testcase := range testCases {
		testName := fmt.Sprintf("Test case: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			err := grantedRoleRepo.LinkGrantedRoleToAccessPath(ctx, mockDB.DB, grantedRole, []string{idutil.ULIDNow()})
			if err != nil {
				assert.EqualError(t, err, testcase.expectErr.Error())
			}

			mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
		})
	}
}

func TestGrantedRole_GetByUserGroup(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userGroupID := pgtype.Text{}
	userGroupID.Set(uuid.NewString())
	_, grantedRoleValues := (&entity.GrantedRole{}).FieldMap()
	argsGrantedRoles := append([]interface{}{}, genSliceMock(len(grantedRoleValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := grantedRoleRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userGroupID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsGrantedRoles...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByUserGroup(ctx, mockDB.DB, userGroupID)
		assert.Nil(t, err)
		assert.NotNil(t, grantedRoles)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := grantedRoleRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userGroupID).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		grantedRoles, err := repo.GetByUserGroup(ctx, mockDB.DB, userGroupID)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := grantedRoleRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userGroupID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsGrantedRoles...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		grantedRoles, err := repo.GetByUserGroup(ctx, mockDB.DB, userGroupID)
		assert.NotNil(t, err)
		assert.Nil(t, grantedRoles)
	})
}

func TestGrantedRole_Upsert(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	grantedRoleRepo := &GrantedRoleRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.GrantedRole{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: upsert multiple granted_role",
			req: []*entity.GrantedRole{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "success: resource path status null",
			req: []*entity.GrantedRole{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "success: skip role id status not equal present",
			req: []*entity.GrantedRole{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Null},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.GrantedRole{
				{
					GrantedRoleID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					UserGroupID:   pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					RoleID:        pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
					ResourcePath:  pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
			},
			expectedErr: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec"),
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
		err := grantedRoleRepo.Upsert(ctx, db, testCase.req.([]*entity.GrantedRole))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestGrantedRoleRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	repo, mockDB := grantedRoleRepoWithSqlMock()
	grantedRoleIDs := database.TextArray([]string{"granted-role-1", "granted-role-2"})

	testCases := []TestCase{
		{
			name:        "error cannot delete granted role",
			expectedErr: errors.New("cannot delete granted role"),
			setup: func(ctx context.Context) {
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &grantedRoleIDs).Once().Return(nil, errors.New("cannot delete granted role"))
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`2`))
				mockDB.DB.On("Exec", mock.Anything, mock.AnythingOfType("string"), &grantedRoleIDs).Once().Return(cmdTag, nil)
				mockDB.DB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.SoftDelete(ctx, mockDB.DB, grantedRoleIDs)
		if err != nil {
			assert.Equal(t, err.Error(), testCase.expectedErr.Error())
		} else {
			assert.Nil(t, err)
		}
	}
}
