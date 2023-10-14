package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BillingService struct {
	OneTimeProductBilling   *BillingServiceForOneTimeProduct
	RecurringProductBilling *BillingServiceForRecurringProduct
}

func (s *BillingService) CreateBillItemForOrderCreate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	if orderItemData.IsOneTimeProduct {
		return s.OneTimeProductBilling.CreateBillItemForOrderCreate(ctx, db, orderItemData)
	}
	return s.RecurringProductBilling.CreateBillItemForOrderCreate(ctx, db, orderItemData)
}

func (s *BillingService) CreateBillItemForOrderUpdate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	if orderItemData.IsOneTimeProduct {
		return s.OneTimeProductBilling.CreateBillItemForOrderUpdate(ctx, db, orderItemData)
	}
	return s.RecurringProductBilling.CreateBillItemForOrderUpdate(ctx, db, orderItemData)
}

func (s *BillingService) CreateBillItemForOrderWithdrawal(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	if orderItemData.IsOneTimeProduct {
		return status.Errorf(codes.InvalidArgument, "we can't withdraw bill item for one time product")
	}
	return s.RecurringProductBilling.CreateBillItemForOrderWithdrawal(ctx, db, orderItemData)
}

func (s *BillingService) CreateBillItemForOrderGraduate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	if orderItemData.IsOneTimeProduct {
		return status.Errorf(codes.InvalidArgument, "we can't graduate bill item for one time product")
	}
	return s.RecurringProductBilling.CreateBillItemForOrderGraduate(ctx, db, orderItemData)
}

func (s *BillingService) CreateBillItemForOrderLOA(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	if orderItemData.IsOneTimeProduct {
		return status.Errorf(codes.InvalidArgument, "we can't pause bill item for one time product")
	}
	return s.RecurringProductBilling.CreateBillItemForOrderLOA(ctx, db, orderItemData)
}

func (s *BillingService) CreateBillItemForOrderCancel(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	if orderItemData.IsOneTimeProduct {
		return s.OneTimeProductBilling.CreateBillItemForOrderCancel(ctx, db, orderItemData)
	}
	return s.RecurringProductBilling.CreateBillItemForOrderCancel(ctx, db, orderItemData)
}

func NewBillingService() utils.IBillingService {
	return &BillingService{
		OneTimeProductBilling:   NewBillingServiceForOneTimeProduct(),
		RecurringProductBilling: NewBillingServiceForRecurringProduct(),
	}
}
