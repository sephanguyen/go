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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ProductSettingRepoWithSqlMock() (*ProductSettingRepo, *testutil.MockDB) {
	productSettingRepo := &ProductSettingRepo{}
	return productSettingRepo, testutil.NewMockDB()
}

func TestProductRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.ProductSetting{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("insert product setting succeeds", func(t *testing.T) {
		productSettingRepoWithSqlMock, mockDB := ProductSettingRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := productSettingRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert product setting fails", func(t *testing.T) {
		productSettingRepoWithSqlMock, mockDB := ProductSettingRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := productSettingRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert ProductSetting: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert product setting", func(t *testing.T) {
		productSettingRepoWithSqlMock, mockDB := ProductSettingRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := productSettingRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert ProductSetting: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestProductSettingRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.ProductSetting{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("update product setting fields succeeds", func(t *testing.T) {
		repo, mockDB := ProductSettingRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update product setting fields fails", func(t *testing.T) {
		productSettingRepoWithSqlMock, mockDB := ProductSettingRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := productSettingRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update ProductSetting: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affect after updating productSetting fields", func(t *testing.T) {
		productSettingRepoWithSqlMock, mockDB := ProductSettingRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := productSettingRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update ProductSetting: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestProductSettingRepoGetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	productID := "1"
	productSettingRepoWithSqlMock, mockDB := ProductSettingRepoWithSqlMock()
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, productID)
		e := &entities.ProductSetting{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		productSetting, err := productSettingRepoWithSqlMock.GetByID(ctx, mockDB.DB, productID)
		assert.Nil(t, err)
		assert.NotNil(t, productSetting)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, productID)
		e := &entities.ProductSetting{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(fmt.Errorf("something error"), fields, values)
		_, err := productSettingRepoWithSqlMock.GetByID(ctx, mockDB.DB, productID)
		assert.NotNil(t, err)
	})
}
