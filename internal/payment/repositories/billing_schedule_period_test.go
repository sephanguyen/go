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

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func BillingSchedulePeriodRepoWithSqlMock() (*BillingSchedulePeriodRepo, *testutil.MockDB) {
	billingSchedulePeriodRepo := &BillingSchedulePeriodRepo{}
	return billingSchedulePeriodRepo, testutil.NewMockDB()
}

func TestBillingSchedulePeriodRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var periodID string = "1"
	billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			periodID,
		)
		entity := &entities.BillingSchedulePeriod{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		period, err := billingSchedulePeriodRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, periodID)
		assert.Nil(t, err)
		assert.NotNil(t, period)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			periodID,
		)
		e := &entities.BillingSchedulePeriod{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := billingSchedulePeriodRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, periodID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)
	})
}

func TestBillingSchedulePeriodRepo_GetPeriodIDsByScheduleIDAndStartTimeForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var scheduleID string = "1"
	billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		entity := &entities.BillingSchedulePeriod{}
		fields, values := entity.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billingSchedulePeriod, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodIDsByScheduleIDAndStartTimeForUpdate(ctx, mockDB.DB, scheduleID, time.Now())
		assert.Nil(t, err)
		assert.NotNil(t, billingSchedulePeriod)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, scheduleID, mock.Anything, mock.Anything)
		e := &entities.BillingSchedulePeriod{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billingSchedulePeriod, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodIDsByScheduleIDAndStartTimeForUpdate(ctx, mockDB.DB, scheduleID, time.Now())
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billingSchedulePeriod)
	})

	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, scheduleID, mock.Anything, mock.Anything)
		billingSchedulePeriod, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodIDsByScheduleIDAndStartTimeForUpdate(ctx, mockDB.DB, scheduleID, time.Now())
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billingSchedulePeriod)
	})
}

func TestBillingSchedulePeriodRepo_GetPeriodIDsInRangeTimeByScheduleID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var scheduleID string = "1"
	billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, scheduleID, mock.Anything, mock.Anything)
		e := &entities.BillingSchedulePeriod{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		billingSchedulePeriod, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodIDsInRangeTimeByScheduleID(ctx, mockDB.DB, scheduleID, time.Now(), time.Now())
		assert.Nil(t, err)
		assert.NotNil(t, billingSchedulePeriod)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, scheduleID, mock.Anything, mock.Anything)
		e := &entities.BillingSchedulePeriod{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		billingRatio, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodIDsInRangeTimeByScheduleID(ctx, mockDB.DB, scheduleID, time.Now(), time.Now())
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billingRatio)
	})

	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, scheduleID, mock.Anything, mock.Anything)
		billingRatio, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodIDsInRangeTimeByScheduleID(ctx, mockDB.DB, scheduleID, time.Now(), time.Now())
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, billingRatio)
	})
}

func TestBillingSchedulePeriodRepo_GetLatestPeriodByScheduleIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var scheduleID string = "1"
	billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			scheduleID,
		)
		entity := &entities.BillingSchedulePeriod{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		period, err := billingSchedulePeriodRepoWithSqlMock.GetLatestPeriodByScheduleIDForUpdate(ctx, mockDB.DB, scheduleID)
		assert.Nil(t, err)
		assert.NotNil(t, period)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			scheduleID,
		)
		e := &entities.BillingSchedulePeriod{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		period, err := billingSchedulePeriodRepoWithSqlMock.GetLatestPeriodByScheduleIDForUpdate(ctx, mockDB.DB, scheduleID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, period)
	})
}

func TestBillingSchedulePeriodRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.BillingSchedulePeriod{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := billingSchedulePeriodRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert billing schedule period fail", func(t *testing.T) {
		billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := billingSchedulePeriodRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BillingSchedulePeriod: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert billing schedule period", func(t *testing.T) {
		billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := billingSchedulePeriodRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BillingSchedulePeriod: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBillingSchedulePeriodRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.BillingSchedulePeriod{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := billingSchedulePeriodRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert billing schedule period fail", func(t *testing.T) {
		billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := billingSchedulePeriodRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update BillingSchedulePeriod: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert billing schedule period", func(t *testing.T) {
		billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := billingSchedulePeriodRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update BillingSchedulePeriod: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBillingSchedulePeriodRepo_GetPeriodByScheduleIDAndEndTime(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var scheduleID string = "1"
	endTime := time.Now()
	billingSchedulePeriodRepoWithSqlMock, mockDB := BillingSchedulePeriodRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			scheduleID,
			endTime,
		)
		entity := &entities.BillingSchedulePeriod{}
		fields, values := entity.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		period, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodByScheduleIDAndEndTime(ctx, mockDB.DB, scheduleID, endTime)
		assert.Nil(t, err)
		assert.NotNil(t, period)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			scheduleID,
			endTime,
		)
		e := &entities.BillingSchedulePeriod{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := billingSchedulePeriodRepoWithSqlMock.GetPeriodByScheduleIDAndEndTime(ctx, mockDB.DB, scheduleID, endTime)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)
	})
}
