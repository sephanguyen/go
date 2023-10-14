package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BankBranchRepoWithSqlMock() (*BankBranchRepo, *testutil.MockDB) {
	repo := &BankBranchRepo{}
	return repo, testutil.NewMockDB()
}

func TestBankBranchRepo_FindRelatedBankOfBankBranches(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BankBranchRepoWithSqlMock()

	scanFields := make([]interface{}, 18)
	for i := 0; i < 18; i++ {
		scanFields[i] = mock.Anything
	}

	rows := mockDB.Rows

	ids := []string{"1", "2", "3"}

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.FindRelatedBankOfBankBranches(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindRelatedBankOfBankBranches(ctx, mockDB.DB, ids)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindRelatedBankOfBankBranches(ctx, mockDB.DB, ids)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBankBranchRepo_FindBankBranchesToExport(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BankBranchRepoWithSqlMock()

	scanFields := make([]interface{}, 15)
	for i := 0; i < 15; i++ {
		scanFields[i] = mock.Anything
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.FindExportableBankBranches(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindExportableBankBranches(ctx, mockDB.DB)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindExportableBankBranches(ctx, mockDB.DB)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBankBranchRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	bankBranchID := "example-bank-branch-id"
	mockE := &entities.BankBranch{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BankBranchRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByID(ctx, mockDB.DB, bankBranchID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("find bank branch failed", func(t *testing.T) {
		repo, mockDB := BankBranchRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByID(ctx, mockDB.DB, bankBranchID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByID BankBranchRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("No rows returned when finding bank branch by id", func(t *testing.T) {
		repo, mockDB := BankBranchRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByID(ctx, mockDB.DB, bankBranchID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByID BankBranchRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBankRepo_FindByBankBranchCodeAndBank(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	bankBranchCode := "example-bank-branch-code"
	bankID := "bank-id"
	mockE := &entities.BankBranch{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BankBranchRepoWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByBankBranchCodeAndBank(ctx, mockDB.DB, bankBranchCode, bankID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("find by bank branch code and bank failed", func(t *testing.T) {
		repo, mockDB := BankBranchRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByBankBranchCodeAndBank(ctx, mockDB.DB, bankBranchCode, bankID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByBankBranchCodeAndBank BankBranchRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("No rows returned when finding bank branch by code and bank", func(t *testing.T) {
		repo, mockDB := BankBranchRepoWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByBankBranchCodeAndBank(ctx, mockDB.DB, bankBranchCode, bankID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByBankBranchCodeAndBank BankBranchRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
