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
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateOrderService_hasEnrolledPriceByProductID(t *testing.T) {
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
			Name:        "Fail case: Error when getting product prices by product id and price type",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         constant.ProductID,
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, constant.ErrDefault)
			},
		},
		{
			Name:         "Happy case (hasEnrolledPrice == true)",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:          constant.ProductID,
			ExpectedErr:  nil,
			ExpectedResp: true,
			Setup: func(ctx context.Context) {
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{
					{
						PriceType: pgtype.Text{String: pb.ProductPriceType_ENROLLED_PRICE.String()},
					},
				}, nil)
			},
		},
		{
			Name:         "Happy case (hasEnrolledPrice == false)",
			Ctx:          interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:          constant.ProductID,
			ExpectedErr:  nil,
			ExpectedResp: false,
			Setup: func(ctx context.Context) {
				productPriceService.On("GetProductPricesByProductIDAndPriceType", ctx, tx, mock.Anything, mock.Anything).Return([]entities.ProductPrice{}, nil)
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

			productID := testCase.Req.(string)
			hasEnrolledPrice, err := s.hasEnrolledPriceByProductID(testCase.Ctx, tx, productID)
			if err != nil {
				fmt.Println(err)
			}

			if testCase.ExpectedErr != nil {
				assert.NotNil(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.Equal(t, testCase.ExpectedErr, err)
				assert.Equal(t, hasEnrolledPrice, testCase.ExpectedResp.(bool))
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
