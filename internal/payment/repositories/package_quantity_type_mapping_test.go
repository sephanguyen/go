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

func PackageQuantityTypeMappingRepoWithSqlMock() (*PackageQuantityTypeMappingRepo, *testutil.MockDB) {
	packageQuantityTypeMappingRepo := &PackageQuantityTypeMappingRepo{}
	return packageQuantityTypeMappingRepo, testutil.NewMockDB()
}

func TestPackageQuantityTypeMappingRepo_GetByPackageTypeForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	packageType := "package_one_time"
	packageQuantityTypeMappingRepoWithSqlMock, mockDB := PackageQuantityTypeMappingRepoWithSqlMock()
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, packageType)
		e := &entities.PackageQuantityTypeMapping{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		packageQuantityTypeMapping, err := packageQuantityTypeMappingRepoWithSqlMock.GetByPackageTypeForUpdate(ctx, mockDB.DB, packageType)
		assert.Nil(t, err)
		assert.NotNil(t, packageQuantityTypeMapping)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, packageType)
		e := &entities.PackageQuantityTypeMapping{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(fmt.Errorf("something error"), fields, values)
		_, err := packageQuantityTypeMappingRepoWithSqlMock.GetByPackageTypeForUpdate(ctx, mockDB.DB, packageType)
		assert.NotNil(t, err)
	})
}

func TestPackageQuantityTypeMappingRepo_Upsert(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := entities.PackageQuantityTypeMapping{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)
	t.Run("happy case", func(t *testing.T) {
		packageQuantityTypeMappingRepoWithSqlMock, mockDB := PackageQuantityTypeMappingRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)

		err := packageQuantityTypeMappingRepoWithSqlMock.Upsert(ctx, mockDB.DB, &mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("insert package quantity type mapping fail", func(t *testing.T) {
		packageQuantityTypeMappingRepoWithSqlMock, mockDB := PackageQuantityTypeMappingRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := packageQuantityTypeMappingRepoWithSqlMock.Upsert(ctx, mockDB.DB, &mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err upsert PackageQuantityTypeMappingRepo: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("no rows affect after insert package quantity type mapping", func(t *testing.T) {
		packageQuantityTypeMappingRepoWithSqlMock, mockDB := PackageQuantityTypeMappingRepoWithSqlMock()
		mockDB.DB.On("Exec", args...).Return(constant.FailCommandTag, nil)

		err := packageQuantityTypeMappingRepoWithSqlMock.Upsert(ctx, mockDB.DB, &mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err upsert PackageQuantityTypeMappingRepo: 0 RowsAffected").Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}
