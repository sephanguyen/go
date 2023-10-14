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
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserGroupMember_UpsertBatch(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	repo := &UserGroupsMemberRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.UserGroupMember{
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
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
			name: "happy case: upsert multiple user group",
			req: []*entity.UserGroupMember{
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
				},
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Present},
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
			req: []*entity.UserGroupMember{
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
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
			name: "success: skip user group id status not equal present",
			req: []*entity.UserGroupMember{
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Null},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
				},
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					ResourcePath: pgtype.Text{String: fmt.Sprint(constants.ManabieSchool), Status: pgtype.Null},
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
			req: []*entity.UserGroupMember{
				{
					UserID:       pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
					UserGroupID:  pgtype.Text{String: idutil.ULIDNow(), Status: pgtype.Present},
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
		err := repo.UpsertBatch(ctx, db, testCase.req.([]*entity.UserGroupMember))
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestUserGroupMember_AssignWithUserGroup(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	userGroupsMemberRepo := &UserGroupsMemberRepo{}
	effectedOne := pgconn.CommandTag([]byte(`1`))
	effectedNil := pgconn.CommandTag([]byte(`0`))
	defaultResourcePath := database.Text(fmt.Sprint(constants.ManabieSchool))
	emptyResourcePath := pgtype.Text{Status: pgtype.Null}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.LegacyUser{
				{
					ID:           database.Text(uuid.NewString()),
					ResourcePath: defaultResourcePath,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(effectedOne, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: ResourcePath status nil",
			req: []*entity.LegacyUser{
				{
					ID:           database.Text(uuid.NewString()),
					ResourcePath: emptyResourcePath,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(effectedOne, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: create multiple student",
			req: []*entity.LegacyUser{
				{
					ID:           database.Text(uuid.NewString()),
					ResourcePath: defaultResourcePath,
				},
				{
					ID:           database.Text(uuid.NewString()),
					ResourcePath: defaultResourcePath,
				},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(effectedOne, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.LegacyUser{
				{
					ID:           database.Text(uuid.NewString()),
					ResourcePath: defaultResourcePath,
				},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "send batch return no record inserted",
			req: []*entity.LegacyUser{
				{
					ID:           database.Text(uuid.NewString()),
					ResourcePath: defaultResourcePath,
				},
			},
			expectedErr: errors.New("user group member not inserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(effectedNil, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := userGroupsMemberRepo.AssignWithUserGroup(ctx, db, testCase.req.([]*entity.LegacyUser), database.Text(idutil.ULIDNow()))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserGroupsMemberRepo_SoftDelete(t *testing.T) {
	t.Parallel()

	mockDB := &mock_database.Ext{}
	repo := &UserGroupsMemberRepo{}
	userGroupMembers := []*entity.UserGroupMember{
		{
			UserID:      database.Text(idutil.ULIDNow()),
			UserGroupID: database.Text(idutil.ULIDNow()),
		},
		{
			UserID:      database.Text(idutil.ULIDNow()),
			UserGroupID: database.Text(idutil.ULIDNow()),
		},
	}

	testCases := []TestCase{
		{
			name:        "error cannot delete user group members",
			expectedErr: fmt.Errorf("error"),
			setup: func(ctx context.Context) {
				mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(fmt.Sprint(len(userGroupMembers))))
				mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(cmdTag, nil)
				mockDB.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.SoftDelete(ctx, mockDB, userGroupMembers)
		assert.Equal(t, err, testCase.expectedErr)
	}
}

func TestUserGroupsMemberRepo_GetByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo := &UserGroupsMemberRepo{}
	id := database.Text(idutil.ULIDNow())
	userGroupMember := &entity.UserGroupMember{}
	fields, values := userGroupMember.FieldMap()

	tests := []struct {
		name        string
		ctx         context.Context
		expectedErr error
		setup       func(context.Context) *mock_database.Ext
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
			_, err := repo.GetByUserID(testCase.ctx, db, id)
			assert.Equal(t, testCase.expectedErr, err)
		})
	}
}
