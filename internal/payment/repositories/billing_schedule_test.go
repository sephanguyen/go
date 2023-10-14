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

func BillingScheduleRepoWithSqlMock() (*BillingScheduleRepo, *testutil.MockDB) {
	billingScheduleRepo := &BillingScheduleRepo{}
	return billingScheduleRepo, testutil.NewMockDB()
}

func TestBillingScheduleRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var billingScheduleID string = "1"
	t.Run(constant.HappyCase, func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			billingScheduleID,
		)
		entity := &entities.BillingSchedule{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		billingSchedule, err := billingScheduleRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, billingScheduleID)
		assert.Nil(t, err)
		assert.NotNil(t, billingSchedule)
	})
	t.Run("err case", func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			billingScheduleID,
		)
		e := &entities.BillingSchedule{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		billingSchedule, err := billingScheduleRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, billingScheduleID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, billingSchedule)

	})
}

func TestBillingScheduleRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.BillingSchedule{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := billingScheduleRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert billing schedule fail", func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := billingScheduleRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BillingSchedule: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert billing schedule", func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := billingScheduleRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert BillingSchedule: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestBillingScheduleRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.BillingSchedule{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := billingScheduleRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert billing schedule fail", func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := billingScheduleRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update BillingSchedule: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert billing schedule", func(t *testing.T) {
		billingScheduleRepoWithSqlMock, mockDB := BillingScheduleRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := billingScheduleRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update BillingSchedule: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
