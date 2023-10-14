package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func FeeRepoWithSqlMock() (*FeeRepo, *testutil.MockDB, *mock_database.Tx) {
	feeRepo := &FeeRepo{}
	return feeRepo, testutil.NewMockDB(), &mock_database.Tx{}
}

func TestFeeRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Fee{}
	_, fieldProductMap := mockEntities.Product.FieldMap()
	_, fieldFeeMap := mockEntities.FieldMap()

	argsFee := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldFeeMap))...)
	argsQueryRow := append([]interface{}{mock.Anything}, genSliceMock(len(fieldProductMap))...)

	t.Run("happy case", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		tx.On("Exec", argsFee...).Once().Return(constant.SuccessCommandTag, nil)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Create(ctx, tx, mockEntities)
		})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert product fail", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Create(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Product: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert fee fail", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsFee...).Once().Return(constant.SuccessCommandTag, pgx.ErrTxClosed)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Create(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Fee: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert fee", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsFee...).Return(constant.FailCommandTag, nil)
		tx.On("QueryRow", argsQueryRow...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Create(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Fee: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestFeeRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Fee{}
	_, fieldProductMap := mockEntities.Product.FieldMap()
	_, fieldFeeMap := mockEntities.FieldMap()

	argsProduct := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldProductMap))...)
	argsFee := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldFeeMap))...)

	t.Run("happy case", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Commit", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(constant.SuccessCommandTag, nil)
		tx.On("Exec", argsFee...).Return(constant.SuccessCommandTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update product fail", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(constant.SuccessCommandTag, pgx.ErrTxClosed)
		tx.On("Exec", argsFee...).Return(constant.SuccessCommandTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Product: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("update fee fail", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(constant.SuccessCommandTag, nil)
		tx.On("Exec", argsFee...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Fee: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update product", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		cmdProductTag := constant.FailCommandTag
		cmdFeeTag := constant.SuccessCommandTag
		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(cmdProductTag, nil)
		tx.On("Exec", argsFee...).Return(cmdFeeTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Product: %d RowsAffected", cmdProductTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after update fee", func(t *testing.T) {
		feeRepoWithSqlMock, mockDB, tx := FeeRepoWithSqlMock()

		cmdProductTag := constant.SuccessCommandTag
		cmdFeeTag := constant.FailCommandTag
		mockDB.DB.On("Begin", mock.Anything, mock.Anything).Return(tx, nil)
		tx.On("Rollback", mock.Anything).Return(nil)
		tx.On("Exec", argsProduct...).Once().Return(cmdProductTag, nil)
		tx.On("Exec", argsFee...).Return(cmdFeeTag, nil)

		err := database.ExecInTx(ctx, mockDB.DB, func(ctx context.Context, tx pgx.Tx) error {
			return feeRepoWithSqlMock.Update(ctx, tx, mockEntities)
		})
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Fee: %d RowsAffected", cmdFeeTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestFeeRepo_GetAll(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		feeRepoWithSqlMock *FeeRepo
		mockDB             *testutil.MockDB
	)

	testcases := []utils.TestCase{
		{
			Name:        constant.FailCaseErrorRow,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(1).Return(true)

				fee := &entities.Fee{}
				fields, _ := fee.FieldMap()
				scanFields := database.GetScanFields(fee, fields)
				rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
				//rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
		{
			Name:        constant.HappyCase,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				rows := mockDB.Rows
				mockDB.DB.On("Query", mock.Anything, mock.Anything).Once().Return(rows, nil)
				rows.On("Next").Times(1).Return(true)

				entity := &entities.Fee{}
				fields, _ := entity.FieldMap()
				scanFields := database.GetScanFields(entity, fields)
				rows.On("Scan", scanFields...).Once().Return(nil)
				rows.On("Next").Once().Return(false)
				rows.On("Close").Once().Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			feeRepoWithSqlMock, mockDB, _ = FeeRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)
			fees, err := feeRepoWithSqlMock.GetAll(ctx, mockDB.DB)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NotNil(t, fees)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestFeeRepo_GetFeeByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	feeRepoWithSqlMock, mockDB, _ := FeeRepoWithSqlMock()
	feeID := "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, feeID)
		entity := entities.Fee{}
		fields, values := entity.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		fee, err := feeRepoWithSqlMock.GetFeeByID(ctx, mockDB.DB, feeID)
		assert.Nil(t, err)
		assert.NotNil(t, fee)
	})

}
