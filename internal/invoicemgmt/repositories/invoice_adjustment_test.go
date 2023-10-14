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

func InvoiceAdjustmentRepoWithSqlMock() (InvoiceAdjustmentRepo, *testutil.MockDB) {
	mockDB := testutil.NewMockDB()
	repo := InvoiceAdjustmentRepo{}

	return repo, mockDB
}

func TestInvoiceAdjustmentRepo_UpsertMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()
	invoiceAdjustmentIDS := []string{"1", "2"}
	invoiceIDS := []string{"invoice-1", "invoice-2"}
	invoiceAdjustmentEntities := []*entities.InvoiceAdjustment{
		{
			InvoiceAdjustmentID: database.Text(invoiceAdjustmentIDS[0]),
			InvoiceID:           database.Text(invoiceIDS[0]),
		},
		{
			InvoiceAdjustmentID: database.Text(invoiceAdjustmentIDS[1]),
			InvoiceID:           database.Text(invoiceIDS[1]),
		},
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectErr    error
		expectedResp interface{}
		setup        func()
	}{
		{
			name:         "err upsert multiple invoice adjustment success",
			request:      invoiceAdjustmentEntities,
			expectErr:    nil,
			expectedResp: invoiceAdjustmentEntities,
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "err upsert multiple invoice adjustment failed send batch error",
			request:      invoiceAdjustmentEntities,
			expectErr:    puddle.ErrClosedPool,
			expectedResp: []*entities.InvoiceAdjustment(nil),
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`1`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Once().Return(cmdTag, puddle.ErrClosedPool)
				batchResults.On("Close").Once().Return(nil)
			},
		},
		{
			name:         "err upsert multiple invoice adjustment failed row affected different error",
			request:      invoiceAdjustmentEntities,
			expectErr:    fmt.Errorf("err upsert multiple invoice adjustment: %d RowsAffected", 0),
			expectedResp: []*entities.InvoiceAdjustment(nil),
			setup: func() {
				batchResults := &mock_database.BatchResults{}
				cmdTag := pgconn.CommandTag([]byte(`0`))
				mockDB.DB.On("SendBatch", mock.Anything, mock.Anything).Once().Return(batchResults)
				batchResults.On("Exec").Return(cmdTag, nil)
				batchResults.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testcase := range testCases {
		testName := fmt.Sprintf("TestCase: %s", testcase.name)
		t.Run(testName, func(t *testing.T) {
			testcase.setup()
			err := repo.UpsertMultiple(ctx, mockDB.DB, testcase.request.([]*entities.InvoiceAdjustment))
			assert.Equal(t, testcase.expectErr, err)
		})
	}
}

func TestInvoiceAdjustment_SoftDeleteByIDs(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	singleInvoiceAdjustmentID := database.TextArray([]string{"test-id"})
	multiInvoiceAdjustmentID := database.TextArray([]string{"test-id", "test-id2"})

	mockE := &entities.InvoiceAdjustment{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case single", func(t *testing.T) {
		repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, singleInvoiceAdjustmentID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("happy case multiple", func(t *testing.T) {
		repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`2`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, multiInvoiceAdjustmentID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("soft delete invoice adjustment record fail", func(t *testing.T) {
		repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, singleInvoiceAdjustmentID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete InvoiceAdjustmentRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after soft deleting invoice adjustment record", func(t *testing.T) {
		repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, singleInvoiceAdjustmentID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete InvoiceAdjustmentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after soft deleting invoice adjustment records", func(t *testing.T) {
		repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.SoftDeleteByIDs(ctx, mockDB.DB, multiInvoiceAdjustmentID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete InvoiceAdjustmentRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceAdjustmentRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()
	mockE := &entities.InvoiceAdjustment{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.FindByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, mockE, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - tx closed", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		e, err := repo.FindByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		e, err := repo.FindByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestInvoiceAdjustmentRepo_FindByInvoiceIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	invoiceIDs := []string{"1", "2"}

	repo, mockDB := InvoiceAdjustmentRepoWithSqlMock()
	invoiceAdjustment := &entities.InvoiceAdjustment{}
	_, fieldMap := invoiceAdjustment.FieldMap()

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

		record, err := repo.FindByInvoiceIDs(ctx, mockDB.DB, invoiceIDs)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindByInvoiceIDs(ctx, mockDB.DB, invoiceIDs)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindByInvoiceIDs(ctx, mockDB.DB, invoiceIDs)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
