package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/kafka/payload"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestCreateOrderService_CreateOrder(t *testing.T) {
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
		userService           *mockServices.IUserServiceForCreateOrder
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get student and name by id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderItems: []*pb.OrderItem{
					{
						ProductId:  constant.ProductID,
						DiscountId: &wrapperspb.StringValue{Value: "1"},
					},
				},
				BillingItems: []*pb.BillingItem{
					{
						ProductId: constant.ProductID,
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          "1",
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: 10,
							DiscountAmount:      50,
						},
						FinalPrice: 450,
					},
					{
						ProductId: "2",
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						FinalPrice: 500,
					},
				},
				StudentDetailPath: &wrapperspb.StringValue{
					Value: "12",
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when get location name by id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderItems: []*pb.OrderItem{
					{
						ProductId:  constant.ProductID,
						DiscountId: &wrapperspb.StringValue{Value: "1"},
					},
				},
				BillingItems: []*pb.BillingItem{
					{
						ProductId: constant.ProductID,
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          "1",
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: 10,
							DiscountAmount:      50,
						},
						FinalPrice: 450,
					},
					{
						ProductId: "2",
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						FinalPrice: 500,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(mock.Anything, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when create order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderItems: []*pb.OrderItem{
					{
						ProductId:  constant.ProductID,
						DiscountId: &wrapperspb.StringValue{Value: "1"},
					},
				},
				BillingItems: []*pb.BillingItem{
					{
						ProductId: constant.ProductID,
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          "1",
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: 10,
							DiscountAmount:      50,
						},
						FinalPrice: 450,
					},
					{
						ProductId: "2",
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						FinalPrice: 500,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(mock.Anything, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{}, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when create order items",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderType:    pb.OrderType_ORDER_TYPE_NEW,
				OrderItems: []*pb.OrderItem{
					{
						ProductId:  constant.ProductID,
						DiscountId: &wrapperspb.StringValue{Value: "1"},
					},
				},
				BillingItems: []*pb.BillingItem{
					{
						ProductId: constant.ProductID,
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          "1",
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: 10,
							DiscountAmount:      50,
						},
						FinalPrice: 450,
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(constant.LocationName, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
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
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
				}, nil)
				productService.
					On("VerifiedProductWithStudentInfoReturnProductInfoAndBillingType", ctx, tx, mock.Anything).
					Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, "grade_name", entities.ProductSetting{}, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Happy case: Create order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderType:    pb.OrderType_ORDER_TYPE_LOA,
				OrderItems: []*pb.OrderItem{
					{
						ProductId:  constant.ProductID,
						DiscountId: &wrapperspb.StringValue{Value: "1"},
					},
				},
				BillingItems: []*pb.BillingItem{
					{
						ProductId: constant.ProductID,
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          "1",
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: 10,
							DiscountAmount:      50,
						},
						FinalPrice: 450,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(constant.LocationName, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
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
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
				}, nil)
				subscriptionService.On("ToNotificationMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&payload.UpsertSystemNotification{}, nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("ValidateProductSettingForLOAOrder", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("MutationStudentProductForLOAOrder", ctx, tx, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderLOA", ctx, tx, mock.Anything).Return(nil)
				subscriptionService.On("Publish", ctx, mock.Anything, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "fail case: with SubscriptionService",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderType:    pb.OrderType_ORDER_TYPE_LOA,
				OrderItems: []*pb.OrderItem{
					{
						ProductId:  constant.ProductID,
						DiscountId: &wrapperspb.StringValue{Value: "1"},
					},
				},
				BillingItems: []*pb.BillingItem{
					{
						ProductId: constant.ProductID,
						Price:     500,
						TaxItem: &pb.TaxBillItem{
							TaxId:         "1",
							TaxPercentage: 10,
							TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
							TaxAmount:     50,
						},
						DiscountItem: &pb.DiscountBillItem{
							DiscountId:          "1",
							DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
							DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
							DiscountAmountValue: 10,
							DiscountAmount:      50,
						},
						FinalPrice: 450,
					},
				},
			},
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(constant.LocationName, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
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
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
				}, nil)
				subscriptionService.On("ToNotificationMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&payload.UpsertSystemNotification{}, nil)
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("ValidateProductSettingForLOAOrder", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("MutationStudentProductForLOAOrder", ctx, tx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderLOA", ctx, tx, mock.Anything).Return(nil)
				subscriptionService.On("Publish", ctx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				// elasticSearchService.On("InsertOrderData", ctx, mock.Anything).Return(nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Happy case: Create order withdrawal without product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderType:    pb.OrderType_ORDER_TYPE_WITHDRAWAL,
				OrderItems:   []*pb.OrderItem{},
				BillingItems: []*pb.BillingItem{},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(constant.LocationName, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
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
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
				}, nil)
				subscriptionService.On("ToNotificationMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&payload.UpsertSystemNotification{}, nil)
				subscriptionService.On("Publish", ctx, mock.Anything, mock.Anything).Return(nil)
				// elasticSearchService.On("InsertOrderData", ctx, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Create order new without product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateOrderRequest{
				StudentId:    constant.StudentID,
				LocationId:   constant.LocationID,
				OrderComment: constant.OrderComment,
				OrderType:    pb.OrderType_ORDER_TYPE_NEW,
				OrderItems:   []*pb.OrderItem{},
				BillingItems: []*pb.BillingItem{},
			},
			ExpectedErr: utils.StatusErrWithDetail(
				codes.FailedPrecondition,
				constant.UpdateLikeOrdersMissingBillItem,
				nil,
			),
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(constant.LocationName, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
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
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
				}, nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
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
			userService = new(mockServices.IUserServiceForCreateOrder)
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
				ProductPriceService:   productPriceService,
				UserService:           userService,
			}

			resp, err := s.CreateOrder(testCase.Ctx, testCase.Req.(*pb.CreateOrderRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
				assert.Equal(t, resp.Successful, true)
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

func TestCreateOrderService_CreateBulkOrder(t *testing.T) {
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
	)

	testcases := []utils.TestCase{
		{
			Name: "Happy case: Create order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateBulkOrderRequest{
				NewOrderRequests: []*pb.CreateBulkOrderRequest_CreateNewOrderRequest{
					{
						StudentId:    constant.StudentID,
						LocationId:   constant.LocationID,
						OrderComment: constant.OrderComment,
						OrderType:    pb.OrderType_ORDER_TYPE_NEW,
						OrderItems: []*pb.OrderItem{
							{
								ProductId:  constant.ProductID,
								DiscountId: &wrapperspb.StringValue{Value: "1"},
							},
						},
						BillingItems: []*pb.BillingItem{
							{
								ProductId: constant.ProductID,
								Price:     500,
								TaxItem: &pb.TaxBillItem{
									TaxId:         "1",
									TaxPercentage: 10,
									TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
									TaxAmount:     50,
								},
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          "1",
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: 10,
									DiscountAmount:      50,
								},
								FinalPrice: 450,
							},
						},
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(constant.LocationName, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
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
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
				}, nil)
				productService.On("VerifiedProductWithStudentInfoReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, "", entities.ProductSetting{}, nil)
				studentService.On("IsEnrolledInLocation", ctx, tx, mock.Anything).Return(true, nil)
				studentProductService.On("ValidateProductSettingForCreateOrder", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateStudentProduct", ctx, tx, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				studentService.On("IsEnrolledInOrg", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderCreate", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProduct", ctx, tx, mock.Anything, mock.Anything).Return(nil)
				subscriptionService.On("Publish", ctx, mock.Anything, mock.Anything).Return(nil)
				// elasticSearchService.On("InsertOrderData", ctx, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Wrong order type",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateBulkOrderRequest{
				NewOrderRequests: []*pb.CreateBulkOrderRequest_CreateNewOrderRequest{
					{
						StudentId:    constant.StudentID,
						LocationId:   constant.LocationID,
						OrderComment: constant.OrderComment,
						OrderType:    pb.OrderType_ORDER_TYPE_UPDATE,
						OrderItems: []*pb.OrderItem{
							{
								ProductId:  constant.ProductID,
								DiscountId: &wrapperspb.StringValue{Value: "1"},
							},
						},
						BillingItems: []*pb.BillingItem{
							{
								ProductId: constant.ProductID,
								Price:     500,
								TaxItem: &pb.TaxBillItem{
									TaxId:         "1",
									TaxPercentage: 10,
									TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
									TaxAmount:     50,
								},
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          "1",
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: 10,
									DiscountAmount:      50,
								},
								FinalPrice: 450,
							},
						},
					},
				},
			},
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "we don't support this order type in bulk order"),
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Error create order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateBulkOrderRequest{
				NewOrderRequests: []*pb.CreateBulkOrderRequest_CreateNewOrderRequest{
					{
						StudentId:    constant.StudentID,
						LocationId:   constant.LocationID,
						OrderComment: constant.OrderComment,
						OrderType:    pb.OrderType_ORDER_TYPE_NEW,
						OrderItems: []*pb.OrderItem{
							{
								ProductId:  constant.ProductID,
								DiscountId: &wrapperspb.StringValue{Value: "1"},
							},
						},
						BillingItems: []*pb.BillingItem{
							{
								ProductId: constant.ProductID,
								Price:     500,
								TaxItem: &pb.TaxBillItem{
									TaxId:         "1",
									TaxPercentage: 10,
									TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
									TaxAmount:     50,
								},
								DiscountItem: &pb.DiscountBillItem{
									DiscountId:          "1",
									DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
									DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
									DiscountAmountValue: 10,
									DiscountAmount:      50,
								},
								FinalPrice: 450,
							},
						},
					},
				},
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when get student and name by id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateBulkOrderRequest{
				NewOrderRequests: []*pb.CreateBulkOrderRequest_CreateNewOrderRequest{{
					StudentId:    constant.StudentID,
					LocationId:   constant.LocationID,
					OrderComment: constant.OrderComment,
					OrderItems: []*pb.OrderItem{
						{
							ProductId:  constant.ProductID,
							DiscountId: &wrapperspb.StringValue{Value: "1"},
						},
					},
					BillingItems: []*pb.BillingItem{
						{
							ProductId: constant.ProductID,
							Price:     500,
							TaxItem: &pb.TaxBillItem{
								TaxId:         "1",
								TaxPercentage: 10,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxAmount:     50,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "1",
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: 10,
								DiscountAmount:      50,
							},
							FinalPrice: 450,
						},
						{
							ProductId: "2",
							Price:     500,
							TaxItem: &pb.TaxBillItem{
								TaxId:         "1",
								TaxPercentage: 10,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxAmount:     50,
							},
							FinalPrice: 500,
						},
					},
				}}},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when get location name by id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateBulkOrderRequest{
				NewOrderRequests: []*pb.CreateBulkOrderRequest_CreateNewOrderRequest{{
					StudentId:    constant.StudentID,
					LocationId:   constant.LocationID,
					OrderComment: constant.OrderComment,
					OrderItems: []*pb.OrderItem{
						{
							ProductId:  constant.ProductID,
							DiscountId: &wrapperspb.StringValue{Value: "1"},
						},
					},
					BillingItems: []*pb.BillingItem{
						{
							ProductId: constant.ProductID,
							Price:     500,
							TaxItem: &pb.TaxBillItem{
								TaxId:         "1",
								TaxPercentage: 10,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxAmount:     50,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "1",
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: 10,
								DiscountAmount:      50,
							},
							FinalPrice: 450,
						},
						{
							ProductId: "2",
							Price:     500,
							TaxItem: &pb.TaxBillItem{
								TaxId:         "1",
								TaxPercentage: 10,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxAmount:     50,
							},
							FinalPrice: 500,
						},
					},
				}}},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(mock.Anything, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when create order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateBulkOrderRequest{
				NewOrderRequests: []*pb.CreateBulkOrderRequest_CreateNewOrderRequest{{
					StudentId:    constant.StudentID,
					LocationId:   constant.LocationID,
					OrderComment: constant.OrderComment,
					OrderItems: []*pb.OrderItem{
						{
							ProductId:  constant.ProductID,
							DiscountId: &wrapperspb.StringValue{Value: "1"},
						},
					},
					BillingItems: []*pb.BillingItem{
						{
							ProductId: constant.ProductID,
							Price:     500,
							TaxItem: &pb.TaxBillItem{
								TaxId:         "1",
								TaxPercentage: 10,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxAmount:     50,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "1",
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: 10,
								DiscountAmount:      50,
							},
							FinalPrice: 450,
						},
						{
							ProductId: "2",
							Price:     500,
							TaxItem: &pb.TaxBillItem{
								TaxId:         "1",
								TaxPercentage: 10,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxAmount:     50,
							},
							FinalPrice: 500,
						},
					},
				}}},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(mock.Anything, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{}, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when create order items",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateBulkOrderRequest{
				NewOrderRequests: []*pb.CreateBulkOrderRequest_CreateNewOrderRequest{{
					StudentId:    constant.StudentID,
					LocationId:   constant.LocationID,
					OrderComment: constant.OrderComment,
					OrderType:    pb.OrderType_ORDER_TYPE_NEW,
					OrderItems: []*pb.OrderItem{
						{
							ProductId:  constant.ProductID,
							DiscountId: &wrapperspb.StringValue{Value: "1"},
						},
					},
					BillingItems: []*pb.BillingItem{
						{
							ProductId: constant.ProductID,
							Price:     500,
							TaxItem: &pb.TaxBillItem{
								TaxId:         "1",
								TaxPercentage: 10,
								TaxCategory:   pb.TaxCategory_TAX_CATEGORY_INCLUSIVE,
								TaxAmount:     50,
							},
							DiscountItem: &pb.DiscountBillItem{
								DiscountId:          "1",
								DiscountType:        pb.DiscountType_DISCOUNT_TYPE_REGULAR,
								DiscountAmountType:  pb.DiscountAmountType_DISCOUNT_AMOUNT_TYPE_PERCENTAGE,
								DiscountAmountValue: 10,
								DiscountAmount:      50,
							},
							FinalPrice: 450,
						},
					},
				}}},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, mock.Anything, mock.Anything).Return(entities.Student{
					StudentID: pgtype.Text{
						String: constant.StudentID,
					},
				}, constant.StudentName, nil)
				locationService.On("GetLocationNameByID", ctx, mock.Anything, mock.Anything).Return(constant.LocationName, nil)
				orderService.On("CreateOrder", ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.Order{
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
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
					},
				}, nil)
				productService.
					On("VerifiedProductWithStudentInfoReturnProductInfoAndBillingType", ctx, tx, mock.Anything).
					Return(entities.Product{}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, "grade_name", entities.ProductSetting{}, constant.ErrDefault)

				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
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
				ProductPriceService:   productPriceService,
			}

			resp, err := s.CreateBulkOrder(testCase.Ctx, testCase.Req.(*pb.CreateBulkOrderRequest))
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.NotNil(t, resp)
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

func TestCreateOrderService_CreateOrderItems(t *testing.T) {
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
			Name: "Fail case: Error when create order items for create",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderCreate,
				map[string]utils.OrderItemData{
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
								String: pb.OrderType_ORDER_TYPE_NEW.String(),
								Status: 2,
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
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productService.On("VerifiedProductWithStudentInfoReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, "grade_name", entities.ProductSetting{}, nil)
				studentProductService.On("CreateStudentProduct", ctx, tx, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				studentService.On("IsEnrolledInLocation", ctx, tx, mock.Anything).Return(true, nil)
				studentProductService.On("ValidateProductSettingForCreateOrder", ctx, tx, mock.Anything).Return(nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				studentService.On("IsEnrolledInOrg", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderCreate", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProduct", ctx, tx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentProductService.On("CreateAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				studentPackageService.On("MutationStudentPackageForCreateOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Happy case: Create new order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderCreate,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_NEW.String()},
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
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				productService.On("VerifiedProductWithStudentInfoReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, "grade_name", entities.ProductSetting{}, nil)
				studentService.On("IsEnrolledInLocation", ctx, tx, mock.Anything).Return(true, nil)
				studentProductService.On("ValidateProductSettingForCreateOrder", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateStudentProduct", ctx, tx, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				studentService.On("IsEnrolledInOrg", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderCreate", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProduct", ctx, tx, mock.Anything, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				studentPackageService.On("MutationStudentPackageForCreateOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Fail case: Error when create order items for cancel",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderCancel,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_NEW.String()},
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
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForCancelOrder", ctx, tx, mock.Anything).Return(entities.StudentProduct{}, entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
				billingService.On("CreateBillItemForOrderCancel", ctx, tx, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case: Create cancel order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderCancel,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
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
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForCancelOrder", ctx, tx, mock.Anything).Return(entities.StudentProduct{}, entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderCancel", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Fail case: Error when create order items for enrollment",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderEnrollment,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_ENROLLMENT.String()},
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
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductWithStudentInfoReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, "grade_name", entities.ProductSetting{}, nil)
				studentProductService.On("CreateStudentProduct", ctx, tx, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				billingService.On("CreateBillItemForOrderCreate", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProduct", ctx, tx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				studentPackageService.On("MutationStudentPackageForCreateOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Happy case: Create enrollment order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderEnrollment,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_ENROLLMENT.String()},
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
			},
			Setup: func(ctx context.Context) {
				studentService.On("ValidateStudentStatusForOrderType", ctx, tx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				productService.On("VerifiedProductWithStudentInfoReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductID: pgtype.Text{
						String: constant.ProductID,
					},
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, "grade_name", entities.ProductSetting{}, nil)
				studentProductService.On("CreateStudentProduct", ctx, tx, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
				billingService.On("CreateBillItemForOrderCreate", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("CreateAssociatedStudentProduct", ctx, tx, mock.Anything, mock.Anything).Return(nil)
				studentPackageService.On("MutationStudentPackageForCreateOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Fail case: Error when create order items for graduate",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderGraduate,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_GRADUATE.String()},
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
				studentProductService.On("MutationStudentProductForGraduateOrder", ctx, tx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderGraduate", ctx, tx, mock.Anything).Return(constant.ErrDefault)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Happy case: Create graduate order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderGraduate,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_GRADUATE.String()},
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
				studentProductService.On("MutationStudentProductForGraduateOrder", ctx, tx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderGraduate", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("DeleteAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Fail case: Error when create order items for update",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderUpdate,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_UPDATE.String()},
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
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productService.On("VerifiedProductReturnProductInfoAndBillingType", ctx, tx, mock.Anything).Return(entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: pgtype.Present,
					},
				}, true, true, pb.ProductType_PRODUCT_TYPE_MATERIAL, entities.ProductSetting{}, nil)
				studentProductService.On("MutationStudentProductForUpdateOrder", ctx, tx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderUpdate", ctx, tx, mock.Anything).Return(constant.ErrDefault)
				studentPackageService.On("MutationStudentPackageForUpdateOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Fail case: Error when create order items for withdraw",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderWithdraw,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String()},
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
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderWithdrawal", ctx, tx, mock.Anything).Return(constant.ErrDefault)
				studentPackageService.On("MutationStudentPackageForCancelOrder", ctx, tx, mock.Anything).Return([]*npb.EventStudentPackage{{}}, nil)
			},
		},
		{
			Name: "Happy case: Create withdrawal order successfully",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: []interface{}{
				utils.OrderWithdraw,
				map[string]utils.OrderItemData{
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
							OrderType: pgtype.Text{String: pb.OrderType_ORDER_TYPE_WITHDRAWAL.String()},
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
				studentProductService.On("MutationStudentProductForWithdrawalOrder", ctx, tx, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, nil)
				orderItemService.On("CreateOrderItem", ctx, tx, mock.Anything).Return(entities.OrderItem{}, nil)
				packageService.On("VerifyPackageDataAndUpsertRelateData", ctx, tx, mock.Anything).Return(utils.PackageInfo{}, nil)
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{{}}, nil)
				studentService.On("CheckIsEnrolledInOrgByStudentIDAndTime", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(true, nil)
				billingService.On("CreateBillItemForOrderWithdrawal", ctx, tx, mock.Anything).Return(nil)
				studentProductService.On("DeleteAssociatedStudentProductByAssociatedStudentProductID", ctx, tx, mock.Anything).Return(nil)
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

			orderTypeReq := testCase.Req.([]interface{})[0].(utils.OrderType)
			mapKeyWithOrderItemDataReq := testCase.Req.([]interface{})[1].(map[string]utils.OrderItemData)
			message, elasticData, err := s.CreateOrderItems(testCase.Ctx, tx, &orderTypeReq, mapKeyWithOrderItemDataReq)

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
				studentPackageService,
			)
		})
	}
}
