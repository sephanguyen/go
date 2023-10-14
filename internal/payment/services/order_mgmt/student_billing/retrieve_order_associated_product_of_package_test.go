package studentbilling

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRetrieveListOfOrderAssociatedProductOfPackages(t *testing.T) {
	t.Parallel()
	Ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		studentProductService *mockServices.IStudentProductForStudentBilling
		billItemService       *mockServices.IBillItemServiceForStudentBilling
		materialService       *mockServices.IMaterialServiceForStudentBilling
		orderService          *mockServices.IOrderServiceForStudentBilling
		packageService        *mockServices.IPackageServiceForStudentBilling
	)
	materialTypeOfBillItem := pb.MaterialType_MATERIAL_TYPE_ONE_TIME.String()
	packageTypeOfBillItem := pb.PackageType_PACKAGE_TYPE_ONE_TIME.String()
	feeTypeOfBillItem := pb.FeeType_FEE_TYPE_ONE_TIME.String()
	quantityType := pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT.String()
	var mapBillItems = make(map[string]*entities.BillItem, 2)
	mapBillItems["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
	}
	mapBillItems["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
	}
	invalidProductType := "invalid_product_type"
	invalidMaterialType := "invalid_material_type"
	invalidFeeType := "invalid_fee_type"
	invalidPackageType := "invalid_package_type"
	invalidQuantityType := "invalid_quantity_type"
	discountName := "discount_name"

	var invalidMapBillItems_ProductType = make(map[string]*entities.BillItem, 2)
	invalidMapBillItems_ProductType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  invalidProductType,
			DiscountName: &discountName,
		}),
	}
	invalidMapBillItems_ProductType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:   "10",
			ProductName: constant.ProductName,
			ProductType: invalidProductType,
		}),
	}

	var invalidMapBillItems_PackageType = make(map[string]*entities.BillItem, 2)
	invalidMapBillItems_PackageType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
			PackageType:  &invalidPackageType,
			DiscountName: &discountName,
		}),
	}
	invalidMapBillItems_PackageType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
			PackageType:  &invalidPackageType,
			DiscountName: &discountName,
		}),
	}

	var invalidMapBillItems_QuantityType = make(map[string]*entities.BillItem, 2)
	invalidMapBillItems_QuantityType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
			PackageType:  &packageTypeOfBillItem,
			QuantityType: &invalidQuantityType,
		}),
	}
	invalidMapBillItems_QuantityType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
			PackageType:  &packageTypeOfBillItem,
			QuantityType: &invalidQuantityType,
		}),
	}

	var invalidMapBillItems_MaterialType = make(map[string]*entities.BillItem, 2)
	invalidMapBillItems_MaterialType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
			MaterialType: &invalidMaterialType,
			DiscountName: &discountName,
		}),
	}
	invalidMapBillItems_MaterialType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
			MaterialType: &invalidMaterialType,
			DiscountName: &discountName,
		}),
	}

	var invalidMapBillItems_FeeType = make(map[string]*entities.BillItem, 2)
	invalidMapBillItems_FeeType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_FEE.String(),
			FeeType:      &invalidFeeType,
			DiscountName: &discountName,
		}),
	}
	invalidMapBillItems_FeeType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_FEE.String(),
			FeeType:      &invalidFeeType,
			DiscountName: &discountName,
		}),
	}

	var mapBillItems_MaterialType = make(map[string]*entities.BillItem, 2)
	mapBillItems_MaterialType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
			MaterialType: &materialTypeOfBillItem,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			Status: pgtype.Null,
		},
	}
	mapBillItems_MaterialType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_MATERIAL.String(),
			MaterialType: &materialTypeOfBillItem,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			Status: pgtype.Null,
		},
	}

	var mapBillItems_FeeType = make(map[string]*entities.BillItem, 2)
	mapBillItems_FeeType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_FEE.String(),
			FeeType:      &feeTypeOfBillItem,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			Status: pgtype.Null,
		},
	}
	mapBillItems_FeeType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_FEE.String(),
			FeeType:      &feeTypeOfBillItem,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			Status: pgtype.Null,
		},
	}

	var mapBillItems_PackageType = make(map[string]*entities.BillItem, 2)
	mapBillItems_PackageType["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
			PackageType:  &packageTypeOfBillItem,
			QuantityType: &quantityType,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			Status: pgtype.Null,
		},
	}
	mapBillItems_PackageType["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
			PackageType:  &packageTypeOfBillItem,
			QuantityType: &quantityType,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			Status: pgtype.Null,
		},
	}

	material_nonUpcomingBilling := entities.Material{
		MaterialID: pgtype.Text{
			String: "material_1",
			Status: pgtype.Present,
		},
		CustomBillingDate: pgtype.Timestamptz{
			Time: database.Timestamptz(time.Date(2022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
		},
	}

	material_UpcomingBilling := entities.Material{
		MaterialID: pgtype.Text{
			String: "material_1",
			Status: pgtype.Present,
		},
		CustomBillingDate: pgtype.Timestamptz{
			Time: database.Timestamptz(time.Date(3022, 2, 3, 4, 5, 6, 7, time.Local)).Time,
		},
	}

	var invalidMapBillItems_FeeType_UpcomingBillItem = make(map[string]*entities.BillItem, 2)
	invalidMapBillItems_FeeType_UpcomingBillItem["student_product_id_1"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_1",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_FEE.String(),
			FeeType:      &feeTypeOfBillItem,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillTo: pgtype.Timestamptz{
			Time:   time.Time{},
			Status: pgtype.Present,
		},
	}
	invalidMapBillItems_FeeType_UpcomingBillItem["student_product_id_2"] = &entities.BillItem{
		StudentProductID: pgtype.Text{
			String: "student_product_id_2",
			Status: pgtype.Present,
		},
		DiscountID: pgtype.Text{
			String: "1",
			Status: pgtype.Present,
		},
		BillingItemDescription: database.JSONB(&entities.BillingItemDescription{
			ProductID:    "10",
			ProductName:  constant.ProductName,
			ProductType:  pb.ProductType_PRODUCT_TYPE_FEE.String(),
			FeeType:      &feeTypeOfBillItem,
			DiscountName: &discountName,
		}),
		BillSchedulePeriodID: pgtype.Text{
			String: "2",
			Status: pgtype.Present,
		},
		BillTo: pgtype.Timestamptz{
			Time:   time.Time{},
			Status: pgtype.Present,
		},
	}

	studentProductResp := []*entities.StudentProduct{
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
				Status: pgtype.Present,
			},
		},
	}

	studentProductWithCancelStatusResp := []*entities.StudentProduct{
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_1",
				Status: pgtype.Present,
			},
			ProductStatus: pgtype.Text{
				String: pb.StudentProductStatus_CANCELLED.String(),
				Status: pgtype.Present,
			},
		},
		{
			StudentProductID: pgtype.Text{
				String: "student_product_id_2",
				Status: pgtype.Present,
			},
			ProductStatus: pgtype.Text{
				String: pb.StudentProductStatus_CANCELLED.String(),
				Status: pgtype.Present,
			},
		},
	}

	TestCases := []utils.TestCase{
		{
			Name: "Fail case: with nil paging",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
			},
			ExpectedErr: fmt.Errorf("invalid paging data with error"),
			Setup: func(ctx context.Context) {
				// Do nothing
			},
		},
		{
			Name: "Fail case: getting student product by student id",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging:           &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{}, []*entities.StudentProduct{
					{
						StudentID: pgtype.Text{
							String: "student_id",
						},
					},
				}, 0, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: present and future billInfo by student product id",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging:           &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, []*entities.StudentProduct{
					{
						StudentID: pgtype.Text{
							String: "student_id",
						},
					},
				}, 0, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(map[string]*entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: get past billInfo by student product id",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging:           &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, []*entities.StudentProduct{
					{
						StudentID: pgtype.Text{
							String: "student_id",
						},
					},
				}, 0, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(mapBillItems, nil)
				billItemService.On("GetMapPastBillItemInfo",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(map[string]*entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Fail when get upcoming biling",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging:           &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(invalidMapBillItems_FeeType_UpcomingBillItem, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]*entities.BillItem{}, nil)
				billItemService.On("GetUpcomingBilling", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Fail when get total associated product of package",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging:           &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(invalidMapBillItems_FeeType_UpcomingBillItem, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]*entities.BillItem{}, nil)
				billItemService.On("GetUpcomingBilling", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.BillItem{}, nil)
				packageService.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), constant.ErrDefault)
			},
		},
		{
			Name: "happy case package type upcoming billing",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 2,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(invalidMapBillItems_FeeType_UpcomingBillItem, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]*entities.BillItem{}, nil)
				billItemService.On("GetUpcomingBilling", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.BillItem{}, nil)
				packageService.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
			},
		},
		{
			Name: "Fail case: when get material",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 2,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				materialService.On("GetMaterialByID", ctx, mock.Anything, mock.Anything).Return(entities.Material{}, constant.ErrDefault)
			},
		},
		{
			Name: "happy case material type non upcoming billing",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 2,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				materialService.On("GetMaterialByID", ctx, mock.Anything, mock.Anything).Return(material_nonUpcomingBilling, nil)
				packageService.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
			},
		},
		{
			Name: "happy case material type upcoming billing but status cancel",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 2,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductWithCancelStatusResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				packageService.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
			},
		},
		{
			Name: "happy case material type upcoming billing",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 2,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_MaterialType, nil)
				materialService.On("GetMaterialByID", ctx, mock.Anything, mock.Anything).Return(material_UpcomingBilling, nil)
				packageService.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
			},
		},
		{
			Name: "happy case fee type",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging: &cpb.Paging{
					Limit: 2,
					Offset: &cpb.Paging_OffsetInteger{
						OffsetInteger: 0,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, []*entities.StudentProduct{}, 0, nil)
				packageService.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int32(0), nil)
			},
		},
		{
			Name: "Fail case: getting pagination",
			Ctx:  interceptors.ContextWithUserID(Ctx, constant.UserID),
			Req: &pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest{
				StudentProductId: "1",
				Paging:           &cpb.Paging{Limit: 10, Offset: &cpb.Paging_OffsetInteger{OffsetInteger: 10}},
			},
			ExpectedErr: status.Error(codes.Internal, "Error offset"),
			Setup: func(ctx context.Context) {
				studentProductService.On("GetStudentAssociatedProductByStudentProductID",
					ctx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return([]string{"student_product_id_1", "student_product_id_2"}, studentProductResp, 2, nil)
				billItemService.On("GetMapPresentAndFutureBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_FeeType, nil)
				billItemService.On("GetMapPastBillItemInfo", ctx, mock.Anything, mock.Anything, mock.Anything).Return(mapBillItems_FeeType, nil)
				billItemService.On("GetUpcomingBilling", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&entities.BillItem{}, nil)
				packageService.On("GetTotalAssociatedPackageWithCourseIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(int32(1), nil)
			},
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			studentProductService = new(mockServices.IStudentProductForStudentBilling)
			billItemService = new(mockServices.IBillItemServiceForStudentBilling)
			materialService = new(mockServices.IMaterialServiceForStudentBilling)
			orderService = new(mockServices.IOrderServiceForStudentBilling)
			packageService = new(mockServices.IPackageServiceForStudentBilling)

			testCase.Setup(testCase.Ctx)
			s := &StudentBilling{
				DB:                    db,
				StudentProductService: studentProductService,
				BillItemService:       billItemService,
				MaterialService:       materialService,
				OrderService:          orderService,
				PackageService:        packageService,
			}

			resp, err := s.RetrieveListOfOrderAssociatedProductOfPackages(testCase.Ctx, testCase.Req.(*pb.RetrieveListOfOrderAssociatedProductOfPackagesRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
				assert.Nil(t, resp)
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, db, studentProductService, billItemService, orderService, materialService, packageService)
		})
	}
}
