package service

import (
	"context"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestProductService_VerifiedProductReturnProductInfoAndBillingType(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		productRepo         *mockRepositories.MockProductRepo
		productGradeRepo    *mockRepositories.MockProductGradeRepo
		productLocationRepo *mockRepositories.MockProductLocationRepo
		productSettingRepo  *mockRepositories.MockProductSettingRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get by id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting product info with id %v has error %v ", constant.ProductID, constant.ErrDefault),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: Error when product type is invalid",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "product type of product id %v is invalid ", constant.ProductID),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
					ProductType: pgtype.Text{
						String: "Invalid_product_type",
					},
				}, nil)
			},
		},
		{
			Name:        "Happy case: No product setting",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
					},
				}, nil)
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{
					ProductID: pgtype.Text{String: constant.ProductID},
					IsEnrollmentRequired: pgtype.Bool{
						Bool: false,
					},
					IsPausable: pgtype.Bool{
						Bool: true,
					},
				}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
					},
				}, nil)
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{
					ProductID: pgtype.Text{String: constant.ProductID},
					IsEnrollmentRequired: pgtype.Bool{
						Bool: false,
					},
					IsPausable: pgtype.Bool{
						Bool: true,
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productRepo = new(mockRepositories.MockProductRepo)
			productGradeRepo = new(mockRepositories.MockProductGradeRepo)
			productLocationRepo = new(mockRepositories.MockProductLocationRepo)
			productSettingRepo = new(mockRepositories.MockProductSettingRepo)
			testCase.Setup(testCase.Ctx)
			s := &ProductService{
				productRepo:         productRepo,
				productGradeRepo:    productGradeRepo,
				productLocationRepo: productLocationRepo,
				productSettingRepo:  productSettingRepo,
			}
			req := testCase.Req.(utils.OrderItemData)
			_, _, _, _, _, err := s.VerifiedProductReturnProductInfoAndBillingType(testCase.Ctx, db, req)

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productRepo, productGradeRepo, productLocationRepo, productSettingRepo)
		})
	}
}

func TestProductService_VerifiedProductWithStudentInfoReturnProductInfoAndBillingType(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                  *mockDb.Ext
		productRepo         *mockRepositories.MockProductRepo
		productGradeRepo    *mockRepositories.MockProductGradeRepo
		productLocationRepo *mockRepositories.MockProductLocationRepo
		gradeRepo           *mockRepositories.MockGradeRepo
		productSettingRepo  *mockRepositories.MockProductSettingRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Fail case: Error when get grade by grade id ",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				gradeRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Grade{}, constant.ErrDefault)
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
					},
				}, nil)
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{
					ProductID: pgtype.Text{String: constant.ProductID},
					IsEnrollmentRequired: pgtype.Bool{
						Bool: false,
					},
					IsPausable: pgtype.Bool{
						Bool: true,
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get product grade by grade and product id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				gradeRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Grade{}, nil)
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
					},
				}, nil)
				productGradeRepo.On("GetByGradeAndProductIDForUpdate", ctx, db, mock.Anything, mock.Anything).Return(entities.ProductGrade{}, constant.ErrDefault)
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{
					ProductID: pgtype.Text{String: constant.ProductID},
					IsEnrollmentRequired: pgtype.Bool{
						Bool: false,
					},
					IsPausable: pgtype.Bool{
						Bool: true,
					},
				}, nil)
			},
		},
		{
			Name:        "Fail case: Error when get product location by location id and product id for update",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "getting product location with id %v has error %v ", constant.ProductID, constant.ErrDefault),
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				gradeRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Grade{}, nil)
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
					},
				}, nil)
				productGradeRepo.On("GetByGradeAndProductIDForUpdate", ctx, db, mock.Anything, mock.Anything).Return(entities.ProductGrade{}, nil)
				productLocationRepo.On("GetByLocationIDAndProductIDForUpdate", ctx, db, mock.Anything, mock.Anything).Return(entities.ProductLocation{}, constant.ErrDefault)
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{
					ProductID: pgtype.Text{String: constant.ProductID},
					IsEnrollmentRequired: pgtype.Bool{
						Bool: false,
					},
					IsPausable: pgtype.Bool{
						Bool: true,
					},
				}, nil)
			},
		},
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: utils.OrderItemData{
				OrderItem: &pb.OrderItem{
					ProductId: constant.ProductID,
				},
			},
			Setup: func(ctx context.Context) {
				gradeRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Grade{}, nil)
				productRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{String: constant.ProductID},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
					},
				}, nil)
				productGradeRepo.On("GetByGradeAndProductIDForUpdate", ctx, db, mock.Anything, mock.Anything).Return(entities.ProductGrade{}, nil)
				productLocationRepo.On("GetByLocationIDAndProductIDForUpdate", ctx, db, mock.Anything, mock.Anything).Return(entities.ProductLocation{}, nil)
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.ProductSetting{
					ProductID: pgtype.Text{String: constant.ProductID},
					IsEnrollmentRequired: pgtype.Bool{
						Bool: false,
					},
					IsPausable: pgtype.Bool{
						Bool: true,
					},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productRepo = new(mockRepositories.MockProductRepo)
			productGradeRepo = new(mockRepositories.MockProductGradeRepo)
			productLocationRepo = new(mockRepositories.MockProductLocationRepo)
			gradeRepo = new(mockRepositories.MockGradeRepo)
			productSettingRepo = new(mockRepositories.MockProductSettingRepo)
			testCase.Setup(testCase.Ctx)
			s := &ProductService{
				productRepo:         productRepo,
				productGradeRepo:    productGradeRepo,
				productLocationRepo: productLocationRepo,
				gradeRepo:           gradeRepo,
				productSettingRepo:  productSettingRepo,
			}
			req := testCase.Req.(utils.OrderItemData)
			_, _, _, _, _, _, err := s.VerifiedProductWithStudentInfoReturnProductInfoAndBillingType(testCase.Ctx, db, req)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productRepo, productGradeRepo, productLocationRepo, gradeRepo)
		})
	}
}

