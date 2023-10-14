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

func InvoiceBillItemRepoWithSqlMock() (*InvoiceBillItemRepo, *testutil.MockDB) {
	repo := &InvoiceBillItemRepo{}
	return repo, testutil.NewMockDB()
}

func TestInvoiceBillItemRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.InvoiceBillItem{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := InvoiceBillItemRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - create invoice bill item record failed", func(t *testing.T) {
		repo, mockDB := InvoiceBillItemRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert InvoiceBillItemRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - no rows affected after creating bill item record record", func(t *testing.T) {
		repo, mockDB := InvoiceBillItemRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert InvoiceBillItemRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceBillItemRepo_FindAllByInvoiceID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case - one invoice bill item record", func(t *testing.T) {
		repo, mockDB := InvoiceBillItemRepoWithSqlMock()

		e := &entities.InvoiceBillItem{}
		_ = e.InvoiceBillItemID.Set(mock.AnythingOfType("string"))
		_ = e.InvoiceID.Set(mock.AnythingOfType("string"))
		_ = e.BillItemSequenceNumber.Set(mock.AnythingOfType("string"))
		_ = e.PastBillingStatus.Set(mock.AnythingOfType("string"))
		_ = e.CreatedAt.Set(mock.AnythingOfType("time"))

		fields, fieldMap := e.FieldMap()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			fieldMap,
		})

		invoiceBillItems, err := repo.FindAllByInvoiceID(ctx, mockDB.DB, e.InvoiceID.String)

		assert.Nil(t, err)
		assert.Equal(t, invoiceBillItems.ToArray(), []*entities.InvoiceBillItem{e})

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("happy case - multiple invoice bill item records", func(t *testing.T) {
		repo, mockDB := InvoiceBillItemRepoWithSqlMock()

		e := &entities.InvoiceBillItem{}
		_ = e.InvoiceBillItemID.Set(mock.AnythingOfType("string"))
		_ = e.InvoiceID.Set(mock.AnythingOfType("string"))
		_ = e.BillItemSequenceNumber.Set(mock.AnythingOfType("string"))
		_ = e.PastBillingStatus.Set(mock.AnythingOfType("string"))
		_ = e.CreatedAt.Set(mock.AnythingOfType("time"))

		fields, fieldMap := e.FieldMap()

		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			fieldMap,
			fieldMap,
		})

		invoiceBillItems, err := repo.FindAllByInvoiceID(ctx, mockDB.DB, e.InvoiceID.String)

		assert.Nil(t, err)
		assert.Equal(t, invoiceBillItems.ToArray(), []*entities.InvoiceBillItem{e, e})

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - retrieve invoice bill item record failed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		mockDB.MockQueryArgs(t, pgx.ErrTxClosed, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		invoiceBillItems, err := repo.RetrieveInvoiceByInvoiceID(ctx, mockDB.DB, "123")

		assert.NotNil(t, err)
		assert.Nil(t, invoiceBillItems)
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

		invoiceBillItems, err := repo.RetrieveInvoiceByInvoiceID(ctx, mockDB.DB, "123")

		assert.NotNil(t, err)
		assert.Nil(t, invoiceBillItems)
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
