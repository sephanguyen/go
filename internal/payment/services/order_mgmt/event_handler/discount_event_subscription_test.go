package eventhandler

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	discountEntities "github.com/manabie-com/backend/internal/discount/entities"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockNats "github.com/manabie-com/backend/mock/golibs/nats"
	mockRepositories "github.com/manabie-com/backend/mock/payment/repositories"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestDiscountEventSubscription_Subscribe(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		jsm *mockNats.JetStreamManagement
	)

	testcases := []utils.TestCase{
		{
			Name:        "Happy case",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&nats.Subscription{}, nil)
			},
		},
		{
			Name:        "Fail case: error parsing data",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				jsm.On("QueueSubscribe", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&nats.Subscription{}, constant.ErrDefault)
			},
		},
	}

	for _, testCase := range testcases {
		jsm = new(mockNats.JetStreamManagement)

		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			s := &DiscountEventSubscription{
				JSM: jsm,
			}

			err := s.Subscribe()
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, jsm)
		})
	}
}

func TestDiscountEventSubscription_CreateUpdateOrderRequest(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		orderItemRepo                      *mockRepositories.MockOrderItemRepo
		orderItemCourseRepo                *mockRepositories.MockOrderItemCourseRepo
		productRepo                        *mockRepositories.MockProductRepo
		billingSchedulePeriodRepo          *mockRepositories.MockBillingSchedulePeriodRepo
		billItemRepo                       *mockRepositories.MockBillItemRepo
		productPriceRepo                   *mockRepositories.MockProductPriceRepo
		taxRepo                            *mockRepositories.MockTaxRepo
		orderService                       *mockServices.IOrderServiceServiceForDiscountEventSubscription
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: error retrieving product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &discountEntities.UpdateProductDiscount{
				StudentID:    mock.Anything,
				ProductID:    mock.Anything,
				DiscountID:   mock.Anything,
				DiscountType: pb.DiscountType_DISCOUNT_TYPE_FAMILY,
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: no bill item generated for recurring product",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: fmt.Errorf("no bill item"),
			Req: &discountEntities.UpdateProductDiscount{
				StudentID:    mock.Anything,
				ProductID:    mock.Anything,
				DiscountID:   mock.Anything,
				DiscountType: pb.DiscountType_DISCOUNT_TYPE_FAMILY,
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(), Status: pgtype.Present},
				}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{}, nil)
			},
		},
		{
			Name:        "Fail case: error creating order for recurring material",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: &discountEntities.UpdateProductDiscount{
				StudentID:             mock.Anything,
				LocationID:            mock.Anything,
				ProductID:             mock.Anything,
				StudentProductID:      mock.Anything,
				DiscountID:            mock.Anything,
				DiscountType:          pb.DiscountType_DISCOUNT_TYPE_FAMILY,
				EffectiveDate:         time.Now(),
				StudentProductEndDate: time.Now().AddDate(0, 0, 10),
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductType:          pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(), Status: pgtype.Present},
					DisableProRatingFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{Int: big.NewInt(100), Status: pgtype.Present},
				}, nil)
				taxRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Tax{
					TaxID:         pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					TaxPercentage: pgtype.Int4{Int: 20, Status: pgtype.Present},
					TaxCategory:   pgtype.Text{String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(), Status: pgtype.Present},
				}, nil)
				orderService.On("CreateOrder", ctx, mock.Anything).Return(&pb.CreateOrderResponse{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: recurring material",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: &discountEntities.UpdateProductDiscount{
				StudentID:             mock.Anything,
				LocationID:            mock.Anything,
				ProductID:             mock.Anything,
				StudentProductID:      mock.Anything,
				DiscountID:            mock.Anything,
				DiscountType:          pb.DiscountType_DISCOUNT_TYPE_FAMILY,
				EffectiveDate:         time.Now(),
				StudentProductEndDate: time.Now().AddDate(0, 0, 10),
			},
			Setup: func(ctx context.Context) {
				productRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Product{
					ProductType:          pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(), Status: pgtype.Present},
					DisableProRatingFlag: pgtype.Bool{Bool: true, Status: pgtype.Present},
				}, nil)
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{
					Price: pgtype.Numeric{Int: big.NewInt(100), Status: pgtype.Present},
				}, nil)
				taxRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Tax{
					TaxID:         pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					TaxPercentage: pgtype.Int4{Int: 20, Status: pgtype.Present},
					TaxCategory:   pgtype.Text{String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(), Status: pgtype.Present},
				}, nil)
				orderService.On("CreateOrder", ctx, mock.Anything).Return(&pb.CreateOrderResponse{
					OrderId: mock.Anything,
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		orderItemRepo = new(mockRepositories.MockOrderItemRepo)
		orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
		productRepo = new(mockRepositories.MockProductRepo)
		billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
		billItemRepo = new(mockRepositories.MockBillItemRepo)
		productPriceRepo = new(mockRepositories.MockProductPriceRepo)
		taxRepo = new(mockRepositories.MockTaxRepo)
		orderService = new(mockServices.IOrderServiceServiceForDiscountEventSubscription)
		studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)

		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			s := &DiscountEventSubscription{
				Logger:                             zap.NewNop(),
				OrderService:                       orderService,
				OrderItemRepo:                      orderItemRepo,
				OrderItemCourseRepo:                orderItemCourseRepo,
				ProductRepo:                        productRepo,
				BillingSchedulePeriodRepo:          billingSchedulePeriodRepo,
				BillItemRepo:                       billItemRepo,
				ProductPriceRepo:                   productPriceRepo,
				TaxRepo:                            taxRepo,
				StudentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}

			data := testCase.Req.(*discountEntities.UpdateProductDiscount)
			err := s.CreateUpdateOrderRequest(testCase.Ctx, data)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, orderItemRepo, orderItemCourseRepo, productRepo, billingSchedulePeriodRepo, billItemRepo, productPriceRepo, taxRepo)
		})
	}
}

