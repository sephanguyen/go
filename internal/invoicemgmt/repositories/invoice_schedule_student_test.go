package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func InvoiceScheduleStudentRepoWithSqlMock() (*InvoiceScheduleStudentRepo, *testutil.MockDB) {
	repo := &InvoiceScheduleStudentRepo{}
	return repo, testutil.NewMockDB()
}

func TestInvoiceScheduleStudentRepo_CreateMultiple(t *testing.T) {
	t.Parallel()

	e1 := &entities.InvoiceScheduleStudent{}
	e2 := &entities.InvoiceScheduleStudent{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case 1 input", func(t *testing.T) {
		parentRepo, mockDB := InvoiceScheduleStudentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entities.InvoiceScheduleStudent{e1})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("happy case (2 parents)", func(t *testing.T) {
		parentRepo, mockDB := InvoiceScheduleStudentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entities.InvoiceScheduleStudent{e1, e2})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("batch Exec failed", func(t *testing.T) {
		parentRepo, mockDB := InvoiceScheduleStudentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`1`))
		batchResults.On("Exec").Return(cmdTag, pgx.ErrTxClosed)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entities.InvoiceScheduleStudent{e1})
		assert.EqualError(t, errors.Wrap(pgx.ErrTxClosed, "batchResults.Exec"), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})

	t.Run("batch Exec no rows affected", func(t *testing.T) {
		parentRepo, mockDB := InvoiceScheduleStudentRepoWithSqlMock()

		batchResults := &mock_database.BatchResults{}
		mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
		cmdTag := pgconn.CommandTag([]byte(`0`))
		batchResults.On("Exec").Return(cmdTag, nil)
		batchResults.On("Close").Once().Return(nil)
		err := parentRepo.CreateMultiple(ctx, mockDB.DB, []*entities.InvoiceScheduleStudent{e1})
		assert.EqualError(t, fmt.Errorf("invoiceScheduleStudents not inserted"), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, batchResults)
	})
}
