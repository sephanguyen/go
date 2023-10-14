package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"
	"github.com/manabie-com/backend/internal/discount/entities"
	mock_database "github.com/manabie-com/backend/mock/golibs/database"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ProductGroupRepoWithMock() (*ProductGroupRepo, *testutil.MockDB, *mock_database.Tx) {
	repo := &ProductGroupRepo{}
	return repo, testutil.NewMockDB(), &mock_database.Tx{}
}

func TestProductGroupRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.ProductGroup{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("insert product group succeeds", func(t *testing.T) {
		productGroupRepoWithSqlMock, mockDB, _ := ProductGroupRepoWithMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := productGroupRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert product setting fails", func(t *testing.T) {
		productGroupRepoWithSqlMock, mockDB, _ := ProductGroupRepoWithMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := productGroupRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert ProductGroup: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affected after insert product setting", func(t *testing.T) {
		productGroupRepoWithSqlMock, mockDB, _ := ProductGroupRepoWithMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := productGroupRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err insert ProductGroup: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestProductGroupRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.ProductGroup{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	t.Run("update product setting fields succeeds", func(t *testing.T) {
		repo, mockDB, _ := ProductGroupRepoWithMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := repo.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update product setting fields fails", func(t *testing.T) {
		productGroupRepoWithSqlMock, mockDB, _ := ProductGroupRepoWithMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := productGroupRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update ProductGroup: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affect after updating productGroup fields", func(t *testing.T) {
		productGroupRepoWithSqlMock, mockDB, _ := ProductGroupRepoWithMock()

		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := productGroupRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update ProductGroup: %d RowsAffected", constant.FailCommandTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestProductGroupRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	productGroupRepoWithSqlMock, mockDB, _ := ProductGroupRepoWithMock()
	var productGroupID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productGroupID,
		)
		entities := &entities.ProductGroup{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		productGroup, err := productGroupRepoWithSqlMock.GetByID(ctx, mockDB.DB, productGroupID)
		assert.Nil(t, err)
		assert.NotNil(t, productGroup)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productGroupID,
		)
		entities := &entities.ProductGroup{}
		fields, values := entities.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		productGroup, err := productGroupRepoWithSqlMock.GetByID(ctx, mockDB.DB, productGroupID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, productGroup)

	})
}