func TestProductService_GetProductsByIDs(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db          *mockDb.Ext
		productRepo *mockRepositories.MockProductRepo
	)
	testcases := []utils.TestCase{
		{
			Name:        "Failed case: Error when getting products by ids",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Success case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				productRepo.On("GetByIDs", ctx, mock.Anything, mock.Anything).Return([]entities.Product{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &ProductService{
				productRepo: productRepo,
			}
			_, err := s.GetProductsByIDs(testCase.Ctx, db, []string{"1", "2"})

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, db, productRepo)
		})
	}
}

func TestProductService_GetProductTypeByProductID(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db           *mockDb.Ext
		packageRepo  *mockRepositories.MockPackageRepo
		materialRepo *mockRepositories.MockMaterialRepo
		feeRepo      *mockRepositories.MockFeeRepo
	)
	casePackage := pb.ProductType_PRODUCT_TYPE_PACKAGE.String()
	caseMaterial := pb.ProductType_PRODUCT_TYPE_MATERIAL.String()
	caseFee := pb.ProductType_PRODUCT_TYPE_FEE.String()
	testcases := []utils.TestCase{
		{
			Name:        casePackage,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			ExpectedResp: pb.ProductSpecificType{
				ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
				PackageType:  pb.PackageType_PACKAGE_TYPE_ONE_TIME,
				MaterialType: 0,
				FeeType:      0,
			},
			Setup: func(ctx context.Context) {
				packageRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, "1").Return(entities.Package{PackageType: pgtype.Text{
					String: pb.PackageType_PACKAGE_TYPE_ONE_TIME.String(),
				}}, nil)
			},
		},
		{
			Name:        caseMaterial,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			ExpectedResp: pb.ProductSpecificType{
				ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL,
				PackageType:  0,
				MaterialType: pb.MaterialType_MATERIAL_TYPE_ONE_TIME,
				FeeType:      0,
			},
			Setup: func(ctx context.Context) {
				materialRepo.On("GetByIDForUpdate", mock.Anything, mock.Anything, "1").Return(entities.Material{MaterialType: pgtype.Text{
					String: pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String(),
				}}, nil)
			},
		},
		{
			Name:        caseFee,
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			ExpectedResp: pb.ProductSpecificType{
				ProductType:  pb.ProductType_PRODUCT_TYPE_FEE,
				PackageType:  0,
				MaterialType: 0,
				FeeType:      pb.FeeType_FEE_TYPE_RECURRING,
			},
			Setup: func(ctx context.Context) {
				feeRepo.On("GetFeeByID", mock.Anything, mock.Anything, "1").Return(entities.Fee{FeeType: pgtype.Text{
					String: pb.FeeType_FEE_TYPE_RECURRING.String(),
				}}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			packageRepo = new(mockRepositories.MockPackageRepo)
			materialRepo = new(mockRepositories.MockMaterialRepo)
			feeRepo = new(mockRepositories.MockFeeRepo)
			testCase.Setup(testCase.Ctx)
			s := &ProductService{
				packageRepo:  packageRepo,
				materialRepo: materialRepo,
				feeRepo:      feeRepo,
			}
			productType, err := s.GetProductTypeByProductID(testCase.Ctx, db, "1", testCase.Name)

			assert.Equal(t, testCase.ExpectedResp, productType)
			assert.Nil(t, err)
		})
	}
}

