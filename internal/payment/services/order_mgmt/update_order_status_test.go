package ordermgmt

import (
	"context"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/payment/constant"
	"github.com/manabie-com/backend/internal/payment/utils"
	mockDb "github.com/manabie-com/backend/mock/golibs/database"
	mockServices "github.com/manabie-com/backend/mock/payment/services/order_mgmt"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateOrderStatus(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), utils.TimeOut)
	defer cancel()

	var (
		db           *mockDb.Ext
		tx           *mockDb.Tx
		orderService *mockServices.IOrderServiceForUpdateOrderStatus
	)

	testcases := []utils.TestCase{
		{
			Name: "Fail case: Error when void order by order id",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.UpdateOrderStatusRequest{
				UpdateOrdersStatuses: []*pb.UpdateOrderStatusRequest_UpdateOrderStatus{
					{
						OrderId:     "1",
						OrderStatus: pb.OrderStatus_ORDER_STATUS_ALL,
					},
				},
			},
			ExpectedErr: fmt.Errorf("invalid order id or order status with order id"),
			Setup: func(ctx context.Context) {
				tx.On("Rollback", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Happy case: with err inside update order status",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.UpdateOrderStatusRequest{
				UpdateOrdersStatuses: []*pb.UpdateOrderStatusRequest_UpdateOrderStatus{
					{
						OrderId:     "1",
						OrderStatus: pb.OrderStatus_ORDER_STATUS_PENDING,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("UpdateOrderStatus", ctx, tx, mock.Anything, mock.Anything).Return(constant.ErrDefault)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
		{
			Name: "Happy case: without err inside update order status",
			Ctx:  interceptors.ContextWithUserID(ctx, constant.UserID),
			Req: &pb.UpdateOrderStatusRequest{
				UpdateOrdersStatuses: []*pb.UpdateOrderStatusRequest_UpdateOrderStatus{
					{
						OrderId:     "1",
						OrderStatus: pb.OrderStatus_ORDER_STATUS_SUBMITTED,
					},
				},
			},
			ExpectedErr: nil,
			Setup: func(ctx context.Context) {
				orderService.On("UpdateOrderStatus", ctx, tx, mock.Anything, mock.Anything).Return(nil)
				tx.On("Commit", mock.Anything).Return(nil)
				db.On("Begin", mock.Anything).Return(tx, nil)
			},
		},
	}

	for _, testCase := range testcases {
		t.Run(testCase.Name, func(t *testing.T) {
			db = new(mockDb.Ext)
			tx = new(mockDb.Tx)
			orderService = new(mockServices.IOrderServiceForUpdateOrderStatus)

			testCase.Setup(testCase.Ctx)
			s := &UpdateOrderStatus{
				DB:           db,
				OrderService: orderService,
			}

			resp, err := s.UpdateOrderStatus(testCase.Ctx, testCase.Req.(*pb.UpdateOrderStatusRequest))
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
			)
		})
	}
}
