package repositories

import (
	"context"
	"errors"
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/mock/testutil"

	"github.com/jackc/pgx/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func ProductRepoWithSqlMock() (*ProductRepo, *testutil.MockDB) {
	productRepo := &ProductRepo{}
	return productRepo, testutil.NewMockDB()
}

func TestProductRepo_GetByIDForUpdate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	productRepoWithSqlMock, mockDB := ProductRepoWithSqlMock()

	const productID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productID,
		)
		entity := &entities.Product{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		product, err := productRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, productID)
		assert.Nil(t, err)
		assert.NotNil(t, product)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productID,
		)
		e := &entities.Product{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		product, err := productRepoWithSqlMock.GetByIDForUpdate(ctx, mockDB.DB, productID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, product)

	})
}

func TestProductRepo_GetByID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	productRepoWithSqlMock, mockDB := ProductRepoWithSqlMock()

	const productID string = "1"
	t.Run("Success", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productID,
		)
		entity := &entities.Product{}
		fields, values := entity.FieldMap()

		mockDB.MockRowScanFields(nil, fields, values)
		product, err := productRepoWithSqlMock.GetByID(ctx, mockDB.DB, productID)
		assert.Nil(t, err)
		assert.NotNil(t, product)

	})
	t.Run("err case", func(t *testing.T) {
		mockDB.MockQueryRowArgs(t, mock.Anything,
			mock.Anything,
			productID,
		)
		e := &entities.Product{}
		fields, values := e.FieldMap()

		mockDB.MockRowScanFields(pgx.ErrNoRows, fields, values)
		product, err := productRepoWithSqlMock.GetByID(ctx, mockDB.DB, productID)
		assert.True(t, errors.Is(err, pgx.ErrNoRows))
		assert.NotNil(t, product)

	})
}

func TestGetByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("Get products success", func(t *testing.T) {
		productRepoWithSqlMock, mockDB := ProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		product := &entities.Product{}
		fields, _ := product.FieldMap()
		scanFields := database.GetScanFields(product, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := productRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, []string{"1", "2"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Get order item fail", func(t *testing.T) {
		productRepoWithSqlMock, mockDB := ProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		product := &entities.Product{}
		fields, _ := product.FieldMap()
		scanFields := database.GetScanFields(product, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := productRepoWithSqlMock.GetByIDs(ctx, mockDB.DB, []string{"1", "2"})
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestGetByIDsForExport(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	t.Run("Get products success", func(t *testing.T) {
		productRepoWithSqlMock, mockDB := ProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		product := &entities.Product{}
		fields, _ := product.FieldMap()
		scanFields := database.GetScanFields(product, fields)
		rows.On("Scan", scanFields...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		_, err := productRepoWithSqlMock.GetByIDsForExport(ctx, mockDB.DB, []string{"1", "2"})
		assert.Nil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})
	t.Run("Get order item fail", func(t *testing.T) {
		productRepoWithSqlMock, mockDB := ProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		rows.On("Next").Times(1).Return(true)

		product := &entities.Product{}
		fields, _ := product.FieldMap()
		scanFields := database.GetScanFields(product, fields)
		rows.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
		rows.On("Close").Once().Return(nil)
		_, err := productRepoWithSqlMock.GetByIDsForExport(ctx, mockDB.DB, []string{"1", "2"})
		assert.NotNil(t, err)

		mock.AssertExpectationsForObjects(t, mockDB.DB, mockDB.Rows)
	})

}

func TestProductRepo_GetProductStatsByFilter(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		productRepoWithSqlMock *ProductRepo
		mockDB                 *testutil.MockDB

		productTypes = []*pb.ProductSpecificType{
			{
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageType:  pb.PackageType_PACKAGE_TYPE_ONE_TIME,
				MaterialType: 0,
				FeeType:      0,
			},
			{
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
				PackageType:  0,
				MaterialType: pb.MaterialType_MATERIAL_TYPE_ONE_TIME,
				FeeType:      0,
			},
		}
		studentGrades       = []string{"1", "2"}
		limit         int64 = 10
		offset        int64 = 1
	)

	expectedProductStats := &entities.ProductStats{
		TotalItems: pgtype.Int8{
			Int: 2,
		},
		TotalOfActive: pgtype.Int8{
			Int: 2,
		},
		TotalOfInactive: pgtype.Int8{},
	}
	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when scan",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: ProductListFilter{
				ProductTypes:  productTypes,
				StudentGrades: studentGrades,
				Limit:         &limit,
				Offset:        &offset,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productStats := &entities.ProductStats{}
				fields, values := productStats.FieldProductStatsMap()
				scanFields := utils.GetScanFields(fields, values, fields)

				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", scanFields...).Once().Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: ProductListFilter{
				ProductTypes:  productTypes,
				StudentGrades: studentGrades,
				Limit:         &limit,
				Offset:        &offset,
			},
			ExpectedResp: expectedProductStats,
			Setup: func(ctx context.Context) {
				productStats := &entities.ProductStats{}
				fields, values := productStats.FieldProductStatsMap()
				scanFields := utils.GetScanFields(fields, values, fields)

				mockDB.MockQueryRowArgs(t, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)

				mockDB.DB.On("QueryRow").Once().Return(mockDB.Row)
				mockDB.Row.On("Scan", scanFields...).Once().Run(func(args mock.Arguments) {
					refExpectedProductStats := expectedProductStats
					args[0].(*pgtype.Int8).Int = refExpectedProductStats.TotalItems.Int
					args[1].(*pgtype.Int8).Int = refExpectedProductStats.TotalOfActive.Int
					args[2].(*pgtype.Int8).Int = refExpectedProductStats.TotalOfInactive.Int
				}).Return(nil)
			},
		},
	}
	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			productRepoWithSqlMock, mockDB = ProductRepoWithSqlMock()
			testCase.Setup(testCase.Ctx)

			req := testCase.Req.(ProductListFilter)
			productStatsResp, err := productRepoWithSqlMock.GetProductStatsByFilter(testCase.Ctx, mockDB.DB, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				expectedResp := testCase.ExpectedResp.(*entities.ProductStats)
				assert.Equal(t, expectedResp.TotalItems, productStatsResp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfActive, productStatsResp.TotalOfActive)
				assert.Equal(t, expectedResp.TotalOfInactive, productStatsResp.TotalOfInactive)
				assert.Equal(t, testCase.ExpectedErr, err)
			}
		})
	}
}

func TestProductRepo_GetListOfProductsByFilter(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var filter ProductListFilter

	t.Run("Get list of products success", func(t *testing.T) {
		ProductRepoWithSqlMock, mockDB := ProductRepoWithSqlMock()
		rows := mockDB.Rows
		mockDB.DB.On("Query", mock.Anything, mock.Anything, mock.Anything).Once().Return(rows, nil)
		product := &entities.Product{}
		fields, _ := product.FieldMap()
		scanFiled := database.GetScanFields(product, fields)
		rows.On("Scan", scanFiled...).Once().Return(nil)
		rows.On("Next").Once().Return(false)
		rows.On("Close").Once().Return(nil)
		products, err := ProductRepoWithSqlMock.GetProductsByFilter(ctx, mockDB.DB, filter)
		var expectedProducts []entities.Product
		assert.Nil(t, err)
		assert.Equal(t, expectedProducts, products)
	})
}