func TestProductService_GetProductStatsByFilter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db          *mockDb.Ext
		productRepo *mockRepositories.MockProductRepo
	)
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: entities.ProductStats{
				TotalItems: pgtype.Int8{
					Int:    2,
					Status: pgtype.Present,
				},
				TotalOfActive: pgtype.Int8{
					Int:    1,
					Status: pgtype.Present,
				},
				TotalOfInactive: pgtype.Int8{
					Int:    1,
					Status: pgtype.Present,
				},
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetProductStatsByFilter", ctx, mock.Anything, mock.Anything).Return(
					entities.ProductStats{
						TotalItems: pgtype.Int8{
							Int:    2,
							Status: pgtype.Present,
						},
						TotalOfActive: pgtype.Int8{
							Int:    1,
							Status: pgtype.Present,
						},
						TotalOfInactive: pgtype.Int8{
							Int:    1,
							Status: pgtype.Present,
						},
					}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &ProductService{
				productRepo: productRepo,
			}

			req := pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes: []*pb.ProductSpecificType{
						{
							ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
							PackageType:  pb.PackageType_PACKAGE_TYPE_SCHEDULED,
							MaterialType: 0,
							FeeType:      0,
						},
					},
					StudentGrades: []string{
						"grade-1",
					},
				},
				Keyword:       "product",
				ProductStatus: 0,
				Paging:        nil,
			}
			productStats, err := s.GetProductStatsByFilter(testCase.Ctx, db, &req)

			assert.Equal(t, testCase.ExpectedResp, productStats)
			assert.Nil(t, err)
		})
	}
}

func TestProductService_GetListOfProductsByFilter(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db          *mockDb.Ext
		productRepo *mockRepositories.MockProductRepo
	)
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: []entities.Product{
				{
					ProductID: pgtype.Text{
						String: "product-1",
						Status: pgtype.Present,
					},
				},
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetProductsByFilter", ctx, mock.Anything, mock.Anything).Return(
					[]entities.Product{
						{
							ProductID: pgtype.Text{
								String: "product-1",
								Status: pgtype.Present,
							},
						},
					}, nil)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productRepo = new(mockRepositories.MockProductRepo)
			testCase.Setup(testCase.Ctx)
			s := &ProductService{
				productRepo: productRepo,
			}

			req := pb.RetrieveListOfProductsRequest{
				Filter: &pb.RetrieveListOfProductsFilter{
					ProductTypes: []*pb.ProductSpecificType{
						{
							ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE,
							PackageType:  pb.PackageType_PACKAGE_TYPE_SCHEDULED,
							MaterialType: 0,
							FeeType:      0,
						},
					},
					StudentGrades: []string{
						"grade-1",
					},
				},
				Keyword:       "product",
				ProductStatus: 0,
				Paging:        nil,
			}
			productStats, err := s.GetListOfProductsByFilter(testCase.Ctx, db, &req, int64(10), int64(10))

			assert.Equal(t, testCase.ExpectedResp, productStats)
			assert.Nil(t, err)
		})
	}
}

func TestProductService_GetProductSettingByProductID(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()
	var (
		db                 *mockDb.Ext
		productSettingRepo *mockRepositories.MockProductSettingRepo
	)
	testCases := []utils.TestCase{
		{
			Name: "happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedResp: entities.ProductSetting{
				ProductID: pgtype.Text{
					String: "productID",
					Status: pgtype.Present,
				},
			},
			Setup: func(ctx context.Context) {
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(
					entities.ProductSetting{
						ProductID: pgtype.Text{
							String: "productID",
							Status: pgtype.Present,
						},
					}, nil)
			},
		},
		{
			Name:        "error case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: status.Errorf(codes.Internal, "error while get product setting by product id %v: %v", "productID", constant.ErrDefault),
			Setup: func(ctx context.Context) {
				productSettingRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(
					entities.ProductSetting{
						ProductID: pgtype.Text{
							String: "productID",
							Status: pgtype.Present,
						},
					}, constant.ErrDefault)
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			productSettingRepo = new(mockRepositories.MockProductSettingRepo)
			testCase.Setup(testCase.Ctx)
			s := &ProductService{
				productSettingRepo: productSettingRepo,
			}

			productStats, err := s.GetProductSettingByProductID(testCase.Ctx, db, "productID")
			if testCase.ExpectedErr != nil {
				assert.Equal(t, testCase.ExpectedErr, err)
			} else {
				assert.Equal(t, testCase.ExpectedResp, productStats)
				assert.Nil(t, err)
			}

		})
	}
}
