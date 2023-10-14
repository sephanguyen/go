package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BillItemRepoWithSqlMock() (*BillItemRepo, *testutil.MockDB) {
	repo := &BillItemRepo{}
	return repo, testutil.NewMockDB()
}

func TestBillItemRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	billItemID := database.Int8(1)
	repo, mockDB := BillItemRepoWithSqlMock()
	mockE := &entities.BillItem{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.Anything}, genSliceMock(len(fieldMap))...)

	t.Run("select bill_item failed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByID(ctx, mockDB.DB, int32(billItemID.Int))
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("db.QueryRowEx 1: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByID(ctx, mockDB.DB, int32(billItemID.Int))
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("No rows after bill_item select", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := repo.FindByID(ctx, mockDB.DB, int32(billItemID.Int))

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("db.QueryRowEx 1: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBillItemRepo_FindByStatuses(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	billItemStatuses := []string{"BILLED", "PENDING"}

	repo, mockDB := BillItemRepoWithSqlMock()
	billItem := &entities.BillItem{}
	_, fieldMap := billItem.FieldMap()

	scanFields := []interface{}{}
	for range fieldMap {
		scanFields = append(scanFields, mock.Anything)
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.FindByStatuses(ctx, mockDB.DB, billItemStatuses)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindByStatuses(ctx, mockDB.DB, billItemStatuses)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindByStatuses(ctx, mockDB.DB, billItemStatuses)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBillItemRepo_FindByOrderID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BillItemRepoWithSqlMock()
	billItem := &entities.BillItem{}
	_, fieldMap := billItem.FieldMap()

	scanFields := []interface{}{}
	for range fieldMap {
		scanFields = append(scanFields, mock.Anything)
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.FindByOrderID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindByOrderID(ctx, mockDB.DB, mock.Anything)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindByOrderID(ctx, mockDB.DB, mock.Anything)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBillItemRepo_RetrieveBillItemsByInvoiceReferenceNum(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BillItemRepoWithSqlMock()

	scanFields := []interface{}{}
	for i := 0; i < 8; i++ {
		scanFields = append(scanFields, mock.Anything)
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.RetrieveBillItemsByInvoiceReferenceNum(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.RetrieveBillItemsByInvoiceReferenceNum(ctx, mockDB.DB, mock.Anything)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.RetrieveBillItemsByInvoiceReferenceNum(ctx, mockDB.DB, mock.Anything)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestCourseStudyPlan_GetBillItemTotalByStudentAndReference(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.BillItem{}
	_, fieldMap := mockE.FieldMap()

	studentReferenceID := "student-test"
	invoiceReferenceID := "invoice-reference"

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BillItemRepoWithSqlMock()

		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		_, err := repo.GetBillItemTotalByStudentAndReference(ctx, mockDB.DB, studentReferenceID, invoiceReferenceID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("find bill item with invoice reference and student", func(t *testing.T) {
		repo, mockDB := BillItemRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

		_, err := repo.GetBillItemTotalByStudentAndReference(ctx, mockDB.DB, studentReferenceID, invoiceReferenceID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err BillItemRepo GetBillItemTotalByStudentAndReference: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("No rows affected GetBillItemTotalByStudentAndReference", func(t *testing.T) {
		repo, mockDB := BillItemRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		_, err := repo.GetBillItemTotalByStudentAndReference(ctx, mockDB.DB, studentReferenceID, invoiceReferenceID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err BillItemRepo GetBillItemTotalByStudentAndReference: %w", pgx.ErrNoRows).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
}

func TestBillItemRepo_FindInvoiceBillItemMapByInvoiceIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	invoiceIDs := []string{"1", "2"}

	repo, mockDB := BillItemRepoWithSqlMock()

	scanFields := []interface{}{}
	for i := 0; i < 6; i++ {
		scanFields = append(scanFields, mock.Anything)
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.FindInvoiceBillItemMapByInvoiceIDs(ctx, mockDB.DB, invoiceIDs)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindInvoiceBillItemMapByInvoiceIDs(ctx, mockDB.DB, invoiceIDs)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindInvoiceBillItemMapByInvoiceIDs(ctx, mockDB.DB, invoiceIDs)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
