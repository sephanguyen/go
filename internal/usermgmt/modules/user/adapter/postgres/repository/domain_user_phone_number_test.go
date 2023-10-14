package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DomainUserPhoneNumberRepoWithSqlMock() (*DomainUserPhoneNumberRepo, *testutil.MockDB) {
	r := &DomainUserPhoneNumberRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainUserPhoneNumberRepo_SoftDeleteByUsers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainUserPhoneNumberRepoWithSqlMock()
	userIDs := []string{"userID-1", "userID-2"}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(userIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.SoftDeleteByUserIDs(ctx, mockDB.DB, userIDs)
		assert.Equal(t, err.Error(), InternalError{RawError: errors.Wrap(puddle.ErrClosedPool, "db.Exec")}.Error())
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

func TestDomainUserPhoneNumberRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	repo, mockDB := DomainUserPhoneNumberRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []entity.DomainUserPhoneNumber{
				&UserPhoneNumber{},
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
		ctx := context.Background()
		testCase.setup(ctx)
		err := repo.UpsertMultiple(ctx, mockDB.DB, testCase.req.([]entity.DomainUserPhoneNumber)...)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestDomainUserPhoneNumberRepo_GetByUserID(t *testing.T) {
	ctx := auth.InjectFakeJwtToken(context.Background(), fmt.Sprint(constants.ManabieSchool))
	db := new(mock_database.QueryExecer)
	userIDs := []string{idutil.ULIDNow()}

	var attrs []interface{}
	fieldMap, _ := NewDomainUserPhoneNumber(&entity.DefaultDomainUserPhoneNumber{}).FieldMap()
	for range fieldMap {
		attrs = append(attrs, mock.Anything)
	}

	tests := []struct {
		name    string
		userIDs []string
		wantErr error
		setup   func()
	}{
		{
			name:    "happy case",
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
			wantErr: InternalError{errors.Wrap(fmt.Errorf("error"), "db.Query")},
			setup: func() {
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("error"))
			},
		},
		{
			name:    "error: rows.Err error",
			wantErr: InternalError{errors.Wrap(fmt.Errorf("error"), "rows.Err")},
			setup: func() {
				rows := &mock_database.Rows{}
				db.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Close").Once().Return()
				rows.On("Err").Once().Return(fmt.Errorf("error"))
			},
		},
		{
			name:    "error: rows.Scan error",
			wantErr: InternalError{errors.Wrap(fmt.Errorf("error"), "rows.Scan")},
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
		t.Run(tt.name, func(t *testing.T) {
			ut := new(DomainUserPhoneNumberRepo)
			if tt.setup != nil {
				tt.setup()
			}
			_, err := ut.GetByUserIDs(ctx, db, userIDs)
			if err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.Nil(t, tt.wantErr)
			}
		})
	}
}
