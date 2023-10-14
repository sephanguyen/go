package repository

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainStudentRepoWithSqlMock() (*DomainStudentRepo, *testutil.MockDB) {
	r := &DomainStudentRepo{}
	return r, testutil.NewMockDB()
}

type mockDomainUserAccessPathRepo struct {
	upsertMultipleFn      func(ctx context.Context, db database.QueryExecer, userAccessPaths ...entity.DomainUserAccessPath) error
	softDeleteByUserIDsFn func(ctx context.Context, db database.QueryExecer, userIDs []string) error
}

func (mockDomainUserAccessPathRepo mockDomainUserAccessPathRepo) UpsertMultiple(ctx context.Context, db database.QueryExecer, userAccessPaths ...entity.DomainUserAccessPath) error {
	return mockDomainUserAccessPathRepo.upsertMultipleFn(ctx, db, userAccessPaths...)
}

func (mockDomainUserAccessPathRepo mockDomainUserAccessPathRepo) SoftDeleteByUserIDs(ctx context.Context, db database.QueryExecer, userIDs []string) error {
	return mockDomainUserAccessPathRepo.softDeleteByUserIDsFn(ctx, db, userIDs)
}

type mockDomainUserGroupMemberRepo struct {
	createMultipleFn func(ctx context.Context, db database.QueryExecer, userGroupMembers ...entity.DomainUserGroupMember) error
}

func (mockDomainUserGroupMemberRepo mockDomainUserGroupMemberRepo) CreateMultiple(ctx context.Context, db database.QueryExecer, userGroupMembers ...entity.DomainUserGroupMember) error {
	return mockDomainUserGroupMemberRepo.createMultipleFn(ctx, db, userGroupMembers...)
}

func TestDomainStudentRepo_Upsert(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	domainStudentRepo := DomainStudentRepo{
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

	studentRepoEnt := Student{}

	_, studentValues := studentRepoEnt.FieldMap()
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(studentValues))...)

	testCases := []TestCase{
		{
			name: "happy case",
			req: aggregate.DomainStudent{
				DomainStudent:    entity.NullDomainStudent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				Locations: entity.DomainLocations{
					entity.NullDomainLocation{},
				},
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
			req: aggregate.DomainStudent{
				DomainStudent:    entity.NullDomainStudent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				Locations: entity.DomainLocations{
					entity.NullDomainLocation{},
				},
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
			req: aggregate.DomainStudent{
				DomainStudent:    entity.NullDomainStudent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				Locations: entity.DomainLocations{
					entity.NullDomainLocation{},
				},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "repo.UserRepo.UpsertMultiple")},
			setup: func(ctx context.Context) {
				domainStudentRepo.UserRepo = mockDomainUserRepo{
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
			req: aggregate.DomainStudent{
				DomainStudent:    entity.NullDomainStudent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				Locations: entity.DomainLocations{
					entity.NullDomainLocation{},
				},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "repo.UserRepo.UpsertMultiple")},
			setup: func(ctx context.Context) {
				domainStudentRepo.LegacyUserGroupRepo = mockLegacyUserGroupRepo{
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
			req: aggregate.DomainStudent{
				DomainStudent:    entity.NullDomainStudent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				Locations: entity.DomainLocations{
					entity.NullDomainLocation{},
				},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "repo.UserRepo.UpsertMultiple")},
			setup: func(ctx context.Context) {
				domainStudentRepo.UserGroupMemberRepo = mockDomainUserGroupMemberRepo{
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
		{
			name: "failed to create by user access path repo",
			req: aggregate.DomainStudent{
				DomainStudent:    entity.NullDomainStudent{},
				LegacyUserGroups: entity.LegacyUserGroups{entity.EmptyLegacyUserGroup{}},
				UserGroupMembers: entity.DomainUserGroupMembers{entity.EmptyUserGroupMember{}},
				Locations: entity.DomainLocations{
					entity.NullDomainLocation{},
				},
			},
			expectedErr: InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "repo.UserAccessPathRepo.upsertMultiple")},
			setup: func(ctx context.Context) {
				domainStudentRepo.UserAccessPathRepo = mockDomainUserAccessPathRepo{
					upsertMultipleFn: func(ctx context.Context, db database.QueryExecer, userAccessPaths ...entity.DomainUserAccessPath) error {
						return puddle.ErrClosedPool
					},
					softDeleteByUserIDsFn: func(ctx context.Context, db database.QueryExecer, userIDs []string) error {
						return nil
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
			if tt.setup != nil {
				tt.setup(ctx)
			}
			err := domainStudentRepo.
				UpsertMultiple(ctx, mockDB.DB, true, tt.req.(aggregate.DomainStudent))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}

}

func TestDomainStudentRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentIDs := []string{"student-id-001", "student-id-002"}
	_, studentRepoEnt := (&Student{}).FieldMap()
	argsDomainStudents := append([]interface{}{}, genSliceMock(len(studentRepoEnt))...)
	repo, mockDB := DomainStudentRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req:  studentIDs,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainStudents...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			req:         studentIDs,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			req:         studentIDs,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainStudents...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByIDs(ctx, mockDB.DB, tt.req.([]string))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}

}

func TestDomainStudentRepo_GetUsersByExternalUserIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	externalUserIDs := []string{"test-01@test.com", "test-02@test.com"}
	_, userRepoEnt := (&User{}).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)
	repo, mockDB := DomainStudentRepoWithSqlMock()

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

func TestDomainStudentRepo_GetByEmails(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	emails := []string{"test-01@test.com", "test-02@test.com"}
	_, userRepoEnt := (&User{}).FieldMap()
	argsDomainUsers := append([]interface{}{}, genSliceMock(len(userRepoEnt))...)

	_, studentRepoEnt := (&Student{}).FieldMap()
	argsDomainStudents := append([]interface{}{}, genSliceMock(len(studentRepoEnt))...)
	repo, mockDB := DomainStudentRepoWithSqlMock()

	argsDomainStudents = append(argsDomainUsers, argsDomainStudents...)

	testCases := []TestCase{
		{
			name: "happy case",
			req:  emails,
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainStudents...).Once().Return(nil)

				mockDB.Rows.On("Next").Once().Return(false)
				mockDB.Rows.On("Err").Once().Return(nil)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "db Query returns error",
			req:         emails,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "db.Query")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, pgx.ErrTxClosed)
			},
		},
		{
			name:        "rows Scan returns error",
			req:         emails,
			expectedErr: InternalError{RawError: errors.Wrap(pgx.ErrTxClosed, "rows.Scan")},
			setup: func(ctx context.Context) {
				mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Once().Return(mockDB.Rows, nil)
				mockDB.Rows.On("Next").Once().Return(true)
				mockDB.Rows.On("Scan", argsDomainStudents...).Once().Return(pgx.ErrTxClosed)
				mockDB.Rows.On("Close").Once().Return(nil)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			_, err := repo.GetByEmails(ctx, mockDB.DB, tt.req.([]string))
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
			}
		})
	}
}
