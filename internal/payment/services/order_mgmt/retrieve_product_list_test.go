package ordermgmt

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRetrieveProductListService_RetrieveListOfProductsWithFilter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var (
		db              *mockDb.Ext
		productService  *mockServices.IProductServiceForProductList
		locationService *mockServices.ILocationServiceForProductList
	)
	var (
		productTypeFilter []*pb.ProductSpecificType
		gradeFilter       []string
		mockProducts      []entities.Product
		mockLocationIDs   []string
		mockLocations     []entities.Location
		mockGradeIds      []string
		expectedResp      *pb.RetrieveListOfProductsResponse
		mockItems         []*pb.RetrieveListOfProductsResponse_Product
		mockLocationInfor []*pb.LocationInfo
		mockGradeNames    []string
		//mockMapProductIDsWithLocationIDs map[string][]string
	)

	now := time.Now()
	productTypeFilter = append(productTypeFilter, &pb.ProductSpecificType{
		ProductType:  pb.ProductType_PRODUCT_TYPE_FEE,
		PackageType:  0,
		MaterialType: 0,
		FeeType:      pb.FeeType_FEE_TYPE_NONE,
	})
	gradeFilter = append(gradeFilter, "grade-id")
	mockGradeIds = []string{
		"grade1", "grade2",
	}
	mockGradeNames = []string{
		"grade1", "grade2",
	}
	mockProducts = []entities.Product{
		{
			ProductID: pgtype.Text{
				String: "1",
			},
			Name: pgtype.Text{
				String: "1",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "2",
			},
			Name: pgtype.Text{
				String: "2",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "3",
			},
			Name: pgtype.Text{
				String: "3",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "4",
			},
			Name: pgtype.Text{
				String: "4",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "5",
			},
			Name: pgtype.Text{
				String: "5",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "6",
			},
			Name: pgtype.Text{
				String: "6",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -2),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -1),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "7",
			},
			Name: pgtype.Text{
				String: "7",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -2),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, -1),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "8",
			},
			Name: pgtype.Text{
				String: "8",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 2),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "9",
			},
			Name: pgtype.Text{
				String: "9",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 2),
			},
		},
		{
			ProductID: pgtype.Text{
				String: "10",
			},
			Name: pgtype.Text{
				String: "10",
			},
			ProductType: pgtype.Text{
				String: pb.ProductType_PRODUCT_TYPE_FEE.String(),
			},
			UpdatedAt: pgtype.Timestamptz{Time: now},
			CreatedAt: pgtype.Timestamptz{Time: now},
			AvailableFrom: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 1),
			},
			AvailableUntil: pgtype.Timestamptz{
				Time: now.AddDate(0, 0, 2),
			},
		},
	}

	mockLocationIDs = []string{
		"Location_1",
		"Location_2",
	}
	mockLocations = []entities.Location{
		{
			LocationID: pgtype.Text{
				String: constant.LocationID,
			},
			Name: pgtype.Text{String: constant.LocationName},
		},
		{
			LocationID: pgtype.Text{
				String: "location_id_1",
			},
			Name: pgtype.Text{String: constant.LocationName},
		},
	}
	mockLocationInfor = []*pb.LocationInfo{
		{
			LocationId:   "",
			LocationName: "",
		},
		{
			LocationId:   "",
			LocationName: "",
		},
	}
	mockItems = []*pb.RetrieveListOfProductsResponse_Product{
		{
			ProductName:   "1",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_ACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "2",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_ACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "3",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_ACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "4",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_ACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "5",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_ACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "6",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_INACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "7",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_INACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "8",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_INACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "9",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_INACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
		{
			ProductName:   "10",
			ProductStatus: pb.ProductStatus_PRODUCT_STATUS_INACTIVE,
			Grades:        mockGradeIds,
			LocationInfo:  mockLocationInfor,
		},
	}

	expectedResp = &pb.RetrieveListOfProductsResponse{
		Items:    mockItems,
		NextPage: nil,
		PreviousPage: &cpb.Paging{
			Limit: 10,
			Offset: &cpb.Paging_OffsetInteger{
				OffsetInteger: 0,
			},
		},
		TotalItems:      10,
		TotalOfActive:   5,
		TotalOfInactive: 5,
	}

	testCases := []utils.TestCase{
		{
			Name: "Failed case: while get product stats",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes:  productTypeFilter,
					StudentGrades: gradeFilter,
				},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("error while get product stats: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				productService.On("GetProductStatsByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductStats{
					TotalItems: pgtype.Int8{
						Int: 10,
					},
					TotalOfActive: pgtype.Int8{
						Int: 5,
					},
					TotalOfInactive: pgtype.Int8{
						Int: 5,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: while get list products by filter",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes:  productTypeFilter,
					StudentGrades: gradeFilter,
				},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("error while get list of products by filter: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				productService.On("GetProductStatsByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductStats{
					TotalItems: pgtype.Int8{
						Int: 10,
					},
					TotalOfActive: pgtype.Int8{
						Int: 5,
					},
					TotalOfInactive: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				productService.On("GetListOfProductsByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockProducts, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: while get location of products as a map - Get location by ID",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes:  productTypeFilter,
					StudentGrades: gradeFilter,
				},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("error while get locations of products: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				productService.On("GetProductStatsByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductStats{
					TotalItems: pgtype.Int8{
						Int: 10,
					},
					TotalOfActive: pgtype.Int8{
						Int: 10,
					},
					TotalOfInactive: pgtype.Int8{
						Int: 0,
					},
				}, nil)
				productService.On("GetListOfProductsByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockProducts, nil)
				productService.On("GetLocationIDsWithProductID", mock.Anything, mock.Anything, mock.Anything).Return([]string{}, constant.ErrDefault)
			},
		},
		{
			Name: "Failed case: while get grade of products as a map",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes:  productTypeFilter,
					StudentGrades: gradeFilter,
				},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedResp: nil,
			ExpectedErr:  fmt.Errorf("error while get grades of products: error while get grade of product 1: %v", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				productService.On("GetProductStatsByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductStats{
					TotalItems: pgtype.Int8{
						Int: 10,
					},
					TotalOfActive: pgtype.Int8{
						Int: 10,
					},
					TotalOfInactive: pgtype.Int8{
						Int: 0,
					},
				}, nil)
				productService.On("GetListOfProductsByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockProducts, nil)
				productService.On("GetLocationIDsWithProductID", mock.Anything, mock.Anything, mock.Anything).Return(mockLocationIDs, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(mockLocations, nil)
				productService.On("GetGradeIDsByProductID", mock.Anything, mock.Anything, mock.Anything).Return(mockGradeIds, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes:  productTypeFilter,
					StudentGrades: gradeFilter,
				},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedResp: expectedResp,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				productService.On("GetProductStatsByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductStats{
					TotalItems: pgtype.Int8{
						Int: 10,
					},
					TotalOfActive: pgtype.Int8{
						Int: 5,
					},
					TotalOfInactive: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				productService.On("GetListOfProductsByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockProducts, nil)
				productService.On("GetLocationIDsWithProductID", mock.Anything, mock.Anything, mock.Anything).Return(mockLocationIDs, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(mockLocations, nil)
				productService.On("GetGradeIDsByProductID", mock.Anything, mock.Anything, mock.Anything).Return(mockGradeIds, nil)
				productService.On("GetGradeNamesByIDs", mock.Anything, mock.Anything, mock.Anything).Return(mockGradeNames, nil)
				productService.On("GetProductTypeByProductID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pb.ProductSpecificType{
					ProductType:  pb.ProductType_PRODUCT_TYPE_FEE,
					PackageType:  0,
					MaterialType: 0,
					FeeType:      pb.FeeType_FEE_TYPE_ONE_TIME,
				}, nil)
			},
		},
		{
			Name: "HappyCaseWithoutFilter",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes:  nil,
					StudentGrades: nil,
				},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10},
				},
			},
			ExpectedResp: expectedResp,
			ExpectedErr:  nil,
			Setup: func(ctx context.Context) {
				productService.On("GetProductStatsByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductStats{
					TotalItems: pgtype.Int8{
						Int: 10,
					},
					TotalOfActive: pgtype.Int8{
						Int: 5,
					},
					TotalOfInactive: pgtype.Int8{
						Int: 5,
					},
				}, nil)
				productService.On("GetListOfProductsByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockProducts, nil)
				productService.On("GetLocationIDsWithProductID", mock.Anything, mock.Anything, mock.Anything).Return(mockLocationIDs, nil)
				locationService.On("GetLocationsByIDs", mock.Anything, mock.Anything, mock.Anything).Return(mockLocations, nil)
				productService.On("GetGradeIDsByProductID", mock.Anything, mock.Anything, mock.Anything).Return(mockGradeIds, nil)
				productService.On("GetGradeNamesByIDs", mock.Anything, mock.Anything, mock.Anything).Return(mockGradeNames, nil)
				productService.On("GetProductTypeByProductID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(pb.ProductSpecificType{
					ProductType:  pb.ProductType_PRODUCT_TYPE_FEE,
					PackageType:  0,
					MaterialType: 0,
					FeeType:      pb.FeeType_FEE_TYPE_ONE_TIME,
				}, nil)
			},
		},
		{
			Name: "HappyCase when have no product",
			Ctx:  interceptors.ContextWithUserID(ctx, "user-id"),
			Req: &pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes:  nil,
					StudentGrades: nil,
				},
				Paging: &cpb.Paging{
					Limit:  10,
					Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 0},
				},
			},
			ExpectedResp: &pb.RetrieveListOfProductsResponse{
				TotalItems:      1,
				TotalOfActive:   0,
				TotalOfInactive: 0,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				productService.On("GetProductStatsByFilter", mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductStats{
					TotalItems: pgtype.Int8{
						Int: 1,
					},
					TotalOfActive: pgtype.Int8{
						Int: 0,
					},
					TotalOfInactive: pgtype.Int8{
						Int: 0,
					},
				}, nil)
				productService.On("GetListOfProductsByFilter", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)

			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)

			productService = new(mockServices.IProductServiceForProductList)
			locationService = new(mockServices.ILocationServiceForProductList)

			testCase.Setup(testCase.Ctx)
			s := &ProductList{
				DB:              db,
				ProductService:  productService,
				LocationService: locationService,
			}
			req := testCase.Req.(*pb.RetrieveListOfProductsRequest)
			resp, err := s.RetrieveListOfProducts(testCase.Ctx, req)
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, testCase.ExpectedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				expectedResp := testCase.ExpectedResp.(*pb.RetrieveListOfProductsResponse)
				assert.Equal(t, len(expectedResp.Items), len(resp.Items))
				for idx, expectedItem := range expectedResp.Items {
					item := resp.Items[idx]
					assert.Equal(t, expectedItem.ProductName, item.ProductName)
					assert.Equal(t, expectedItem.ProductStatus, item.ProductStatus)
					assert.Equal(t, expectedItem.Grades, item.Grades)
					assert.Equal(t, expectedItem.LocationInfo, item.LocationInfo)
				}

				if expectedResp.PreviousPage == nil {
					assert.Nil(t, resp.PreviousPage)
				} else {
					assert.Equal(t, expectedResp.PreviousPage.GetOffsetInteger(), resp.PreviousPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.PreviousPage.Limit, resp.PreviousPage.Limit)
				}

				if expectedResp.NextPage == nil {
					assert.Nil(t, resp.NextPage)
				} else {
					assert.Equal(t, expectedResp.NextPage.GetOffsetInteger(), resp.NextPage.GetOffsetInteger())
					assert.Equal(t, expectedResp.NextPage.Limit, resp.NextPage.Limit)
				}
				assert.Equal(t, expectedResp.TotalItems, resp.TotalItems)
				assert.Equal(t, expectedResp.TotalOfInactive, resp.TotalOfInactive)
				assert.Equal(t, expectedResp.TotalOfActive, resp.TotalOfActive)
			}
		})
	}
}
