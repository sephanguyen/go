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

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func InvoiceRepoWithSqlMock() (*InvoiceRepo, *testutil.MockDB) {
	repo := &InvoiceRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentInvoiceRecordsRepo_RetrieveRecordsByStudentID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := pgtype.Text{}
	_ = studentID.Set(uuid.NewString())
	selectFields := []string{"invoice_id", "status", "total"}
	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(selectFields))...)
	repo, mockDB := InvoiceRepoWithSqlMock()
	limit := pgtype.Int8{Int: 100}
	offset := pgtype.Int8{Int: 0}
	t.Run("failed to select invoice records", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		invoiceRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, studentID.String, limit, offset)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err retrieve records InvoiceRepo: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, invoiceRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})
	t.Run("No rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		invoiceRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, studentID.String, limit, offset)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err retrieve records InvoiceRepo: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, invoiceRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("success retrieving single invoice records", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		e := &entities.Invoice{}
		_ = e.InvoiceID.Set("12")
		_ = e.Status.Set("FAILED")
		_ = e.Total.Set(1822.00)
		value := database.GetScanFields(e, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			value,
		})

		invoiceRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, studentID.String, limit, offset)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.Invoice{
			{
				InvoiceID: database.Text("12"),
				Status:    database.Text("FAILED"),
				Total:     database.Numeric(1822.00),
			},
		}, invoiceRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})

	t.Run("success retrieving multiple invoice records", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)
		var invoices []*entities.Invoice
		e := &entities.Invoice{}
		_ = e.InvoiceID.Set("14")
		_ = e.Status.Set("VOID")
		_ = e.Total.Set(1500.00)
		invoices = append(invoices, e)
		valueOne := database.GetScanFields(e, selectFields)

		e = &entities.Invoice{}
		_ = e.InvoiceID.Set("15")
		_ = e.Status.Set("PAID")
		_ = e.Total.Set(2500.00)
		invoices = append(invoices, e)

		value := database.GetScanFields(e, selectFields)

		mockDB.MockScanArray(nil, selectFields, [][]interface{}{
			valueOne,
			value,
		})

		invoiceRecords, err := repo.RetrieveRecordsByStudentID(ctx, mockDB.DB, studentID.String, limit, offset)
		assert.Nil(t, err)
		assert.Equal(t, invoices, invoiceRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, selectFields...)
	})
}

func TestInvoiceRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Invoice{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - update invoice record failed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update InvoiceRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - no rows affected after updating invoice record", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceRepo_Create(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Invoice{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("generate invoice failed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Invoice: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})

	t.Run("No rows affected after invoice inserted", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()
		mockDB.DB.On("QueryRow", args...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrNoRows)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Invoice: %w", pgx.ErrNoRows).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Row)
	})
}

