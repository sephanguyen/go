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

func AccountingCategoryRepoWithSqlMock() (*AccountingCategoryRepo, *testutil.MockDB) {
	accountingCategoryRepo := &AccountingCategoryRepo{}
	return accountingCategoryRepo, testutil.NewMockDB()
}

func TestAccountingCategoryRepo_Create(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.AccountingCategory{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		repo, mockDB := AccountingCategoryRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert accounting category fail", func(t *testing.T) {
		accountingCategoryRepoWithSqlMock, mockDB := AccountingCategoryRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := accountingCategoryRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert AccountingCategory: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert accounting category", func(t *testing.T) {
		accountingCategoryRepoWithSqlMock, mockDB := AccountingCategoryRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := accountingCategoryRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert AccountingCategory: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestAccountingCategoryRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.AccountingCategory{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("happy case", func(t *testing.T) {
		accountingCategoryRepoWithSqlMock, mockDB := AccountingCategoryRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := accountingCategoryRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert accounting category fail", func(t *testing.T) {
		accountingCategoryRepoWithSqlMock, mockDB := AccountingCategoryRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := accountingCategoryRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update AccountingCategory: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert accounting category", func(t *testing.T) {
		accountingCategoryRepoWithSqlMock, mockDB := AccountingCategoryRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := accountingCategoryRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update AccountingCategory: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
