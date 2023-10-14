package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BillingRatioRepoWithSqlMock() (*BillingRatioRepo, *testutil.MockDB) {
	billingRatioRepo := &BillingRatioRepo{}
	return billingRatioRepo, testutil.NewMockDB()
}

func TestBillingRatioRepo_GetFirstRatioByBillingSchedulePeriodIDAndFromTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

	const periodID string = "1"
	var startTime = time.Now()
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			periodID,
			startTime,
			startTime,
		)
		entity := &entities.BillingRatio{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		billingRatio, err := billingRatioRepoWithSqlMock.GetFirstRatioByBillingSchedulePeriodIDAndFromTime(ctx, mockDB.DB, periodID, startTime)
		assert.Nil(t, err)
		assert.NotNil(t, billingRatio)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			periodID,
			startTime,
			startTime,
		)
		entity := &entities.BillingRatio{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		billingRatio, err := billingRatioRepoWithSqlMock.GetFirstRatioByBillingSchedulePeriodIDAndFromTime(ctx, mockDB.DB, periodID, startTime)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, billingRatio)

	})
}

func TestBillingRatioRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.BillingRatio{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("insert billing ratio succeeds", func(t *testing.T) {
		billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := billingRatioRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert billing ratio fails", func(t *testing.T) {
		billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := billingRatioRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BillingRatio: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert billing ratio", func(t *testing.T) {
		billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := billingRatioRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BillingRatio: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBillingRatioRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.BillingRatio{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("update billing ratio fields succeeds", func(t *testing.T) {
		billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := billingRatioRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update billing ratio fields fails", func(t *testing.T) {
		billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := billingRatioRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update BillingRatio: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affect after updating billing ratio fields", func(t *testing.T) {
		billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := billingRatioRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update BillingRatio: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBillingRatioRepo_GetNextRatioByBillingSchedulePeriodIDAndPrevious(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	billingRatioRepoWithSqlMock, mockDB := BillingRatioRepoWithSqlMock()

	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := entities.BillingRatio{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		billingRatio, err := billingRatioRepoWithSqlMock.GetNextRatioByBillingSchedulePeriodIDAndPrevious(ctx, mockDB.DB, entities.BillingRatio{})
		assert.Nil(t, err)
		assert.NotNil(t, billingRatio)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := entities.BillingRatio{
			BillingSchedulePeriodID: pgtype.Text{
				String: "1",
				Status: pgtype.Present,
			},
		}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		billingRatio, err := billingRatioRepoWithSqlMock.GetNextRatioByBillingSchedulePeriodIDAndPrevious(ctx, mockDB.DB, entity)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, billingRatio)

	})
}
