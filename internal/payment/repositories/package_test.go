package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func PackageRepoWithSqlMock() (*PackageRepo, *testutil.MockDB) {
	packageRepo := &PackageRepo{}
	return packageRepo, testutil.NewMockDB()
}

func TestPackageRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var productId string = "1"
	packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()
	t.Run("happy case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, productId)
		e := &entities.Package{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(nil, fields, values)
		productPrice, err := packageRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, productId)
		assert.Nil(t, err)
		assert.NotNil(t, productPrice)
	})
	t.Run("err case scan row", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, productId)
		e := &entities.Package{}
		fields, values := e.FieldMap()
		mockDB.MockRowScanFields(fmt.Errorf("something error"), fields, values)
		_, err := packageRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, productId)
		assert.NotNil(t, err)
	})
}

func TestPackageRepo_Create(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Package{}

	t.Run("insert package succeeds", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		product := entities.Product{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		pkg := &entities.Package{
			Product: product,
		}
		_, productValues := product.FieldMap()

		_, packageValues := pkg.FieldMap()

		argsProduct := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(productValues))...)
		argsPackage := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(packageValues))...)
		mockDB.DB.On("QueryRow", argsProduct...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)
		mockDB.DB.On("Exec", argsPackage...).Return(constant.SuccessCommandTag, nil)

		err := packageRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert product fails", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		product := entities.Product{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		_, productValues := product.FieldMap()

		argsProduct := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(productValues))...)
		mockDB.DB.On("QueryRow", argsProduct...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(fmt.Errorf("err something"))

		err := packageRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		require.NotNil(t, err)
		assert.Equal(t, "err insert Product: err something", err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("insert package fail", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		product := entities.Product{
			ResourcePath: pgtype.Text{Status: pgtype.Null},
		}
		pkg := &entities.Package{
			Product: product,
		}
		_, productValues := product.FieldMap()

		_, packageValues := pkg.FieldMap()

		argsProduct := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(productValues))...)
		argsPackage := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(packageValues))...)
		mockDB.DB.On("QueryRow", argsProduct...).Return(mockDB.Row)
		mockDB.Row.On("Scan", mock.Anything).Return(nil)
		mockDB.DB.On("Exec", argsPackage...).Return(constant.SuccessCommandTag, fmt.Errorf("err something"))

		err := packageRepoWithSqlMock.Create(ctx, mockDB.DB, mockEntities)
		require.NotNil(t, err)
		assert.Equal(t, "err insert PackageRepo: err something", err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPackageRepo_Update(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	mockEntities := &entities.Package{}
	_, fieldMap := mockEntities.FieldMap()

	args := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(fieldMap))...)

	_, productFieldMap := mockEntities.Product.FieldMap()

	productArgs := append([]interface{}{mock.Anything, mock.AnythingOfType("string")}, genSliceMock(len(productFieldMap))...)

	t.Run("update package fields succeeds", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, nil)
		mockDB.DB.On("Exec", productArgs...).Return(constant.SuccessCommandTag, nil)

		err := packageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update package fields succeeds", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		mockDB.DB.On("Exec", productArgs...).Return(constant.FailCommandTag, nil)

		err := packageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		require.NotNil(t, err)
		assert.Equal(t, "err update Product: 0 RowsAffected", err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update product fields fails", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		mockDB.DB.On("Exec", productArgs...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)

		err := packageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Product: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("update package fields fails", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		mockDB.DB.On("Exec", args...).Return(constant.SuccessCommandTag, pgx.ErrTxClosed)
		mockDB.DB.On("Exec", productArgs...).Return(constant.SuccessCommandTag, nil)

		err := packageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Package: %w", pgx.ErrTxClosed).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

	t.Run("no rows affect after updating package fields", func(t *testing.T) {
		packageRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

		cmdTag := constant.FailCommandTag
		productCmdTag := constant.SuccessCommandTag
		mockDB.DB.On("Exec", args...).Return(cmdTag, nil)
		mockDB.DB.On("Exec", productArgs...).Return(productCmdTag, nil)

		err := packageRepoWithSqlMock.Update(ctx, mockDB.DB, mockEntities)
		assert.NotNil(t, err)
		assert.Equal(t, fmt.Errorf("err update Package: %d RowsAffected", cmdTag.RowsAffected()).Error(), err.Error())

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
}

func TestPackageRepo_GetPackagesForExport(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	packageMockRepoWithSqlMock, mockDB := PackageRepoWithSqlMock()

	t.Run(constant.HappyCase, func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		e := &entities.Package{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(nil, fields, [][]interface{}{
			values,
		})
		packageMock, err := packageMockRepoWithSqlMock.GetPackagesForExport(ctx, mockDB.DB)
		assert.Nil(t, err)
		assert.NotNil(t, packageMock)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryArgs(t, nil, mock.Anything, mock.Anything)
		e := &entities.Package{}
		fields, values := e.FieldMap()

		mockDB.MockScanArray(pgx.ErrTxClosed, fields, [][]interface{}{
			values,
		})
		packageMock, err := packageMockRepoWithSqlMock.GetPackagesForExport(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrTxClosed))
		assert.Nil(t, packageMock)

	})
	t.Run("err case query", func(t *testing.T) {
		mockDB.MockQueryArgs(t, pgx.ErrNoRows, mock.Anything, mock.Anything)
		packageMock, err := packageMockRepoWithSqlMock.GetPackagesForExport(ctx, mockDB.DB)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.Nil(t, packageMock)
	})
}
