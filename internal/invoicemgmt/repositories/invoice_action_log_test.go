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

func InvoiceActionLogRepoWithSqlMock() (*InvoiceActionLogRepo, *testutil.MockDB) {
	repo := &InvoiceActionLogRepo{}
	return repo, testutil.NewMockDB()
}
func genSliceMock(n int) []interface{} {
	result := []interface{}{}
	for i := 0; i < n; i++ {
		result = append(result, mock.Anything)
	}
	return result
}

func TestInvoiceActionLogRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.InvoiceActionLog{}
	_, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := InvoiceActionLogRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert invoice_action_log failed", func(t *testing.T) {
		repo, mockDB := InvoiceActionLogRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`1`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, pgx.ErrTxClosed)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert InvoiceActionLogRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("No rows affected after invoice_action_log inserted", func(t *testing.T) {
		repo, mockDB := InvoiceActionLogRepoWithSqlMock()

		cmdTag := pgconn.CommandTag([]byte(`0`))
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockE)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert InvoiceActionLogRepo: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestInvoiceActionLogRepo_GetLatestRecordByInvoiceID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockE := &entities.InvoiceActionLog{}
	fields, fieldMap := mockE.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy test case - GetLatestRecordByInvoiceID", func(t *testing.T) {
		repo, mockDB := InvoiceActionLogRepoWithSqlMock()

		mockDB.MockQueryArgs(t, nil, args...)
		mockDB.MockScanArray(nil, fields, [][]interface{}{fieldMap})

		e, err := repo.GetLatestRecordByInvoiceID(ctx, mockDB.DB, "1")

		assert.Nil(t, err)
		assert.NotNil(t, e)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})

	t.Run("negative test case - GetLatestRecordByInvoiceID failed", func(t *testing.T) {
		repo, mockDB := InvoiceActionLogRepoWithSqlMock()

		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrTxClosed)
		latestRecord, err := repo.GetLatestRecordByInvoiceID(ctx, mockDB.DB, "1")
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))

		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrTxClosed).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("negative test case - GetLatestRecordByInvoiceID no rows affected", func(t *testing.T) {
		repo, mockDB := InvoiceActionLogRepoWithSqlMock()

		mockDB.DB.On("Query", args...).Once().Return(nil, pgx.ErrNoRows)
		latestRecord, err := repo.GetLatestRecordByInvoiceID(ctx, mockDB.DB, "1")

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Equal(t, fmt.Errorf("err db.Query: %w", pgx.ErrNoRows).Error(), err.Error())
		assert.Nil(t, latestRecord)

		mock.AssertExpectationsForObjects(t, mockDB.DB)
	})
}

func TestInvoiceActionLogRepo_CreateMultiple(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	repo, mockDB := InvoiceActionLogRepoWithSqlMock()

	testCases := []TestCase{
		{
			name: "happy case",
			req: []*entities.InvoiceActionLog{
				{InvoiceActionID: database.Text("existing-id")},
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
			req: []*entities.InvoiceActionLog{
				{InvoiceActionID: database.Text("existing-id")},
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
		err := repo.CreateMultiple(ctx, mockDB.DB, testCase.req.([]*entities.InvoiceActionLog))
		if testCase.expectedErr != nil {
			assert.Equal(t, testCase.expectedErr.Error(), err.Error())
		} else {
			assert.Equal(t, testCase.expectedErr, err)
		}
	}
}
