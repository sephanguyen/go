package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserGroupRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	userGroupRepo := &UserGroupRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.UserGroup{
				{
					UserID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			name: "happy case: create multiple teachers",
			req: []*entity.UserGroup{
				{
					UserID: pgtype.Text{String: uuid.NewString(), Status: pgtype.Present},
				},
				{
					UserID: pgtype.Text{String: "2", Status: pgtype.Present},
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
			req: []*entity.UserGroup{
				{
					UserID: pgtype.Text{String: "1", Status: pgtype.Present},
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
			name: "send batch return ",
			req: []*entity.UserGroup{
				{
					UserID: pgtype.Text{String: "1", Status: pgtype.Present},
				},
			},
			expectedErr: errors.New("userGroup not inserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := userGroupRepo.CreateMultiple(ctx, db, testCase.req.([]*entity.UserGroup))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserGroupRepo_UpdateStatus(t *testing.T) {
	t.Parallel()
	userGroupRepo := &UserGroupRepo{}
	db := &mock_database.Ext{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: entity.UserGroup{
				UserID: pgtype.Text{String: "1"},
				Status: pgtype.Text{String: entity.UserGroupAdmin},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "not found record",
			req: entity.UserGroup{
				UserID: pgtype.Text{String: "-1"},
				Status: pgtype.Text{String: entity.UserGroupAdmin},
			},
			expectedErr: errors.New("not found any records"),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "connection closed",
			req: entity.UserGroup{
				UserID: pgtype.Text{String: "-1"},
				Status: pgtype.Text{String: entity.UserGroupAdmin},
			},
			expectedErr: errors.New("db.Exec: closed pool"),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				// cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, puddle.ErrClosedPool)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		userGroup := testCase.req.(entity.UserGroup)
		err := userGroupRepo.UpdateStatus(ctx, db, userGroup.UserID, userGroup.Status)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserGroupRepo_Upsert(t *testing.T) {
	t.Parallel()
	userGroupRepo := &UserGroupRepo{}
	db := &mock_database.Ext{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: &entity.UserGroup{
				UserID: pgtype.Text{String: "1", Status: pgtype.Present},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}

				cmdTag := pgconn.CommandTag([]byte(`1`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "connection closed",
			req: &entity.UserGroup{
				UserID: pgtype.Text{String: "1", Status: pgtype.Present},
			},
			expectedErr: errors.New("db.Exec: closed pool"),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, puddle.ErrClosedPool)
			},
		},
		{
			name: "no rows affected",
			req: &entity.UserGroup{
				UserID: pgtype.Text{String: "1", Status: pgtype.Present},
			},
			expectedErr: errors.New("cannot upsert userGroup"),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
					mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := userGroupRepo.Upsert(ctx, db, testCase.req.(*entity.UserGroup))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestUserGroupRepo_UpdateOrigin(t *testing.T) {
	t.Parallel()
	userGroupRepo := &UserGroupRepo{}
	db := &mock_database.Ext{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: entity.UserGroup{
				UserID:   database.Text(idutil.ULIDNow()),
				Status:   database.Text(entity.UserGroupSchoolAdmin),
				IsOrigin: database.Bool(false),
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag([]byte(successTag))
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "not found record",
			req: entity.UserGroup{
				UserID:   database.Text(idutil.ULIDNow()),
				Status:   database.Text(entity.UserGroupSchoolAdmin),
				IsOrigin: database.Bool(false),
			},
			expectedErr: fmt.Errorf("not found any records"),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				cmdTag := pgconn.CommandTag(failedTag)
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmdTag, nil)
			},
		},
		{
			name: "connection closed",
			req: entity.UserGroup{
				UserID:   database.Text(idutil.ULIDNow()),
				Status:   database.Text(entity.UserGroupSchoolAdmin),
				IsOrigin: database.Bool(false),
			},
			expectedErr: fmt.Errorf("db.Exec: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				db = &mock_database.Ext{}
				db.On("Exec", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, puddle.ErrClosedPool)
			},
		},
	}

	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		userGroup := testCase.req.(entity.UserGroup)
		err := userGroupRepo.UpdateOrigin(ctx, db, userGroup.UserID, userGroup.IsOrigin)
		assert.Equal(t, testCase.expectedErr, err)
	}
}

func TestUserGroupRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	userGroupRepo := &UserGroupRepo{}
	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entity.UserGroup{
				{UserID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(successTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: create multiple user groups",
			req: []*entity.UserGroup{
				{UserID: database.Text(idutil.ULIDNow())},
				{UserID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(successTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []*entity.UserGroup{
				{UserID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: fmt.Errorf("batchResults.Exec: %v", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "send batch return ",
			req: []*entity.UserGroup{
				{UserID: database.Text(idutil.ULIDNow())},
			},
			expectedErr: fmt.Errorf("user group not inserted"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(failedTag)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testCases {
		ctx := context.Background()
		testCase.setup(ctx)
		err := userGroupRepo.UpsertMultiple(ctx, db, testCase.req.([]*entity.UserGroup))
		assert.Equal(t, testCase.expectedErr, err)
	}
}
