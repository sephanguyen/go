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
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BankAccountRepoWithSqlMock() (*BankAccountRepo, *testutil.MockDB) {
	repo := &BankAccountRepo{}
	return repo, testutil.NewMockDB()
}

func TestBankAccountRepo_FindByStudentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := "example-student-id"
	mockE := &entities.BankAccount{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BankAccountRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("failed to find bank account", func(t *testing.T) {
		repo, mockDB := BankAccountRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByID BankAccountRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows returned after queried to find bank account", func(t *testing.T) {
		repo, mockDB := BankAccountRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByID(ctx, mockDB.DB, studentID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByID BankAccountRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBankAccountRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	bankAccountID := "example-bank-account-id"
	mockE := &entities.BankAccount{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BankAccountRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByID(ctx, mockDB.DB, bankAccountID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("find billing address failed", func(t *testing.T) {
		repo, mockDB := BankAccountRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByID(ctx, mockDB.DB, bankAccountID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByID BankAccountRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("No rows returned when finding billing address by id", func(t *testing.T) {
		repo, mockDB := BankAccountRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByID(ctx, mockDB.DB, bankAccountID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByID BankAccountRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBankAccountRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BankAccountRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.BankAccount{
				{BankAccountID: database.Text("existing-bank-account-id")},
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
			req: []*entities.BankAccount{
				{BankAccountID: database.Text("existing-bank-account-id")},
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
		err := repo.Upsert(ctx, mockDB.DB, testCase.req.([]*entities.BankAccount)...)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
