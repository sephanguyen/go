package repositories

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DiscountRepoWithSqlMock() (*DiscountRepo, *testutil.MockDB) {
	discountRepo := &DiscountRepo{}
	return discountRepo, testutil.NewMockDB()
}

func TestDiscountRepo_GetByDiscountType(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDiscountRepo, mockDB := DiscountRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discounts, err := mockDiscountRepo.GetByDiscountType(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.Equal(t, e, discounts[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discounts, err := mockDiscountRepo.GetByDiscountType(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discounts)
	})
}

func TestDiscountRepo_GetByDiscountTagIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDiscountRepo, mockDB := DiscountRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			nil,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discounts, err := mockDiscountRepo.GetByDiscountTagIDs(ctx, mockDB.DB, []string{mock.Anything})
		assert.Nil(t, err)
		assert.Equal(t, e, discounts[0])
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(
			t,
			pgx.ErrNoRows,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{values})

		discounts, err := mockDiscountRepo.GetByDiscountTagIDs(ctx, mockDB.DB, []string{mock.Anything})
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discounts)
	})
}

func TestDiscountRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDiscountRepo, mockDB := DiscountRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.Discount{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		discount, err := mockDiscountRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, discount)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := mockDiscountRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)
	})
}

func TestDiscountRepo_GetMaxDiscountByTypeAndDiscountTagIDs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDiscountRepo, mockDB := DiscountRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.Discount{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		discount, err := mockDiscountRepo.GetMaxDiscountByTypeAndDiscountTagIDs(ctx, mockDB.DB, mock.Anything, []string{mock.Anything})
		assert.Nil(t, err)
		assert.NotNil(t, discount)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := mockDiscountRepo.GetMaxDiscountByTypeAndDiscountTagIDs(ctx, mockDB.DB, mock.Anything, []string{mock.Anything})
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)
	})
}

func TestDiscountRepo_GetMaxProductDiscountByProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDiscountRepo, mockDB := DiscountRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.Discount{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		discount, err := mockDiscountRepo.GetMaxProductDiscountByProductID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, discount)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.Discount{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discount, err := mockDiscountRepo.GetMaxProductDiscountByProductID(ctx, mockDB.DB, mock.Anything)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, discount)
	})
}
