package repositories

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ProductPriceRepoWithSqlMock() (*ProductPriceRepo, *testutil.MockDB) {
	productPriceRepo := &ProductPriceRepo{}
	return productPriceRepo, testutil.NewMockDB()
}

func TestProductPriceRepo_GetByProductIDAndPriceType(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var productId string = "1"
	priceType := pb.ProductPriceType_DEFAULT_PRICE.String()
	productPriceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()
	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, productId, mock.Anything)
		e := &entities.ProductPrice{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		productPrice, err := productPriceRepoWithSqlMock.GetByProductIDAndPriceType(ctx, mockDB.DB, productId, priceType)
		assert.Nil(t, err)
		assert.NotNil(t, productPrice)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything, productId, mock.Anything)
		e := &entities.ProductPrice{}
		fields, values := e.FieldMap()
		mockDB.MockScanArray(pgx.ErrNoRows, fields, [][]interface{}{
			values,
		})
		productPrice, err := productPriceRepoWithSqlMock.GetByProductIDAndPriceType(ctx, mockDB.DB, productId, priceType)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, productPrice)
	})

	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything, productId, mock.Anything)
		productPrice, err := productPriceRepoWithSqlMock.GetByProductIDAndPriceType(ctx, mockDB.DB, productId, priceType)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, productPrice)
	})
}

func TestProductPriceRepo_GetByProductIDAndQuantityAndPriceType(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	priceType := pb.ProductPriceType_DEFAULT_PRICE.String()
	t.Run(constant.HappyCase, func(t *testing.T) {
		priceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)

		price := &entities.ProductPrice{}
		fields, _ := price.FieldMap()
		scanFields := database.GetScanFields(price, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		_, err := priceRepoWithSqlMock.GetByProductIDAndQuantityAndPriceType(ctx, mockDB.DB, "1", 1, priceType)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		priceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)

		price := &entities.ProductPrice{}
		fields, _ := price.FieldMap()
		scanFields := database.GetScanFields(price, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		_, err := priceRepoWithSqlMock.GetByProductIDAndQuantityAndPriceType(ctx, mockDB.DB, "1", 1, priceType)
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestProductPriceRepo_GetByProductIDAndBillingSchedulePeriodIDAndQuantity(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	priceType := pb.ProductPriceType_DEFAULT_PRICE.String()
	t.Run(constant.HappyCase, func(t *testing.T) {
		priceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)

		price := &entities.ProductPrice{}
		fields, _ := price.FieldMap()
		scanFields := database.GetScanFields(price, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		_, err := priceRepoWithSqlMock.GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(ctx, mockDB.DB, "1", "1", 1, priceType)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		priceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)

		price := &entities.ProductPrice{}
		fields, _ := price.FieldMap()
		scanFields := database.GetScanFields(price, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		_, err := priceRepoWithSqlMock.GetByProductIDAndBillingSchedulePeriodIDAndQuantityAndPriceType(ctx, mockDB.DB, "1", "1", 1, priceType)
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestProductPriceRepo_GetByProductIDAndBillingSchedulePeriodID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	priceType := pb.ProductPriceType_DEFAULT_PRICE.String()
	t.Run(constant.HappyCase, func(t *testing.T) {
		priceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)

		price := &entities.ProductPrice{}
		fields, _ := price.FieldMap()
		scanFields := database.GetScanFields(price, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		_, err := priceRepoWithSqlMock.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, mockDB.DB, "1", "1", priceType)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run(constant.FailCaseErrorRow, func(t *testing.T) {
		priceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("QueryRow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)

		price := &entities.ProductPrice{}
		fields, _ := price.FieldMap()
		scanFields := database.GetScanFields(price, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		_, err := priceRepoWithSqlMock.GetByProductIDAndBillingSchedulePeriodIDAndPriceType(ctx, mockDB.DB, "1", "1", priceType)
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestProductPriceRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.ProductPrice{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		productPriceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := productPriceRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert product_price fail", func(t *testing.T) {
		productPriceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := productPriceRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert ProductPrice: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert product_price", func(t *testing.T) {
		productPriceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := productPriceRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert ProductPrice: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestProductPriceRepo_Delete(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.ProductPrice{}
	err := mockEntities.ProductID.Set("1")
	if err != nil {
		return
	}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run(constant.HappyCase, func(t *testing.T) {
		productPriceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := productPriceRepoWithSqlMock.DeleteByProductID(ctx, mockDB.DB, mockEntities.ProductID)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("delete product_price fail", func(t *testing.T) {
		productPriceRepoWithSqlMock, mockDB := ProductPriceRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := productPriceRepoWithSqlMock.DeleteByProductID(ctx, mockDB.DB, mockEntities.ProductID)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err delete ProductPrice: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
