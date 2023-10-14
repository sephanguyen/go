package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func InvoiceScheduleHistoryRepoWithSqlMock() (*InvoiceScheduleHistoryRepo, *testutil.MockDB) {
	repo := &InvoiceScheduleHistoryRepo{}
	return repo, testutil.NewMockDB()
}

func TestInvoiceScheduleHistoryRepo_Create(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.InvoiceScheduleHistory{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := InvoiceScheduleHistoryRepoWithSqlMock()

		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("create failed", func(t *testing.T) {
		repo, mockDB := InvoiceScheduleHistoryRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert InvoiceScheduleHistory: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("No rows affected after insert", func(t *testing.T) {
		repo, mockDB := InvoiceScheduleHistoryRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert InvoiceScheduleHistory: %w", pgx.ErrNoRows).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
}

func TestInvoiceScheduleHistoryRepo_UpdateWithFields(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.InvoiceScheduleHistory{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - UpdateWithFields", func(t *testing.T) {
		repo, mockDB := InvoiceScheduleHistoryRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - UpdateWithFields failed", func(t *testing.T) {
		repo, mockDB := InvoiceScheduleHistoryRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields InvoiceScheduleHistory: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceScheduleHistoryRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields InvoiceScheduleHistory: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
