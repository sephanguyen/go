package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/entities"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	mockBillingService "github.com/manabie-com/backend/mock/payment/utils"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrderService_OrderItemWithdrawal(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                    *mockDb.Ext
		tx                    *mockDb.Tx
		orderService          *mockServices.IOrderServiceForCreateOrder
		productService        *mockServices.IProductServiceForCreateOrder
		productPriceService   *mockServices.IProductPriceServiceForCreateOrder
		studentService        *mockServices.IStudentServiceForCreateOrder
		billingService        *mockBillingService.IBillingService
		subscriptionService   *mockServices.ISubscriptionServiceForCreateOrder
		locationService       *mockServices.ILocationServiceForCreateOrder
		orderItemService      *mockServices.IOrderItemServiceForCreateOrder
		elasticSearchService  *mockServices.IElasticSearchServiceForCreateOrder
		studentProductService *mockServices.IStudentProductServiceForCreateOrder
		packageService        *mockServices.IPackageServiceForCreateOrder
		studentPackageService *mockServices.IStudentPackageForCreateOrder
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error on checking student status in location",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": utils.OrderItemData{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when verified product return product info and billing type",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": utils.OrderItemData{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.
					On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).
					Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when mutation student product for withdrawal order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": utils.OrderItemData{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when create order item",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": utils.OrderItemData{
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when verify package data and upsert relate data",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": {
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					ProductInfo: entities.Product{
						ProductType: pgtype.Text{
							String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
							Status: pgtype.Present,
						},
					},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, constant.ErrDefault)
			},
		},
		{
			Name: "Fail case: Error when create bill item for order withdrawal",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": {
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					ProductInfo: entities.Product{
						ProductType: pgtype.Text{
							String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
							Status: pgtype.Present,
						},
					},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderWithdrawal", ctx, tx, mock.Anything).Return(constant.ErrDefault)

			},
		},
		{
			Name: "Fail case: Error when get product prices by product id and price type",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": {
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					ProductInfo: entities.Product{
						ProductType: pgtype.Text{
							String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
							Status: pgtype.Present,
						},
					},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, constant.ErrDefault)

			},
		},
		{
			Name: "Fail case: Error when check is enrolled in org by student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": {
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					ProductInfo: entities.Product{
						ProductType: pgtype.Text{
							String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
							Status: pgtype.Present,
						},
					},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: Create withdrawal order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: map[string]utils.OrderItemData{
				"key_1": {
					Order: entities.Order{
						OrderID: pgtype.Text{
							String: constant.OrderID,
						},
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						StudentFullName: pgtype.Text{
							String: constant.StudentName,
						},
						LocationID: pgtype.Text{
							String: constant.LocationID,
						},
						OrderType: pgtype.Text{
							String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String(),
						},
					},
					StudentInfo: entities.Student{
						StudentID: pgtype.Text{
							String: constant.StudentID,
						},
						GradeID: pgtype.Text{
							String: constant.DefaultGrade.String,
						},
					},
					ProductInfo: entities.Product{
						ProductType: pgtype.Text{
							String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
							Status: pgtype.Present,
						},
					},
					PackageInfo:            utils.PackageInfo{},
					StudentProduct:         entities.StudentProduct{},
					StudentName:            constant.StudentName,
					LocationName:           constant.LocationName,
					IsOneTimeProduct:       false,
					IsDisableProRatingFlag: false,
					ProductType:            1,
					OrderItem: &pb.OrderItem{
						ProductId: constant.ProductID,
					},
					BillItems: []utils.BillingItemData{
						{
							BillingItem: &pb.BillingItem{
								ProductId:  constant.ProductID,
								Price:      constant.DefaultPrice,
								FinalPrice: constant.DefaultPrice,
							},
							IsUpcoming: false,
						},
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				studentProductService.On("DeleteAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderWithdrawal", ctx, tx, mock.Anything).Return(nil)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)

			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			orderService = new(mockServices.IOrderServiceForCreateOrder)
			productService = new(mockServices.IProductServiceForCreateOrder)
			productPriceService = new(mockServices.IProductPriceServiceForCreateOrder)
			studentService = new(mockServices.IStudentServiceForCreateOrder)
			billingService = new(mockBillingService.IBillingService)
			subscriptionService = new(mockServices.ISubscriptionServiceForCreateOrder)
			locationService = new(mockServices.ILocationServiceForCreateOrder)
			orderItemService = new(mockServices.IOrderItemServiceForCreateOrder)
			elasticSearchService = new(mockServices.IElasticSearchServiceForCreateOrder)
			studentProductService = new(mockServices.IStudentProductServiceForCreateOrder)
			packageService = new(mockServices.IPackageServiceForCreateOrder)
			studentPackageService = new(mockServices.IStudentPackageForCreateOrder)

			testCase.Setup(testCase.Ctx)
			s := &CreateOrderService{
				DB:                    db,
				OrderService:          orderService,
				ProductService:        productService,
				StudentService:        studentService,
				BillingService:        billingService,
				SubscriptionService:   subscriptionService,
				LocationService:       locationService,
				OrderItemService:      orderItemService,
				ElasticSearchService:  elasticSearchService,
				StudentProductService: studentProductService,
				PackageService:        packageService,
				StudentPackageService: studentPackageService,
				ProductPriceService:   productPriceService,
			}

			message, elasticData, err := s.OrderItemWithdrawal(testCase.Ctx, tx, testCase.Req.(map[string]utils.OrderItemData))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, message)
				assert.NotNil(t, elasticData)
			}

			mock.AssertExpectationsForObjects(
				t,
				db,
				orderService,
				productService,
				studentService,
				billingService,
				subscriptionService,
				locationService,
				orderItemService,
				elasticSearchService,
				studentProductService,
				packageService,
			)
		})
	}
}
