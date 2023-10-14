package ordermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/database"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IOrderServiceForUpdateOrderStatus interface {
	UpdateOrderStatus(
		ctx context.Context,
		db database.QueryExecer,
		orderId string,
		orderStatus pb.OrderStatus,
	) (err error)
}

type UpdateOrderStatus struct {
	DB database.Ext

	OrderService IOrderServiceForUpdateOrderStatus
}

func (s *UpdateOrderStatus) UpdateOrderStatus(ctx context.Context, req *pb.UpdateOrderStatusRequest) (res *pb.UpdateOrderStatusResponse, err error) {
	errors := make([]*pb.UpdateOrderStatusResponse_UpdateOrderStatusError, 0, len(req.UpdateOrdersStatuses))
	err = database.ExecInTxWithContextDeadline(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for _, updatedOrder := range req.UpdateOrdersStatuses {
			if len(strings.TrimSpace(updatedOrder.OrderId)) == 0 || updatedOrder.OrderStatus == pb.OrderStatus_ORDER_STATUS_ALL {
				return status.Errorf(codes.InvalidArgument, "invalid order id or order status with order id %v and status %v", updatedOrder.OrderId, updatedOrder.OrderStatus.String())
			}
			if err := s.OrderService.UpdateOrderStatus(ctx, tx, updatedOrder.OrderId, updatedOrder.OrderStatus); err != nil {
				errors = append(errors, &pb.UpdateOrderStatusResponse_UpdateOrderStatusError{
					OrderId: updatedOrder.OrderId,
					Error:   fmt.Sprintf("unable to update order status: %s", err),
				})
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	res = &pb.UpdateOrderStatusResponse{
		Errors: errors,
	}
	return
}

func NewUpdateOrderStatus(db database.Ext) *UpdateOrderStatus {
	return &UpdateOrderStatus{
		DB:           db,
		OrderService: orderService.NewOrderService(),
	}
}
