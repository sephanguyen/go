package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/invoicemgmt/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"
	invoice_pb "github.com/manabie-com/backend/pkg/manabuf/invoicemgmt/v1"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func PaymentRepoWithSqlMock() (*PaymentRepo, *testutil.MockDB) {
	repo := &PaymentRepo{}
	return repo, testutil.NewMockDB()
}

func TestPaymentRepo_GetLatestPaymentDueDateByInvoiceID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()
	mockE := &entities.Payment{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("get latest payment due date record failed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		latestRecord, err := repo.GetLatestPaymentDueDateByInvoiceID(ctx, mockDB.DB, "12")
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		latestRecord, err := repo.GetLatestPaymentDueDateByInvoiceID(ctx, mockDB.DB, "42")
		assert.Nil(t, err)
		assert.Equal(t, mockE, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("No rows affected when getting latest payment due date record", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := repo.GetLatestPaymentDueDateByInvoiceID(ctx, mockDB.DB, "55")

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPaymentRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Payment{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - create payment record failed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert PaymentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test - no rows affected after creating payment record", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert PaymentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPaymentRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Payment{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - Update", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update failed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update PaymentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update PaymentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPaymentRepo_UpdateWithFields(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Payment{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - UpdateWithFields", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - UpdateWithFields failed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields PaymentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateWithFields(ctx, mockDB.DB, mockE, []string{"field"})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err updateWithFields PaymentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPaymentRepo_FindByPaymentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()
	mockE := &entities.Payment{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.FindByPaymentID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, mockE, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		e, err := repo.FindByPaymentID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		e, err := repo.FindByPaymentID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPaymentRepo_FindByPaymentSequenceNumber(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()
	mockE := &entities.Payment{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.FindByPaymentSequenceNumber(ctx, mockDB.DB, 0)
		assert.Nil(t, err)
		assert.Equal(t, mockE, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		e, err := repo.FindByPaymentSequenceNumber(ctx, mockDB.DB, 0)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		e, err := repo.FindByPaymentSequenceNumber(ctx, mockDB.DB, 0)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentPaymentDetailRepo_FindPaymentInvoiceByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()

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

		record, err := repo.FindPaymentInvoiceByIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindPaymentInvoiceByIDs(ctx, mockDB.DB, ids)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindPaymentInvoiceByIDs(ctx, mockDB.DB, ids)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPaymentRepo_UpdateIsExportedByPaymentRequestFileID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Payment{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - Update", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByPaymentRequestFileID(ctx, mockDB.DB, mock.Anything, true)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update failed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateIsExportedByPaymentRequestFileID(ctx, mockDB.DB, mock.Anything, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID PaymentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByPaymentRequestFileID(ctx, mockDB.DB, mock.Anything, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentRequestFileID PaymentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPaymentRepo_UpdateIsExportedByPaymentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Payment{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - Update", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByPaymentIDs(ctx, mockDB.DB, []string{"23"}, true)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update failed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateIsExportedByPaymentIDs(ctx, mockDB.DB, []string{"24"}, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentIDs PaymentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateIsExportedByPaymentIDs(ctx, mockDB.DB, []string{"26"}, true)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateIsExportedByPaymentIDs PaymentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPaymentRepo_UpdateStatusAndAmountByPaymentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.Payment{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - Update", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateStatusAndAmountByPaymentIDs(ctx, mockDB.DB, []string{"23"}, mock.Anything, 0)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update failed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.UpdateStatusAndAmountByPaymentIDs(ctx, mockDB.DB, []string{"24"}, mock.Anything, 0)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateStatusAndAmountByPaymentIDs PaymentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - Update no rows affected", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.UpdateStatusAndAmountByPaymentIDs(ctx, mockDB.DB, []string{"26"}, mock.Anything, 0)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err UpdateStatusAndAmountByPaymentIDs PaymentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPaymentRepo_CountOtherPaymentsByBulkPaymentIDNotInStatus(t *testing.T) {
	t.Parallel()

	mockDB := testutil.NewMockDB()

	repo := &PaymentRepo{}

	ctx := context.Background()

	totalCount := 1

	e := &entities.Payment{}
	_ = e.PaymentID.Set(database.Text("12"))
	_ = e.InvoiceID.Set(database.Text("122"))
	_ = e.PaymentStatus.Set(invoice_pb.BulkPaymentStatus_BULK_PAYMENT_PENDING.String())
	_ = e.BulkPaymentID.Set(database.Text("33"))

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(nil, []string{"total_count"}, []interface{}{&totalCount})

		res, err := repo.CountOtherPaymentsByBulkPaymentIDNotInStatus(ctx, mockDB.DB, e.BulkPaymentID.String, "23", invoice_pb.PaymentStatus_PAYMENT_FAILED.String())

		assert.Nil(t, err)
		assert.Equal(t, totalCount, res)
	})

	t.Run("rows scan field error", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockDB.MockRowScanFields(pgx.ErrNoRows, []string{"total_count"}, []interface{}{&totalCount})

		res, err := repo.CountOtherPaymentsByBulkPaymentIDNotInStatus(ctx, mockDB.DB, e.BulkPaymentID.String, "23", invoice_pb.PaymentStatus_PAYMENT_FAILED.String())

		assert.Equal(t, 0, res)
		assert.ErrorIs(t, err, pgx.ErrNoRows)
	})
}

func TestPaymentRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.Payment{
				{PaymentID: database.Text("existing-payment-id")},
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
			req: []*entities.Payment{
				{PaymentID: database.Text("existing-payment-id")},
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
		err := repo.CreateMultiple(ctx, mockDB.DB, testCase.req.([]*entities.Payment))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestPaymentRepo_FindByPaymentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()
	payment := &entities.Payment{}
	_, fieldMap := payment.FieldMap()

	scanFields := []interface{}{}
	for range fieldMap {
		scanFields = append(scanFields, mock.Anything)
	}

	rows := mockDB.Rows

	paymentIDs := []string{"1", "2"}

	t.Run("happy case", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Err").Once().Return(nil)

		record, err := repo.FindByPaymentIDs(ctx, mockDB.DB, paymentIDs)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindByPaymentIDs(ctx, mockDB.DB, paymentIDs)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindByPaymentIDs(ctx, mockDB.DB, paymentIDs)

		assert.Equal(t, "Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPaymentRepo_PaymentSeqNumberLockAdvisory(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.AnythingOfType("*bool")).Return(nil)

		lockAcquired, err := repo.PaymentSeqNumberLockAdvisory(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotNil(t, lockAcquired)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("tx closed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything)
		mockDB.Row.On("Scan", mock.AnythingOfType("*bool")).Once().Return(pgx.ErrTxClosed)

		_, err := repo.PaymentSeqNumberLockAdvisory(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		resourcePath := golibs.ResourcePathFromCtx(ctx)
		assert.Equal(t, fmt.Errorf("err PaymentSeqNumberLockAdvisory PaymentRepo: %w - resourcePath: %s", pgx.ErrTxClosed, resourcePath).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPaymentRepo_PaymentSeqNumberUnLockAdvisory(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")})

		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := repo.PaymentSeqNumberUnLockAdvisory(ctx, mockDB.DB)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("tx closed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")})

		mockDB.MockExecArgs(t, cmdTag, pgx.ErrTxClosed, args...)

		err := repo.PaymentSeqNumberUnLockAdvisory(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		resourcePath := golibs.ResourcePathFromCtx(ctx)
		assert.Equal(t, fmt.Errorf("err PaymentSeqNumberLockAdvisory PaymentRepo: %w - resourcePath: %s", pgx.ErrTxClosed, resourcePath).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPaymentRepo_InsertPaymentNumbersTempTable(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()

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
		err := repo.InsertPaymentNumbersTempTable(ctx, mockDB.DB, []int{1, 2})
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestPaymentRepo_FindPaymentInvoiceUserFromTempTable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := PaymentRepoWithSqlMock()
	invoice := &entities.Invoice{}
	payment := &entities.Payment{}
	userBasicInfo := &entities.UserBasicInfo{}

	fields := []interface{}{}
	_, invoiceFields := invoice.FieldMap()
	_, paymentFields := payment.FieldMap()
	_, userBasicInfoFields := userBasicInfo.FieldMap()

	fields = append(fields, invoiceFields...)
	fields = append(fields, paymentFields...)
	fields = append(fields, userBasicInfoFields...)

	scanFields := []interface{}{}
	for range fields {
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

		record, err := repo.FindPaymentInvoiceUserFromTempTable(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindPaymentInvoiceUserFromTempTable(ctx, mockDB.DB)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindPaymentInvoiceUserFromTempTable(ctx, mockDB.DB)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestPaymentRepo_UpdateMultipleWithFields(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	payments := []*entities.Payment{
		{
			PaymentID: database.Text("payment-1"),
		}, {
			PaymentID: database.Text("payment-2"),
		},
	}

	fields := []string{"payment_date", "amount", "payment_status"}
	args := []interface{}{mock.Anything, mock.Anything}
	for i := 0; i < len(fields)*len(payments)+len(payments); i++ {
		args = append(args, mock.Anything)
	}

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))

		mockDB.MockExecArgs(t, cmdTag, nil, args...)

		err := repo.UpdateMultipleWithFields(ctx, mockDB.DB, payments, fields)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("tx closed", func(t *testing.T) {
		repo, mockDB := PaymentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))

		mockDB.MockExecArgs(t, cmdTag, pgx.ErrTxClosed, args...)

		err := repo.UpdateMultipleWithFields(ctx, mockDB.DB, payments, fields)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, pgx.ErrTxClosed.Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
