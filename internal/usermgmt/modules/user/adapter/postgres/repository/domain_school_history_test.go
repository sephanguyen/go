package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainSchoolHistoryRepoWithSqlMock() (*DomainSchoolHistoryRepo, *testutil.MockDB) {
	r := &DomainSchoolHistoryRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainSchoolHistoryRepo_SetCurrentSchoolByStudentIDsAndSchoolIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainSchoolHistoryRepoWithSqlMock()
	userIDs := []string{"userID-1", "userID-2"}
	schoolIDs := []string{"userID-1", "userID-2"}

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(userIDs), database.TextArray(schoolIDs))

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
		},
		{
			name:        "err update",
			expectedErr: fmt.Errorf("err db.Exec: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			err := repo.SetCurrentSchoolByStudentIDsAndSchoolIDs(ctx, mockDB.DB, userIDs, schoolIDs)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
				mockDB.RawStmt.AssertUpdatedFields(t, "is_current")
				mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
					"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 2}},
					"school_id":  {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
				})
			}
		})
	}
}

func TestDomainSchoolHistoryRepo_SoftDeleteByStudentIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainSchoolHistoryRepoWithSqlMock()
	userIDs := []string{"userID-1", "userID-2"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(userIDs))

	testCases := []TestCase{
		{
			name: "happy case",
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)
			},
		},
		{
			name:        "err update",
			expectedErr: fmt.Errorf("err db.Exec: %w", puddle.ErrClosedPool),
			setup: func(ctx context.Context) {
				mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(ctx)
			}
			err := repo.SoftDeleteByStudentIDs(ctx, mockDB.DB, userIDs)
			if err != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.expectedErr)
				mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
				mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
					"student_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
					"deleted_at": {HasNullTest: true},
				})
			}
		})
	}
}

func TestDomainSchoolHistoryRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainSchoolHistoryRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []entity.DomainSchoolHistory{
				&SchoolHistory{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(ctx)
		err := repo.UpsertMultiple(ctx, mockDB.DB, testCase.req.([]entity.DomainSchoolHistory)...)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
