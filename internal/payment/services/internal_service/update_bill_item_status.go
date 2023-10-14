package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *InternalService) UpdateBillItemStatus(ctx context.Context, req *pb.UpdateBillItemStatusRequest) (res *pb.UpdateBillItemStatusResponse, err error) {
	var errors []*pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError

	billingItemIDsLength := len(req.UpdateBillItems)
	if billingItemIDsLength < 1 {
		err = status.Error(codes.InvalidArgument, "billing items cannot be empty")
		return
	}

	errors = make([]*pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError, 0, billingItemIDsLength)

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		for _, item := range req.UpdateBillItems {
			var (
				orderID     string
				orderStatus pb.OrderStatus
			)
			orderID, err = s.billItemService.UpdateBillItemStatusAndReturnOrderID(ctx, tx, item.BillItemSequenceNumber, item.BillingStatusTo.String())
			if err != nil {
				errors = append(errors, &pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{BillItemSequenceNumber: item.BillItemSequenceNumber, Error: err.Error()})
				continue
			}

			orderStatus = pb.OrderStatus_ORDER_STATUS_SUBMITTED
			if item.BillingStatusTo == pb.BillingStatus_BILLING_STATUS_INVOICED {
				orderStatus = pb.OrderStatus_ORDER_STATUS_INVOICED
			}

			err = s.orderService.UpdateOrderStatus(ctx, tx, orderID, orderStatus)
			if err != nil {
				errors = append(errors, &pb.UpdateBillItemStatusResponse_UpdateBillItemStatusError{BillItemSequenceNumber: item.BillItemSequenceNumber, Error: err.Error()})
			}
		}
		return nil
	})
	res = &pb.UpdateBillItemStatusResponse{Errors: errors}
	return
}