func TestDiscountEventSubscription_GenerateOrderItemForUpdateProductDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		orderItemRepo                  *mockRepositories.MockOrderItemRepo
		orderItemCourseRepo            *mockRepositories.MockOrderItemCourseRepo
		productRepo                    *mockRepositories.MockProductRepo
		packageRepo                    *mockRepositories.MockPackageRepo
		packageQuantityTypeMappingRepo *mockRepositories.MockPackageQuantityTypeMappingRepo
		packageCourseRepo              *mockRepositories.MockPackageCourseRepo
	)

	testcases := []utils.TestCase{
		{
			Name:        "Fail case: error retrieving order item",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error mapping order item course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error retrieving package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				packageRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Package{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error retrieving package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				packageRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Package{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error retrieving quantity type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{}, nil)
				packageRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_FREQUENCY.String(), Status: pgtype.Present}}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", ctx, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK, constant.ErrDefault)
			},
		},
		{
			Name:        "Fail case: error retrieving package course",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: constant.ErrDefault,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{
					mock.Anything: {},
				}, nil)
				packageRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_FREQUENCY.String(), Status: pgtype.Present}}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", ctx, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT, nil)
				packageCourseRepo.On("GetByPackageIDAndCourseID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.PackageCourse{}, constant.ErrDefault)
			},
		},
		{
			Name:        "Happy case: frequency base package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{
					mock.Anything: {},
				}, nil)
				packageRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_FREQUENCY.String(), Status: pgtype.Present}}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", ctx, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_SLOT_PER_WEEK, nil)
			},
		},
		{
			Name:        "Happy case: schedule base package",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
				orderItemCourseRepo.On("GetMapOrderItemCourseByOrderIDAndPackageID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(map[string]entities.OrderItemCourse{
					mock.Anything: {},
				}, nil)
				packageRepo.On("GetByID", ctx, mock.Anything, mock.Anything).Return(entities.Package{
					PackageType: pgtype.Text{String: pb.PackageType_PACKAGE_TYPE_FREQUENCY.String(), Status: pgtype.Present}}, nil)
				packageQuantityTypeMappingRepo.On("GetByPackageTypeForUpdate", ctx, mock.Anything, mock.Anything).Return(pb.QuantityType_QUANTITY_TYPE_COURSE_WEIGHT, nil)
				packageCourseRepo.On("GetByPackageIDAndCourseID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.PackageCourse{
					CourseWeight: pgtype.Int4{Int: 3},
				}, nil)
			},
		},
		{
			Name:        "Happy case: recurring material",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			ExpectedErr: nil,
			Req: entities.Product{
				ProductType: pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(), Status: pgtype.Present},
			},
			Setup: func(ctx context.Context) {
				orderItemRepo.On("GetOrderItemByStudentProductID", ctx, mock.Anything, mock.Anything).Return(entities.OrderItem{}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		orderItemRepo = new(mockRepositories.MockOrderItemRepo)
		orderItemCourseRepo = new(mockRepositories.MockOrderItemCourseRepo)
		packageRepo = new(mockRepositories.MockPackageRepo)
		packageQuantityTypeMappingRepo = new(mockRepositories.MockPackageQuantityTypeMappingRepo)
		packageCourseRepo = new(mockRepositories.MockPackageCourseRepo)

		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			s := &DiscountEventSubscription{
				OrderItemRepo:           orderItemRepo,
				OrderItemCourseRepo:     orderItemCourseRepo,
				ProductRepo:             productRepo,
				PackageRepo:             packageRepo,
				PackageQuantityTypeRepo: packageQuantityTypeMappingRepo,
				PackageCourseRepo:       packageCourseRepo,
			}

			product := testCase.Req.(entities.Product)
			resp, _, _, err := s.GenerateOrderItemForUpdateProductDiscount(testCase.Ctx, discountEntities.UpdateProductDiscount{}, product)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
			}

			mock.AssertExpectationsForObjects(t, orderItemRepo, orderItemCourseRepo, packageRepo, packageQuantityTypeMappingRepo, packageCourseRepo)
		})
	}
}

