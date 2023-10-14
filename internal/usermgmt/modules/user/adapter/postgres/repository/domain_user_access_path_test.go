package repository

import (
	"context"
	"errors"
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

func DomainUserAccessPathRepoWithSqlMock() (*DomainUserAccessPathRepo, *testutil.MockDB) {
	r := &DomainUserAccessPathRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainUserAccessPathRepo_upsertMultiple(t *testing.T) {
	t.Parallel()
	db := &mock_database.Ext{}
	domainUserAccessPathRepo := &DomainUserAccessPathRepo{}

	testCases := []TestCase{
		{
			name: "happy case",
			req: []entity.DomainUserAccessPath{
				entity.DefaultUserAccessPath{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(`1`)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "happy case: create multiple teachers",
			req: []entity.DomainUserAccessPath{
				entity.DefaultUserAccessPath{},
				entity.DefaultUserAccessPath{},
			},
			expectedErr: nil,
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag(`1`)
				db.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name: "error send batch",
			req: []entity.DomainUserAccessPath{
				entity.DefaultUserAccessPath{},
			},
			expectedErr: errors.New("batchResults.Exec: closed pool"),
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
		err := domainUserAccessPathRepo.UpsertMultiple(ctx, db, testCase.req.([]entity.DomainUserAccessPath)...)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestDomainUserAccessPathRepo_SoftDeleteByUsers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainUserAccessPathRepoWithSqlMock()
	userIDs := []string{"userID-1", "userID-2"}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(userIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.SoftDeleteByUserIDs(ctx, mockDB.DB, userIDs)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(userIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDeleteByUserIDs(ctx, mockDB.DB, userIDs)
		assert.Nil(t, err)

		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestDomainUserAccessPathRepo_GetByUserIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	db := new(mock_database.QueryExecer)

	var attrs []interface{}
	fieldMap, _ := newDomainUserAccessPath(&entity.DefaultUserAccessPath{}).FieldMap()
	for range fieldMap {
		attrs = append(attrs, mock.Anything)
	}

	userIDs := []string{"userID-1", "userID-2"}

	tests := []struct {
		name    string
		setup   func()
		userIDs []string
		wantErr error
	}{
		{
			name:    "success",
			wantErr: nil,
			setup: func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(len(userIDs)).Return(true)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Scan", attrs...).Times(len(userIDs)).Return(nil)
			},
		},
		{
			name:    "error: db.Query error",
			wantErr: fmt.Errorf("db.Query: %v", fmt.Errorf("error")),
			setup: func() {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:    "error: rows.Err error",
			wantErr: fmt.Errorf("rows.Err: %v", fmt.Errorf("error")),
			setup: func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:    "error: rows.Scan error",
			wantErr: fmt.Errorf("rows.Scan: %v", fmt.Errorf("error")),
			setup: func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(nil)
				rows.On("Next").Once().Return(true)
				rows.On("Scan", attrs...).Once().Return(fmt.Errorf("error"))
			},
		},
	}

	for _, tt := range tests {
		ut := new(DomainUserAccessPathRepo)
		if tt.setup != nil {
			tt.setup()
		}
		_, err := ut.GetByUserIDs(ctx, db, userIDs)
		assert.Equal(t, err, tt.wantErr)
	}
}
