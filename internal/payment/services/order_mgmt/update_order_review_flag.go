package ordermgmt

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	billItemService "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	orderService "github.com/manabie-com/backend/internal/payment/services/domain_service/order"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgx/v4"
)

type IOrderServiceForUpdateOrderReviewFlag interface {
	UpdateOrderReview(
		ctx context.Context,
		db database.QueryExecer,
		orderId string,
		isReview bool,
		orderVersionNumber int32,
	) (err error)
}

type IBillItemServiceForUpdateOrderReviewFlag interface {
	UpdateReviewFlagForBillItem(
		ctx context.Context,
		db database.QueryExecer,
		orderID string,
		isReviewFlag bool,
	) (err error)
}

type UpdateOrderReviewFlag struct {
	DB database.Ext

	OrderService    IOrderServiceForUpdateOrderReviewFlag
	BillItemService IBillItemServiceForUpdateOrderReviewFlag
}

func (s *UpdateOrderReviewFlag) UpdateOrderReviewedFlag(ctx context.Context, req *pb.UpdateOrderReviewedFlagRequest) (res *pb.UpdateOrderReviewedFlagResponse, err error) {
	err = database.ExecInTxWithContextDeadline(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) (err error) {
		err = s.OrderService.UpdateOrderReview(ctx, tx, req.OrderId, req.IsReviewed, req.OrderVersionNumber)
		if err != nil {
			return
		}
		err = s.BillItemService.UpdateReviewFlagForBillItem(ctx, tx, req.OrderId, req.IsReviewed)
		return
	})
	if err != nil {
		return
	}
	res = &pb.UpdateOrderReviewedFlagResponse{Successful: true}
	return
}

func NewUpdateOrderReviewFlag(db database.Ext) *UpdateOrderReviewFlag {
	return &UpdateOrderReviewFlag{
		DB:              db,
		OrderService:    orderService.NewOrderService(),
		BillItemService: billItemService.NewBillItemService(),
	}
}