func TestDiscountEventSubscription_GenerateBillItemForUpdateProductDiscount(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		billItemRepo                       *mockRepositories.MockBillItemRepo
		billingRatioRepo                   *mockRepositories.MockBillingRatioRepo
		billingSchedulePeriodRepo          *mockRepositories.MockBillingSchedulePeriodRepo
		productPriceRepo                   *mockRepositories.MockProductPriceRepo
		taxRepo                            *mockRepositories.MockTaxRepo
		studentEnrollmentStatusHistoryRepo *mockRepositories.MockStudentEnrollmentStatusHistoryRepo
	)

	updateProductDiscountData := discountEntities.UpdateProductDiscount{
		ProductID:             mock.Anything,
		StudentProductID:      mock.Anything,
		DiscountID:            mock.Anything,
		DiscountType:          pb.DiscountType_DISCOUNT_TYPE_FAMILY,
		DiscountAmountType:    pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
		DiscountAmountValue:   10,
		EffectiveDate:         time.Now(),
		StudentProductEndDate: time.Now().AddDate(0, 2, 0),
	}

	materialProduct := entities.Product{
		DisableProRatingFlag: pgtype.Bool{Bool: false, Status: pgtype.Present},
		ProductType:          pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_MATERIAL.String(), Status: pgtype.Present},
	}

	packageProduct := entities.Product{
		DisableProRatingFlag: pgtype.Bool{Bool: false, Status: pgtype.Present},
		ProductType:          pgtype.Text{String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(), Status: pgtype.Present},
	}

	courseItems := []*pb.CourseItem{}

	testcases := []utils.TestCase{
		{
			Name: "Fail case: error on retrieve billing periods",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				updateProductDiscountData,
				materialProduct,
				courseItems,
				pb.QuantityType_QUANTITY_TYPE_NONE,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error on retrieve old bill item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				updateProductDiscountData,
				materialProduct,
				courseItems,
				pb.QuantityType_QUANTITY_TYPE_NONE,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error on retrieve product price",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				updateProductDiscountData,
				materialProduct,
				courseItems,
				pb.QuantityType_QUANTITY_TYPE_NONE,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error on retrieve billing ratio",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				updateProductDiscountData,
				materialProduct,
				courseItems,
				pb.QuantityType_QUANTITY_TYPE_NONE,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: error on retrieve tax",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				updateProductDiscountData,
				materialProduct,
				courseItems,
				pb.QuantityType_QUANTITY_TYPE_NONE,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{}, nil)
				taxRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Tax{}, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: recurring package with enrolled status",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				updateProductDiscountData,
				packageProduct,
				courseItems,
				pb.QuantityType_QUANTITY_TYPE_NONE,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
						BillingDate:             pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, -15), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{
					{
						StudentID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					},
				}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				productPriceRepo.On("GetByProductIDAndQuantityAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{
					BillingRatioNumerator:   pgtype.Int4{Int: 1, Status: pgtype.Present},
					BillingRatioDenominator: pgtype.Int4{Int: 1, Status: pgtype.Present},
				}, nil)
				taxRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Tax{
					TaxID:         pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					TaxPercentage: pgtype.Int4{Int: 20, Status: pgtype.Present},
					TaxCategory:   pgtype.Text{String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(), Status: pgtype.Present},
				}, nil)
			},
		},
		{
			Name: "Happy case: recurring material with enrolled status",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				updateProductDiscountData,
				materialProduct,
				courseItems,
				pb.QuantityType_QUANTITY_TYPE_NONE,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				billingSchedulePeriodRepo.On("GetAllBillingPeriodsByBillingScheduleID", ctx, mock.Anything, mock.Anything).Return([]entities.BillingSchedulePeriod{
					{
						BillingSchedulePeriodID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
						StartDate:               pgtype.Timestamptz{Time: time.Now().AddDate(0, -1, 0), Status: pgtype.Present},
						EndDate:                 pgtype.Timestamptz{Time: time.Now().AddDate(0, 1, 0), Status: pgtype.Present},
						BillingDate:             pgtype.Timestamptz{Time: time.Now().AddDate(0, 0, -15), Status: pgtype.Present},
					},
				}, nil)
				studentEnrollmentStatusHistoryRepo.On("GetListEnrolledStatusByStudentIDAndTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return([]*entities.StudentEnrollmentStatusHistory{
					{
						StudentID: pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					},
				}, nil)
				billItemRepo.On("GetBillItemByStudentProductIDAndPeriodID", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillItem{}, nil)
				productPriceRepo.On("GetByProductIDAndBillingSchedulePeriodIDAndPriceType", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.ProductPrice{}, nil)
				billingRatioRepo.On("GetFirstRatioByBillingSchedulePeriodIDAndFromTime", ctx, mock.Anything, mock.Anything, mock.Anything).Return(entities.BillingRatio{
					BillingRatioNumerator:   pgtype.Int4{Int: 1, Status: pgtype.Present},
					BillingRatioDenominator: pgtype.Int4{Int: 1, Status: pgtype.Present},
				}, nil)
				taxRepo.On("GetByIDForUpdate", ctx, mock.Anything, mock.Anything).Return(entities.Tax{
					TaxID:         pgtype.Text{String: mock.Anything, Status: pgtype.Present},
					TaxPercentage: pgtype.Int4{Int: 20, Status: pgtype.Present},
					TaxCategory:   pgtype.Text{String: pb.TaxCategory_TAX_CATEGORY_INCLUSIVE.String(), Status: pgtype.Present},
				}, nil)
			},
		},
	}

	for _, testCase := range testcases {
		billItemRepo = new(mockRepositories.MockBillItemRepo)
		billingRatioRepo = new(mockRepositories.MockBillingRatioRepo)
		billingSchedulePeriodRepo = new(mockRepositories.MockBillingSchedulePeriodRepo)
		productPriceRepo = new(mockRepositories.MockProductPriceRepo)
		taxRepo = new(mockRepositories.MockTaxRepo)
		studentEnrollmentStatusHistoryRepo = new(mockRepositories.MockStudentEnrollmentStatusHistoryRepo)

		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(testCase.Ctx)
			s := &DiscountEventSubscription{
				BillItemRepo:                       billItemRepo,
				BillingRatioRepo:                   billingRatioRepo,
				BillingSchedulePeriodRepo:          billingSchedulePeriodRepo,
				ProductPriceRepo:                   productPriceRepo,
				TaxRepo:                            taxRepo,
				StudentEnrollmentStatusHistoryRepo: studentEnrollmentStatusHistoryRepo,
			}

			data := testCase.Req.([]interface{})[0].(discountEntities.UpdateProductDiscount)
			product := testCase.Req.([]interface{})[1].(entities.Product)
			courseItems := testCase.Req.([]interface{})[2].([]*pb.CourseItem)
			quantityType := testCase.Req.([]interface{})[3].(pb.QuantityType)
			_, _, err := s.GenerateBillItemsForUpdateProductDiscount(testCase.Ctx, data, product, courseItems, quantityType)
			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
			}

			mock.AssertExpectationsForObjects(t, billItemRepo, billingRatioRepo, billingSchedulePeriodRepo, productPriceRepo, taxRepo, studentEnrollmentStatusHistoryRepo)
		})
	}
}
