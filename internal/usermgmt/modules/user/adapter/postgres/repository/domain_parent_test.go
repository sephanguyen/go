package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainParentRepoWithSqlMock() (*DomainParentRepo, *testutil.MockDB) {
	r := &DomainParentRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainParentRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	domainParentRepo := DomainParentRepo{
		UserRepo: mockDomainUserRepo{
			upsertMultipleFn: func(ctx context.Context, db database.QueryExecer, isEnableUsername bool, usersToCreate ...entity.User) error {
				return nil
			},
		},
		LegacyUserGroupRepo: mockLegacyUserGroupRepo{
			createMultipleFn: func(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error {
				return nil
			},
		},
		UserAccessPathRepo: mockDomainUserAccessPathRepo{
			upsertMultipleFn: func(ctx context.Context, db database.QueryExecer, userAccessPaths ...entity.DomainUserAccessPath) error {
				return nil
			},
			softDeleteByUserIDsFn: func(ctx context.Context, db database.QueryExecer, userIDs []string) error {
				return nil
			},
		},
		UserGroupMemberRepo: mockDomainUserGroupMemberRepo{
			createMultipleFn: func(ctx context.Context, db database.QueryExecer, userGroupMembers ...entity.DomainUserGroupMember) error {
				return nil
			},
		},
	}

	mockDB := testutil.NewMockDB()

	parentRepoEnt := Parent{}
	_, parentValues := parentRepoEnt.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(parentValues))...)
	testCases := []TestCase{
		{
			name: "happy case",
			req: aggregate.DomainParent{
				DomainParent:     entity.NullDomainParent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
			},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(`1`)
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "failed to create",
			req: aggregate.DomainParent{
				DomainParent:     entity.NullDomainParent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "batchResults.Exec")},
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec", args...).Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "failed to create by user repo",
			req: aggregate.DomainParent{
				DomainParent:     entity.NullDomainParent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "repo.UserRepo.UpsertMultiple")},
			setup: func(ctx context.Context) {
				domainParentRepo.UserRepo = mockDomainUserRepo{
					upsertMultipleFn: func(ctx context.Context, db database.QueryExecer, isEnableUsername bool, usersToCreate ...entity.User) error {
						return puddle.ErrClosedPool
					},
				}
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec", args...).Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "failed to create by legacy user group repo",
			req: aggregate.DomainParent{
				DomainParent:     entity.NullDomainParent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "repo.UserRepo.UpsertMultiple")},
			setup: func(ctx context.Context) {
				domainParentRepo.LegacyUserGroupRepo = mockLegacyUserGroupRepo{
					createMultipleFn: func(ctx context.Context, db database.QueryExecer, legacyUserGroups ...entity.LegacyUserGroup) error {
						return puddle.ErrClosedPool
					},
				}
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec", args...).Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "failed to create by user group member repo",
			req: aggregate.DomainParent{
				DomainParent:     entity.NullDomainParent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "repo.UserRepo.UpsertMultiple")},
			setup: func(ctx context.Context) {
				domainParentRepo.UserGroupMemberRepo = mockDomainUserGroupMemberRepo{
					createMultipleFn: func(ctx context.Context, db database.QueryExecer, userGroupMembers ...entity.DomainUserGroupMember) error {
						return puddle.ErrClosedPool
					},
				}
				batchResults := &mock_database.BatchResults{}
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec", args...).Return(nil, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			if tt.setup != nil {
				tt.setup(ctx)
			}
			err := domainParentRepo.
				UpsertMultiple(ctx, mockDB.DB, true, tt.req.(aggregate.DomainParent))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}

func TestDomainParentRepo_GetUsersByExternalUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	externalUserIDs := []string{"test-01@test.com", "test-02@test.com"}
	_, userRepoEnt := (&User{}).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)
	repo, mockDB := DomainParentRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req:  externalUserIDs,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name: "db Query returns error",
			req:  externalUserIDs,
			expectedErr: InternalError{
				errors.Wrap(pgx.ErrTxClosed, "db.Query"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name: "rows Scan returns error",
			req:  externalUserIDs,
			expectedErr: InternalError{
				RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan"),
			},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainUsers...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetUsersByExternalUserIDs(ctx, mockDB.DB, tt.req.([]string))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}
