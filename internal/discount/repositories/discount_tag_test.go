package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"errors"
	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func DiscountTagRepoWithSqlMock() (*DiscountTagRepo, *testutil.MockDB) {
	discountTagRepo := &DiscountTagRepo{}
	return discountTagRepo, testutil.NewMockDB()
}

func TestDiscountTagRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockDiscountTagRepo, mockDB := DiscountTagRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		entity := &entities.DiscountTag{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		discountTag, err := mockDiscountTagRepo.GetByID(ctx, mockDB.DB, mock.Anything)
		assert.Nil(t, err)
		assert.NotNil(t, discountTag)
	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(
			t,
			mock.Anything,
			mock.Anything,
			mock.Anything,
		)
		e := &entities.DiscountTag{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		discountTag, err := mockDiscountTagRepo.GetByID(ctx, mockDB.DB, mock.Anything)

		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, discountTag)
	})
}

func TestDiscountTagRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.DiscountTag{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		discountTagRepoWithSqlMock, mockDB := DiscountTagRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
		err := discountTagRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)
		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert discount tag fail", func(t *testing.T) {
		discountTagRepoWithSqlMock, mockDB := DiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := discountTagRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert DiscountTag: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert discount tag", func(t *testing.T) {
		discountTagRepoWithSqlMock, mockDB := DiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := discountTagRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert DiscountTag: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestDiscountTagRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.DiscountTag{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		discountTagRepoWithSqlMock, mockDB := DiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := discountTagRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert discount tag fail", func(t *testing.T) {
		discountTagRepoWithSqlMock, mockDB := DiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := discountTagRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update DiscountTag: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert discount tag", func(t *testing.T) {
		discountTagRepoWithSqlMock, mockDB := DiscountTagRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := discountTagRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update DiscountTag: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