func TestInvoiceRepo_UpdateWithFields(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Invoice{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - UpdateWithFields", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - InvoiceRepoWithSqlMock failed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields InvoiceRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceRepo_UpdateIsExportedByPaymentRequestFileID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Invoice{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - Update", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByPaymentRequestFileID(ctx, mockDB.DB, mock.Anything, true)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update failed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateIsExportedByPaymentRequestFileID(ctx, mockDB.DB, mock.Anything, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID InvoiceRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByPaymentRequestFileID(ctx, mockDB.DB, mock.Anything, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceRepo_UpdateIsExportedByInvoiceIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Invoice{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - Update", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByInvoiceIDs(ctx, mockDB.DB, []string{"23"}, true)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update failed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateIsExportedByInvoiceIDs(ctx, mockDB.DB, []string{"24"}, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentIDs InvoiceRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByInvoiceIDs(ctx, mockDB.DB, []string{"26"}, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentIDs InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceRepo_RetrieveInvoiceByInvoiceReferenceID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Invoice{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.Anything}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - RetrieveInvoiceByInvoiceReferenceID", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.RetrieveInvoiceByInvoiceReferenceID(ctx, mockDB.DB, mock.Anything)

		assert.Nil(t, err)
		assert.NotNil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test case - RetrieveInvoiceByInvoiceReferenceID - ErrTxClosed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		latestRecord, err := repo.RetrieveInvoiceByInvoiceReferenceID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - RetrieveInvoiceByInvoiceReferenceID no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := repo.RetrieveInvoiceByInvoiceReferenceID(ctx, mockDB.DB, mock.Anything)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentInvoiceRecordsRepo_RetrievedMigratedInvoices(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	e := &entities.Invoice{}
	_, fieldMap := e.FieldMap()

	scanFields := []interface{}{}
	for range fieldMap {
		scanFields = append(scanFields, mock.Anything)
	}

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.Anything})

	repo, mockDB := InvoiceRepoWithSqlMock()

	t.Run("failed to select invoice records", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		invoiceRecords, err := repo.RetrievedMigratedInvoices(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err InvoiceRepo.RetrievedMigratedInvoices: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, invoiceRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})
	t.Run("No rows affected", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		invoiceRecords, err := repo.RetrievedMigratedInvoices(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err InvoiceRepo.RetrievedMigratedInvoices: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, invoiceRecords)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	rows := mockDB.Rows

	t.Run("success retrieving invoice records", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		invoiceRecords, err := repo.RetrievedMigratedInvoices(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotEmpty(t, invoiceRecords)
	})
}

func TestInvoiceRepo_InsertInvoiceIDsTempTable(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := InvoiceRepoWithSqlMock()

	testCases := []TestCase{
		{
			name:        "happy case",
			expectedErr: nil,
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", mock.Anything, mock.Anything).Return(cmdTag, nil)

				batchResults := &mock_database.BatchResults{}
				cmdTag = pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:        "err when exec batch",
			expectedErr: errors.Wrap(errors.New("err when exec"), "batchResults.Exec"),
			setup: func(ctx context.Context) {
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("Exec", mock.Anything, mock.Anything).Return(cmdTag, nil)

				batchResults := &mock_database.BatchResults{}
				cmdTag = pgconn.CommandTag([]byte(`0`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, errors.New("err when exec"))
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}

	for _, testCase := range testCases {
		testCase.setup(ctx)
		err := repo.InsertInvoiceIDsTempTable(ctx, mockDB.DB, []string{"1", "2"})
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestInvoiceRepo_FindInvoicesFromInvoiceIDTempTable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := InvoiceRepoWithSqlMock()
	invoice := &entities.Invoice{}
	_, fieldMap := invoice.FieldMap()

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

		record, err := repo.FindInvoicesFromInvoiceIDTempTable(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindInvoicesFromInvoiceIDTempTable(ctx, mockDB.DB)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindInvoicesFromInvoiceIDTempTable(ctx, mockDB.DB)

		assert.Equal(t, "Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestInvoiceRepo_UpdateStatusFromInvoiceIDTempTable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")})

	t.Run("happy test case - Update", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateStatusFromInvoiceIDTempTable(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update failed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateStatusFromInvoiceIDTempTable(ctx, mockDB.DB, mock.Anything)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateStatusFromInvoiceIDTempTable InvoiceRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateStatusFromInvoiceIDTempTable(ctx, mockDB.DB, mock.Anything)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateStatusFromInvoiceIDTempTable InvoiceRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceRepo_UpdateMultipleWithFields(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	invoices := []*entities.Invoice{
		{
			InvoiceID: database.Text("payment-1"),
		}, {
			InvoiceID: database.Text("payment-2"),
		},
	}

	fields := []string{"status", "total", "amount_paid"}
	args := []interface{}{mock.Anything, mock.Anything}
	for i := 0; i < len(fields)*len(invoices)+len(invoices); i++ {
		args = append(args, mock.Anything)
	}

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))

		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := repo.UpdateMultipleWithFields(ctx, mockDB.DB, invoices, fields)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("tx closed", func(t *testing.T) {
		repo, mockDB := InvoiceRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))

		mockDB.MockExecArgs(t, cmdTag, pgx.ErrTxClosed, args...)

		err := repo.UpdateMultipleWithFields(ctx, mockDB.DB, invoices, fields)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, pgx.ErrTxClosed.Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestInvoiceRepo_RetrieveInvoiceData(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := InvoiceRepoWithSqlMock()

	scanFields := make([]interface{}, 21)
	for i := 0; i < 21; i++ {
		scanFields[i] = mock.Anything
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.RetrieveInvoiceData(ctx, mockDB.DB, database.Int8(100), database.Int8(0), mock.Anything)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.RetrieveInvoiceData(ctx, mockDB.DB, database.Int8(100), database.Int8(0), mock.Anything)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.RetrieveInvoiceData(ctx, mockDB.DB, database.Int8(100), database.Int8(0), mock.Anything)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestInvoiceRepo_RetrieveInvoiceStatusCount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := InvoiceRepoWithSqlMock()
	scanFields := make([]interface{}, 6)
	for i := 0; i < 6; i++ {
		scanFields[i] = mock.Anything
	}

	rows := mockDB.Rows

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.RetrieveInvoiceStatusCount(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.RetrieveInvoiceStatusCount(ctx, mockDB.DB, mock.Anything)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.RetrieveInvoiceStatusCount(ctx, mockDB.DB, mock.Anything)

		assert.Equal(t, "rows.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
