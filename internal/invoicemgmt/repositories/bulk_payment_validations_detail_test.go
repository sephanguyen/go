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

func BulkPaymentValidationsDetailWithSqlMock() (*BulkPaymentValidationsDetailRepo, *testutil.MockDB) {
	repo := &BulkPaymentValidationsDetailRepo{}
	return repo, testutil.NewMockDB()
}

func TestBulkPaymentValidationsDetailRepo_RetrieveRecordsByBulkPaymentValidationsID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	bulkPaymentValidationsID := database.Text("test-id")
	mockE := &entities.BulkPaymentValidationsDetail{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("failed to select bulk payment validations details records", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		bulkPaymentValidationDetailsRec, err := repo.RetrieveRecordsByBulkPaymentValidationsID(ctx, mockDB.DB, bulkPaymentValidationsID)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Equal(t, fmt.Errorf("err retrieve records BulkPaymentValidationsDetailRepo: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, bulkPaymentValidationDetailsRec)
		mock.AssertExpectationsForObjects(t, mockDB.DB)

	})

	t.Run("No rows affected", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		bulkPaymentValidationDetailsRec, err := repo.RetrieveRecordsByBulkPaymentValidationsID(ctx, mockDB.DB, bulkPaymentValidationsID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err retrieve records BulkPaymentValidationsDetailRepo: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, bulkPaymentValidationDetailsRec)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("success with select all", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			mock.Anything, mock.Anything)

		e := &entities.BulkPaymentValidationsDetail{}
		_ = e.BulkPaymentValidationsID.Set("test-detail-id")
		fields, _ := mockE.FieldMap()
		value := database.GetScanFields(e, fields)

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			value,
		})
		bulkPaymentValidationDetailsRec, err := repo.RetrieveRecordsByBulkPaymentValidationsID(ctx, mockDB.DB, bulkPaymentValidationsID)
		assert.Nil(t, err)
		assert.Equal(t, []*entities.BulkPaymentValidationsDetail{
			{BulkPaymentValidationsID: database.Text("test-detail-id")},
		}, bulkPaymentValidationDetailsRec)
		mock.AssertExpectationsForObjects(t, mockDB.DB)
		mockDB.RawStmt.AssertSelectedFields(t, fields...)
	})
}

func TestBulkPaymentValidationsDetailRepo_Create(t *testing.T) {

	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.BulkPaymentValidationsDetail{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert failed", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BulkPaymentValidationsDetailRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("No rows affected after inserted", func(t *testing.T) {
		repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		_, err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BulkPaymentValidationsDetailRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBulkPaymentValidationsRepo_FindByPaymentID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()
	mockE := &entities.BulkPaymentValidationsDetail{}
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
		e, err := repo.FindByPaymentID(ctx, mockDB.DB, "")
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test - no rows", func(t *testing.T) {
		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		e, err := repo.FindByPaymentID(ctx, mockDB.DB, "")
		assert.True(t, errors.Is(err, pgx.ErrNoRows))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestBulkPaymentValidationsRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := BulkPaymentValidationsDetailWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.BulkPaymentValidationsDetail{
				{BulkPaymentValidationsDetailID: database.Text("existing-id")},
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
			req: []*entities.BulkPaymentValidationsDetail{
				{BulkPaymentValidationsDetailID: database.Text("existing-id")},
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
		err := repo.CreateMultiple(ctx, mockDB.DB, testCase.req.([]*entities.BulkPaymentValidationsDetail))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
