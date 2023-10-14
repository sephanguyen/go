package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	bill_item "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	discount "github.com/manabie-com/backend/internal/payment/services/domain_service/discount"
	price "github.com/manabie-com/backend/internal/payment/services/domain_service/price"
	tax "github.com/manabie-com/backend/internal/payment/services/domain_service/tax"
	"github.com/manabie-com/backend/internal/payment/utils"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type IDiscountServiceForOneTimeBilling interface {
	IsValidDiscountForOneTimeBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		discountName *string) (err error)
}

type ITaxServiceForOneTimeBilling interface {
	IsValidTaxForOneTimeBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData) (err error)
}

type IPriceServiceForOneTimeBilling interface {
	IsValidPriceForOneTimeBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData) (err error)
	IsValidAdjustmentPriceForOneTimeBilling(
		oldBillItem entities.BillItem,
		orderItemData utils.OrderItemData) (err error)
}

type IBillItemServiceForOneTimeBilling interface {
	CreateNewBillItemForOneTimeBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		discountName string,
	) (err error)
	CreateUpdateBillItemForOneTimeBilling(
		ctx context.Context,
		db database.QueryExecer,
		oldBillItem entities.BillItem,
		orderItemData utils.OrderItemData,
		discountName string,
	) (err error)
	CreateCancelBillItemForOneTimeBilling(
		ctx context.Context,
		db database.QueryExecer,
		oldBillItem entities.BillItem) (err error)
	GetOldBillItemForUpdateOneTimeBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData) (billItem entities.BillItem, err error)
}

type BillingServiceForOneTimeProduct struct {
	DiscountService IDiscountServiceForOneTimeBilling
	TaxService      ITaxServiceForOneTimeBilling
	PriceService    IPriceServiceForOneTimeBilling
	BillItemService IBillItemServiceForOneTimeBilling
}

func (s *BillingServiceForOneTimeProduct) CreateBillItemForOrderCreate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	if len(orderItemData.BillItems) != 1 {
		err = status.Errorf(codes.FailedPrecondition,
			"we can't create bill item for product %v because quantity bill item is %v",
			orderItemData.ProductInfo.ProductID.String,
			len(orderItemData.BillItems))
		return
	}

	var discountName string

	err = utils.GroupErrorFunc(
		s.DiscountService.IsValidDiscountForOneTimeBilling(ctx, db, orderItemData, &discountName),
		s.TaxService.IsValidTaxForOneTimeBilling(ctx, db, orderItemData),
		s.PriceService.IsValidPriceForOneTimeBilling(ctx, db, orderItemData),
		s.BillItemService.CreateNewBillItemForOneTimeBilling(ctx, db, orderItemData, discountName),
	)
	return
}

func (s *BillingServiceForOneTimeProduct) CreateBillItemForOrderUpdate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	var oldBillItem entities.BillItem
	if len(orderItemData.BillItems) != 1 {
		err = status.Errorf(codes.FailedPrecondition,
			"we can't update bill item for product %v because quantity bill item is %v",
			orderItemData.ProductInfo.ProductID.String,
			len(orderItemData.BillItems))
		return
	}
	var discountName string

	oldBillItem, err = s.BillItemService.GetOldBillItemForUpdateOneTimeBilling(ctx, db, orderItemData)
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		s.DiscountService.IsValidDiscountForOneTimeBilling(ctx, db, orderItemData, &discountName),
		s.TaxService.IsValidTaxForOneTimeBilling(ctx, db, orderItemData),
		s.PriceService.IsValidPriceForOneTimeBilling(ctx, db, orderItemData),
		s.PriceService.IsValidAdjustmentPriceForOneTimeBilling(oldBillItem, orderItemData),
		s.BillItemService.CreateUpdateBillItemForOneTimeBilling(ctx, db, oldBillItem, orderItemData, discountName),
	)
	return
}

func (s *BillingServiceForOneTimeProduct) CreateBillItemForOrderCancel(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	var oldBillItem entities.BillItem
	if len(orderItemData.BillItems) != 1 {
		err = status.Errorf(codes.FailedPrecondition,
			"we can't cancel bill item for product %v because quantity bill item is %v",
			orderItemData.ProductInfo.ProductID.String,
			len(orderItemData.BillItems))
		return
	}

	oldBillItem, err = s.BillItemService.GetOldBillItemForUpdateOneTimeBilling(ctx, db, orderItemData)
	if err != nil {
		return
	}
	_ = oldBillItem.OrderID.Set(orderItemData.Order.OrderID.String)
	err = s.BillItemService.CreateCancelBillItemForOneTimeBilling(ctx, db, oldBillItem)
	return
}

func NewBillingServiceForOneTimeProduct() *BillingServiceForOneTimeProduct {
	return &BillingServiceForOneTimeProduct{
		DiscountService: discount.NewDiscountService(),
		TaxService:      tax.NewTaxService(),
		PriceService:    price.NewPriceService(),
		BillItemService: bill_item.NewBillItemService(),
	}
}
