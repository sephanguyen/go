package repository

import (
	"context"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
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

func DomainUserAddressRepoWithSqlMock() (*DomainUserAddressRepo, *testutil.MockDB) {
	r := &DomainUserAddressRepo{}
	return r, testutil.NewMockDB()
}

func TestDomainUserAddressRepo_SoftDeleteByUsers(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := DomainUserAddressRepoWithSqlMock()
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

		mockDB.RawStmt.AssertUpdatedTable(t, "user_address")
		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"user_id":    {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at": {HasNullTest: true},
		})
	})
}

func TestDomainUserAddressRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	repo, mockDB := DomainUserAddressRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []entity.DomainUserAddress{
				entity.DefaultDomainUserAddress{},
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
		err := repo.UpsertMultiple(ctx, mockDB.DB, testCase.req.([]entity.DomainUserAddress)...)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestDomainUserAddressRepo_GetByUserID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	userAddressRepo := NewUserAddress(entity.DefaultDomainUserAddress{})

	userID := pgtype.Text{}
	userID.Set(uuid.NewString())
	_, userAddressValues := userAddressRepo.FieldMap()
	argsUserAddresses := append([]interface{}{}, genSliceMock(len(userAddressValues))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := DomainUserAddressRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsUserAddresses...).Once().Return(nil)

		mockDB.Rows.On("Next").Once().Return(false)
		mockDB.Rows.On("Err").Once().Return(nil)
		mockDB.Rows.On("Close").Once().Return(nil)

		userAddresses, err := repo.GetByUserID(ctx, mockDB.DB, userID)
		assert.Nil(t, err)
		assert.NotNil(t, userAddresses)
	})

	t.Run("db Query returns error", func(t *testing.T) {
		repo, mockDB := DomainUserAddressRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, pgx.ErrTxClosed)

		userAddresses, err := repo.GetByUserID(ctx, mockDB.DB, userID)
		assert.NotNil(t, err)
		assert.Nil(t, userAddresses)
	})

	t.Run("rows Scan returns error", func(t *testing.T) {
		repo, mockDB := userAddressRepoWithSqlMock()
		mockDB.DB.On("Query", mock.Anything, mock.AnythingOfType("string"), &userID).Once().Return(mockDB.Rows, nil)
		mockDB.Rows.On("Next").Once().Return(true)
		mockDB.Rows.On("Scan", argsUserAddresses...).Once().Return(pgx.ErrTxClosed)
		mockDB.Rows.On("Close").Once().Return(nil)

		userAddresses, err := repo.GetByUserID(ctx, mockDB.DB, userID)
		assert.NotNil(t, err)
		assert.Nil(t, userAddresses)
	})
}
