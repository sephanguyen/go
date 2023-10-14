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

func TaxRepoWithSqlMock() (*TaxRepo, *testutil.MockDB) {
	taxRepo := &TaxRepo{}
	return taxRepo, testutil.NewMockDB()
}

func TestTaxRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	taxRepoWithSqlMock, mockDB := TaxRepoWithSqlMock()

	var taxID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			taxID,
		)
		e := &entities.Tax{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		tax, err := taxRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, taxID)
		assert.Nil(t, err)
		assert.NotNil(t, tax)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			taxID,
		)
		e := &entities.Tax{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		tax, err := taxRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, taxID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, tax)

	})
}

func TestTaxRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Tax{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("insert tax succeeds", func(t *testing.T) {
		taxRepoWithSqlMock, mockDB := TaxRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := taxRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert tax fails", func(t *testing.T) {
		taxRepoWithSqlMock, mockDB := TaxRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := taxRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Tax: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert tax", func(t *testing.T) {
		taxRepoWithSqlMock, mockDB := TaxRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := taxRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert Tax: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestTaxRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Tax{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("update tax fields succeeds", func(t *testing.T) {
		repo, mockDB := TaxRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update tax fields fails", func(t *testing.T) {
		taxRepoWithSqlMock, mockDB := TaxRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := taxRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Tax: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affect after updating tax fields", func(t *testing.T) {
		taxRepoWithSqlMock, mockDB := TaxRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := taxRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Tax: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
