package ordermgmt

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
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_VoidOrderService_VoidOrder(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db                      *mockDb.Ext
		tx                      *mockDb.Tx
		orderService            *mockServices.IOrderServiceForVoidOrder
		studentProductService   *mockServices.IStudentProductServiceForVoidOrder
		billItemService         *mockServices.IBillItemServiceForVoidOrder
		studentService          *mockServices.IStudentServiceForVoidOrder
		subscriptionService     *mockServices.ISubscriptionServiceForVoidOrder
		productService          *mockServices.IProductServiceForVoidOrder
		studentPackageService   *mockServices.IStudentPackageServiceForVoidOrder
		upcomingBillItemService *mockServices.IUpcomingBillItemServiceForVoidOrder
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when void order and return order and student product ids",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{}, []string{constant.StudentProductID}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when getting student and name by student id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when voiding bill item by order id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when voiding upcoming bill item by order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("VoidUpcomingBillItemsByOrder", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when voiding upcoming bill item by order",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("VoidUpcomingBillItemsByOrder", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when voiding student product",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("VoidUpcomingBillItemsByOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentProductService.On("VoidStudentProduct", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, entities.Product{}, false, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when voiding student package and student course",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("VoidUpcomingBillItemsByOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentProductService.On("VoidStudentProduct", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: 2,
					},
				}, false, nil)
				studentPackageService.On("VoidStudentPackageAndStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return([]*npb.EventStudentPackage{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when convert notification message",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("VoidUpcomingBillItemsByOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentProductService.On("VoidStudentProduct", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: 2,
					},
				}, false, nil)
				studentPackageService.On("VoidStudentPackageAndStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return([]*npb.EventStudentPackage{}, nil)
				subscriptionService.On("ToNotificationMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&payload.UpsertSystemNotification{}, constant.ErrDefault)
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Fail case: Error when publishing student package event",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: constant.ErrDefault,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
						Status: 2,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("VoidUpcomingBillItemsByOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentProductService.On("VoidStudentProduct", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: 2,
					},
				}, false, nil)
				studentPackageService.On("VoidStudentPackageAndStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return([]*npb.EventStudentPackage{}, nil)
				subscriptionService.On("ToNotificationMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&payload.UpsertSystemNotification{}, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				subscriptionService.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(constant.ErrDefault)
			},
		},
		{
			Name: "Happy case",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.VoidOrderRequest{
				OrderId: constant.OrderID,
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("VoidOrderReturnOrderAndStudentProductIDs", ctx, tx, mock.Anything, mock.Anything).Return(entities.Order{
					StudentID: pgtype.Text{
						String: constant.StudentID,
						Status: pgtype.Present,
					},
					OrderType: pgtype.Text{
						String: pb.OrderType_ORDER_TYPE_NEW.String(),
						Status: 2,
					},
				}, []string{constant.StudentProductID}, nil)
				studentService.On("GetStudentAndNameByID", mock.Anything, mock.Anything, mock.Anything).Return(entities.Student{}, constant.StudentName, nil)
				billItemService.On("VoidBillItemByOrderID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				upcomingBillItemService.On("VoidUpcomingBillItemsByOrder", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				studentProductService.On("VoidStudentProduct", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(entities.StudentProduct{}, entities.Product{
					ProductType: pgtype.Text{
						String: pb.ProductType_PRODUCT_TYPE_PACKAGE.String(),
						Status: 2,
					},
				}, false, nil)
				studentPackageService.On("VoidStudentPackageAndStudentCourse", mock.Anything, mock.Anything, mock.Anything).Return([]*npb.EventStudentPackage{}, nil)
				subscriptionService.On("ToNotificationMessage", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(&payload.UpsertSystemNotification{}, nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
				subscriptionService.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			orderService = new(mockServices.IOrderServiceForVoidOrder)
			studentProductService = new(mockServices.IStudentProductServiceForVoidOrder)
			billItemService = new(mockServices.IBillItemServiceForVoidOrder)
			studentService = new(mockServices.IStudentServiceForVoidOrder)
			subscriptionService = new(mockServices.ISubscriptionServiceForVoidOrder)
			productService = new(mockServices.IProductServiceForVoidOrder)
			studentPackageService = new(mockServices.IStudentPackageServiceForVoidOrder)
			upcomingBillItemService = new(mockServices.IUpcomingBillItemServiceForVoidOrder)
			testCase.Setup(testCase.Ctx)
			s := &VoidOrder{
				DB:                      db,
				OrderService:            orderService,
				StudentProductService:   studentProductService,
				BillItemService:         billItemService,
				StudentService:          studentService,
				SubscriptionService:     subscriptionService,
				ProductService:          productService,
				StudentPackageService:   studentPackageService,
				UpcomingBillItemService: upcomingBillItemService,
			}

			resp, err := s.VoidOrder(testCase.Ctx, testCase.Req.(*pb.VoidOrderRequest))
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
				studentProductService,
				upcomingBillItemService,
			)
		})
	}
}
