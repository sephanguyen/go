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

func DiscountRepoWithSqlMock() (*DiscountRepo, *testutil.MockDB) {
	discountRepo := &DiscountRepo{}
	return discountRepo, testutil.NewMockDB()
}

func TestDiscountRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

	const discountID string = "1"
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			discountID,
		)
		entity := &entities.Discount{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		discount, err := discountRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, discountID)
		assert.Nil(t, err)
		assert.NotNil(t, discount)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			discountID,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := discountRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, discountID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)

	})
}

func TestDiscountRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Discount{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := discountRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert discount fail", func(t *testing.T) {
		discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := discountRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Discount: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert discount", func(t *testing.T) {
		discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := discountRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Discount: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestDiscountRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Discount{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := discountRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert discount fail", func(t *testing.T) {
		discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := discountRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Discount: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert discount", func(t *testing.T) {
		discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := discountRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Discount: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestDiscountRepo_GetByIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	discountRepoWithSqlMock, mockDB := DiscountRepoWithSqlMock()

	discountIDs := []string{"10", "20", "30"}
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			discountIDs,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		discount, err := discountRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, discountIDs)
		assert.Nil(t, err)
		assert.NotNil(t, discount)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything,
			discountIDs,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrTxClosed, fields, [][]interface{}{
			values,
		})
		discount, err := discountRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, discountIDs)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, discount)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, discountIDs)
		discount, err := discountRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, discountIDs)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discount)
	})
}
