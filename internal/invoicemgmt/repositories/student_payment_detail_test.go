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

func StudentPaymentDetailWithSqlMock() (*StudentPaymentDetailRepo, *testutil.MockDB) {
	repo := &StudentPaymentDetailRepo{}
	return repo, testutil.NewMockDB()
}

func TestStudentPaymentDetailRepo_FindByStudentByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentID := "test-id"
	mockE := &entities.StudentPaymentDetail{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentPaymentDetailWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByStudentID(ctx, mockDB.DB, studentID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("find student payment detail failed", func(t *testing.T) {
		repo, mockDB := StudentPaymentDetailWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByStudentID(ctx, mockDB.DB, studentID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByStudentID StudentPaymentDetailRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("No rows after student payment detail find by student id", func(t *testing.T) {
		repo, mockDB := StudentPaymentDetailWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByStudentID(ctx, mockDB.DB, studentID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByStudentID StudentPaymentDetailRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentPaymentDetailRepo_FindStudentBillingByStudentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentPaymentDetailWithSqlMock()

	scanFields := make([]interface{}, 12)
	for i := 0; i < 12; i++ {
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

		record, err := repo.FindStudentBillingByStudentIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindStudentBillingByStudentIDs(ctx, mockDB.DB, ids)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindStudentBillingByStudentIDs(ctx, mockDB.DB, ids)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentPaymentDetailRepo_FindStudentBankDetailsByStudentIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentPaymentDetailWithSqlMock()

	scanFields := make([]interface{}, 12)
	for i := 0; i < 12; i++ {
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

		record, err := repo.FindStudentBankDetailsByStudentIDs(ctx, mockDB.DB, ids)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindStudentBankDetailsByStudentIDs(ctx, mockDB.DB, ids)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindStudentBankDetailsByStudentIDs(ctx, mockDB.DB, ids)

		assert.Equal(t, "row.Scan: Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentPaymentDetailRepo_FindByID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	studentPaymentDetailID := "example-student-payment-detail-id"
	mockE := &entities.StudentPaymentDetail{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := StudentPaymentDetailWithSqlMock()
		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		record, err := repo.FindByID(ctx, mockDB.DB, studentPaymentDetailID)
		assert.Nil(t, err)
		assert.Equal(t, mockE, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("find billing address failed", func(t *testing.T) {
		repo, mockDB := StudentPaymentDetailWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		record, err := repo.FindByID(ctx, mockDB.DB, studentPaymentDetailID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err FindByID StudentPaymentDetailRepo: err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("No rows returned when finding billing address by id", func(t *testing.T) {
		repo, mockDB := StudentPaymentDetailWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		record, err := repo.FindByID(ctx, mockDB.DB, studentPaymentDetailID)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err FindByID StudentPaymentDetailRepo: err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestStudentPaymentDetailRepo_Upsert(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentPaymentDetailWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.StudentPaymentDetail{
				{StudentPaymentDetailID: database.Text("existing-student-payment-detail-id")},
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
			req: []*entities.StudentPaymentDetail{
				{StudentPaymentDetailID: database.Text("existing-student-payment-detail-id")},
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
		err := repo.Upsert(ctx, mockDB.DB, testCase.req.([]*entities.StudentPaymentDetail)...)
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}

func TestStudentPaymentDetailRepo_SoftDelete(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentPaymentDetailWithSqlMock()
	studentPaymentDetailIDs := []string{"example-student-payment-detail-id-1", "example-student-payment-detail-id-2"}

	t.Run("err update", func(t *testing.T) {
		// move primaryField to the last
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(studentPaymentDetailIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), puddle.ErrClosedPool, args...)

		err := repo.SoftDelete(ctx, mockDB.DB, studentPaymentDetailIDs...)
		assert.True(t, errors.Is(err, puddle.ErrClosedPool))
	})

	t.Run("success", func(t *testing.T) {
		args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, database.TextArray(studentPaymentDetailIDs))
		mockDB.MockExecArgs(t, pgconn.CommandTag("1"), nil, args...)

		err := repo.SoftDelete(ctx, mockDB.DB, studentPaymentDetailIDs...)
		assert.Nil(t, err)

		// move primaryField to the last
		mockDB.RawStmt.AssertUpdatedFields(t, "deleted_at")
		mockDB.RawStmt.AssertWhereConditions(t, map[string]*testutil.CheckWhereClauseOpt{
			"student_payment_detail_id": {HasNullTest: false, EqualExpr: &testutil.EqualExpr{IndexArg: 1}},
			"deleted_at":                {HasNullTest: true},
		})
	})
}

func TestStudentPaymentDetailRepo_FindFromInvoiceIDTempTable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := StudentPaymentDetailWithSqlMock()
	studentPaymentDetail := &entities.StudentPaymentDetail{}
	_, fieldMap := studentPaymentDetail.FieldMap()

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

		record, err := repo.FindFromInvoiceIDTempTable(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotEmpty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("db.Query returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(nil, pgx.ErrTxClosed)

		record, err := repo.FindFromInvoiceIDTempTable(ctx, mockDB.DB)

		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("Scan returns error", func(t *testing.T) {

		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Close").Once().Return(nil)
		rows.On("Next").Once().Return(true)
		rows.On("Scan", scanFields...).Once().Return(errors.New("Scan error"))

		record, err := repo.FindFromInvoiceIDTempTable(ctx, mockDB.DB)

		assert.Equal(t, "Scan error", err.Error())
		assert.Empty(t, record)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}
