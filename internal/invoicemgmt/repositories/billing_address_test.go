package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/puddle"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BillingAddressRepoWithSqlMock() (*BillingAddressRepo, *testutil.MockDB) {
	repo := &BillingAddressRepo{}
	return repo, testutil.NewMockDB()
}

func TestBillingAddressRepo_FindByUserID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := "example-student-id"
	mockE := &entities.BillingAddress{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BillingAddressRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByUserID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("find billing address failed", func(t *testing.T) {
		repo, mockDB := BillingAddressRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByUserID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByUserID BillingAddressRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("No rows returned when finding billing address by user id", func(t *testing.T) {
		repo, mockDB := BillingAddressRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByUserID(ctx, mockDB.DB, studentID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByUserID BillingAddressRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBillingAddressRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := "example-student-id"
	mockE := &entities.BillingAddress{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BillingAddressRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("find billing address failed", func(t *testing.T) {
		repo, mockDB := BillingAddressRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByID BillingAddressRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("No rows returned when finding billing address by id", func(t *testing.T) {
		repo, mockDB := BillingAddressRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByID(ctx, mockDB.DB, studentID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByID BillingAddressRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

type TestCase struct {
	name        string
	req         interface{}
	expectedErr error
	setup       func(ctx context.Context)
}

func TestBillingAddressRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BillingAddressRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.BillingAddress{
				{BillingAddressID: database.Text("existing-billing-address-id")},
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
		{
			name: "err when exec batch",
			req: []*entities.BillingAddress{
				{BillingAddressID: database.Text("existing-billing-address-id")},
			},
			expectedErr: errors.Wrap(errors.New("err when exec"), "batchResults.Exec"),
			setup: func(ctx context.Context) {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, errors.New("err when exec"))
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(ctx)
		err := repo.Upsert(ctx, mockDB.DB, testCase.req.([]*entities.BillingAddress)...)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestBillingAddressRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BillingAddressRepoWithSqlMock()
	billingAddressIDs := []string{"example-billing-address-id-1", "example-billing-address-id-2"}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(billingAddressIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.SoftDelete(ctx, mockDB.DB, billingAddressIDs...)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(billingAddressIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDelete(ctx, mockDB.DB, billingAddressIDs...)
		assert.Nil(t, err)

		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"billing_address_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at":         {HasNullTest: true},
		})
	})
}
