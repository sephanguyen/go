package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func OrderActionLogRepoWithSqlMock() (*OrderActionLogRepo, *testutil.MockDB) {
	orderActionLogRepo := &OrderActionLogRepo{}
	return orderActionLogRepo, testutil.NewMockDB()
}

func TestOrderActionLogRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.OrderActionLog{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("insert order action log succeeds", func(t *testing.T) {
		orderActionLogRepoWithSqlMock, mockDB := OrderActionLogRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := orderActionLogRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert order action log fails", func(t *testing.T) {
		orderActionLogRepoWithSqlMock, mockDB := OrderActionLogRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := orderActionLogRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert OrderActionLog: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert order action log", func(t *testing.T) {
		orderActionLogRepoWithSqlMock, mockDB := OrderActionLogRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := orderActionLogRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert OrderActionLog: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestOrderActionLogRepo_GetOrderCreatorsByOrderIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockOrderActionLogRepo, mockDB := OrderActionLogRepoWithSqlMock()
	orderIDs := []string{"1", "2", "3"}
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			orderIDs,
		)
		e := &entities.OrderCreator{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		orderCreators, err := mockOrderActionLogRepo.GetOrderCreatorsByOrderIDs(ctx, mockDB.DB, orderIDs)
		assert.Nil(t, err)
		assert.NotNil(t, orderCreators)

	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, mock.Anything,
			orderIDs,
		)
		e := &entities.OrderCreator{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		orderCreators, err := mockOrderActionLogRepo.GetOrderCreatorsByOrderIDs(ctx, mockDB.DB, orderIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, orderCreators)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, mock.Anything, orderIDs)
		orderCreators, err := mockOrderActionLogRepo.GetOrderCreatorsByOrderIDs(ctx, mockDB.DB, orderIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, orderCreators)
	})
}
