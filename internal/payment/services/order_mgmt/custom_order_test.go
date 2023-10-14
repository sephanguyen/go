package ordermgmt

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
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewCreateCustomOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                  *mockDb.Ext
		tx                  *mockDb.Tx
		orderService        *mockServices.IOrderServiceForCreateCustomOrder
		studentService      *mockServices.IStudentServiceForCreateCustomOrder
		locationService     *mockServices.ILocationServiceForCreateCustomOrder
		elasticService      *mockServices.IElasticSearchServiceForCreateCustomOrder
		orderItemService    *mockServices.IOrderItemServiceForCreateCustomOrder
		billingService      *mockServices.IBillingServiceForCreateCustomOrder
		subscriptionService *mockServices.ISubscriptionServiceForCreateOrder
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when get student name",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateCustomBillingRequest{
				LocationId: "1",
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name:        "Fail case: Error when missing location id",
			Ctx:         interceptors.ContextWithUserID(ctx, constant.UserID),
			Req:         &pb.CreateCustomBillingRequest{},
			ExpectedErr: status.Errorf(codes.FailedPrecondition, "Missing mandatory data: location"),
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", nil)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when call location service",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateCustomBillingRequest{
				LocationId: "1",
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", nil)
				locationService.On("GetLocationNameByID", ctx, tx, mock.Anything).Return("", constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when call order service",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateCustomBillingRequest{
				LocationId: "1",
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", nil)
				locationService.On("GetLocationNameByID", ctx, tx, mock.Anything).Return("", nil)
				orderService.On("CreateCustomOrder",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(entities.Order{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when call order item services",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateCustomBillingRequest{
				LocationId: "1",
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", nil)
				locationService.On("GetLocationNameByID", ctx, tx, mock.Anything).Return("", nil)
				orderService.On("CreateCustomOrder",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(entities.Order{}, nil)
				orderItemService.On("CreateMultiCustomOrderItem",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
				).Return([]entities.OrderItem{{}, {}}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when call billing services",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateCustomBillingRequest{
				LocationId: "1",
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", nil)
				locationService.On("GetLocationNameByID", ctx, tx, mock.Anything).Return("", nil)
				orderService.On("CreateCustomOrder",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(entities.Order{}, nil)
				orderItemService.On("CreateMultiCustomOrderItem",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
				).Return([]entities.OrderItem{{}, {}}, nil)
				billingService.On("CreateBillItemForCustomOrder",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		// {
		// 	Name: "Fail case: Error when call elastic services",
		// 	Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
		// 	Req: &pb.CreateCustomBillingRequest{
		// 		LocationId: "1",
		// 	},
		// 	ExpectedErr: constant.ErrDefault,
		// 	Setup: func(ctx context.Context) {
		// 		studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", nil)
		// 		locationService.On("GetLocationNameByID", ctx, tx, mock.Anything).Return("", nil)
		// 		orderService.On("CreateCustomOrder",
		// 			ctx,
		// 			tx,
		// 			mock.Anything,
		// 			mock.Anything,
		// 			mock.Anything,
		// 		).Return(entities.Order{}, nil)
		// 		orderItemService.On("CreateMultiCustomOrderItem",
		// 			ctx,
		// 			tx,
		// 			mock.Anything,
		// 			mock.Anything,
		// 		).Return([]entities.OrderItem{{}, {}}, nil)
		// 		billingService.On("CreateBillItemForCustomOrder",
		// 			ctx,
		// 			tx,
		// 			mock.Anything,
		// 			mock.Anything,
		// 			mock.Anything,
		// 		).Return(nil)
		// 		tx.On("Commit", mock.Anything).Return(nil)
		// 		db.On("Begin", mock.Anything).Return(tx, nil)
		// 		elasticService.On("InsertOrderData", ctx, mock.Anything).Return(constant.ErrDefault)
		// 	},
		// },
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.CreateCustomBillingRequest{
				LocationId: "1",
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				studentService.On("GetStudentAndNameByID", ctx, tx, mock.Anything).Return(entities.Student{}, "", nil)
				locationService.On("GetLocationNameByID", ctx, tx, mock.Anything).Return("", nil)
				orderService.On("CreateCustomOrder",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(entities.Order{}, nil)
				orderItemService.On("CreateMultiCustomOrderItem",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
				).Return([]entities.OrderItem{{}, {}}, nil)
				billingService.On("CreateBillItemForCustomOrder",
					ctx,
					tx,
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				subscriptionService.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				// elasticService.On("InsertOrderData", ctx, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			orderService = new(mockServices.IOrderServiceForCreateCustomOrder)
			studentService = new(mockServices.IStudentServiceForCreateCustomOrder)
			locationService = new(mockServices.ILocationServiceForCreateCustomOrder)
			elasticService = new(mockServices.IElasticSearchServiceForCreateCustomOrder)
			orderItemService = new(mockServices.IOrderItemServiceForCreateCustomOrder)
			billingService = new(mockServices.IBillingServiceForCreateCustomOrder)
			subscriptionService = new(mockServices.ISubscriptionServiceForCreateOrder)
			testCase.Setup(testCase.Ctx)
			s := &CreateCustomOrder{
				DB:                   db,
				ElasticSearchService: elasticService,
				StudentService:       studentService,
				LocationService:      locationService,
				OrderService:         orderService,
				OrderItemService:     orderItemService,
				BillingService:       billingService,
				SubscriptionService:  subscriptionService,
			}

			resp, err := s.CreateCustomBilling(testCase.Ctx, testCase.Req.(*pb.CreateCustomBillingRequest))
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
				elasticService,
				studentService,
				locationService,
				orderItemService,
				billingService,
			)
		})
	}
}
