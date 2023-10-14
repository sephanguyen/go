package service

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	billItem "github.com/manabie-com/backend/internal/payment/services/domain_service/billing/bill_item"
	"github.com/manabie-com/backend/internal/payment/services/domain_service/billing/billing_schedule"
	discount "github.com/manabie-com/backend/internal/payment/services/domain_service/discount"
	price "github.com/manabie-com/backend/internal/payment/services/domain_service/price"
	tax "github.com/manabie-com/backend/internal/payment/services/domain_service/tax"
	"github.com/manabie-com/backend/internal/payment/utils"
)

type IBillingScheduleServiceForRecurringBilling interface {
	CheckScheduleReturnProRatedItemAndMapPeriodInfo(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (
		proRatedBillItem utils.BillingItemData,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData,
		mapPeriodInfo map[string]entities.BillingSchedulePeriod,
		err error,
	)
}

type IDiscountServiceForRecurringBilling interface {
	IsValidDiscountForRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		proRatedBillItem utils.BillingItemData,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData,
		discountName *string) (err error)
}

type IPriceServiceForRecurringBilling interface {
	IsValidPriceForRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		proRatedBillItem utils.BillingItemData,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData) (proRatedPRice entities.ProductPrice, err error)
	IsValidPriceForUpdateRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		proRatedBillItem utils.BillingItemData,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData,
		mapOldBillingItem map[string]entities.BillItem,
	) (proRatedPRice entities.ProductPrice, err error)
	IsValidPriceForCancelRecurringBilling(
		orderItemData utils.OrderItemData,
		proRatedBillItem utils.BillingItemData,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData,
		mapOldBillingItem map[string]entities.BillItem,
		mapPeriodInfo map[string]entities.BillingSchedulePeriod,
	) (err error)
}

type ITaxServiceForRecurringBilling interface {
	IsValidTaxForRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
	) (err error)
}

type IBillItemServiceForRecurringBilling interface {
	CreateNewBillItemForRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		proRatedBillItem utils.BillingItemData,
		proRatedPrice entities.ProductPrice,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData,
		mapPeriodInfo map[string]entities.BillingSchedulePeriod,
		discountName string,
	) (err error)
	CreateUpdateBillItemForRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		proRatedBillItem utils.BillingItemData,
		proRatedPrice entities.ProductPrice,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData,
		mapPeriodInfo map[string]entities.BillingSchedulePeriod,
		mapOldBillingItem map[string]entities.BillItem,
		discountName string,
	) (err error)
	CreateCancelBillItemForRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		proRatedBillItem utils.BillingItemData,
		ratioOfProRatedBillingItem entities.BillingRatio,
		normalBillItem []utils.BillingItemData,
		mapPeriodInfo map[string]entities.BillingSchedulePeriod,
		mapOldBillingItem map[string]entities.BillItem,
	) (err error)
	GetMapOldBillingItemForRecurringBilling(
		ctx context.Context,
		db database.QueryExecer,
		orderItemData utils.OrderItemData,
		mapPeriodInfo map[string]entities.BillingSchedulePeriod,
	) (mapOldBillingItem map[string]entities.BillItem, err error)
}

type BillingServiceForRecurringProduct struct {
	TaxService             ITaxServiceForRecurringBilling
	PriceService           IPriceServiceForRecurringBilling
	BillItemService        IBillItemServiceForRecurringBilling
	DiscountService        IDiscountServiceForRecurringBilling
	BillingScheduleService IBillingScheduleServiceForRecurringBilling
}

func (s *BillingServiceForRecurringProduct) CreateBillItemForOrderCreate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	var (
		proRatedBillItem           utils.BillingItemData
		ratioOfProRatedBillingItem entities.BillingRatio
		proRatedPrice              entities.ProductPrice
		nonProRatedBillItems       []utils.BillingItemData
		mapPeriodInfo              map[string]entities.BillingSchedulePeriod
		discountName               string
	)

	proRatedBillItem,
		ratioOfProRatedBillingItem,
		nonProRatedBillItems,
		mapPeriodInfo,
		err = s.BillingScheduleService.CheckScheduleReturnProRatedItemAndMapPeriodInfo(
		ctx,
		db,
		orderItemData,
	)
	if err != nil {
		return
	}
	err = utils.GroupErrorFunc(
		s.DiscountService.IsValidDiscountForRecurringBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			nonProRatedBillItems,
			&discountName,
		),
		s.TaxService.IsValidTaxForRecurringBilling(
			ctx,
			db,
			orderItemData,
		),
	)
	if err != nil {
		return
	}

	proRatedPrice, err = s.PriceService.IsValidPriceForRecurringBilling(
		ctx,
		db,
		orderItemData,
		proRatedBillItem,
		ratioOfProRatedBillingItem,
		nonProRatedBillItems,
	)

	if err != nil {
		return
	}

	err = s.BillItemService.CreateNewBillItemForRecurringBilling(
		ctx,
		db,
		orderItemData,
		proRatedBillItem,
		proRatedPrice,
		ratioOfProRatedBillingItem,
		nonProRatedBillItems,
		mapPeriodInfo,
		discountName,
	)
	return
}

func (s *BillingServiceForRecurringProduct) CreateBillItemForOrderUpdate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	var (
		proRatedBillItem           utils.BillingItemData
		ratioOfProRatedBillingItem entities.BillingRatio
		proRatedPrice              entities.ProductPrice
		normalBillItem             []utils.BillingItemData
		mapPeriodInfo              map[string]entities.BillingSchedulePeriod
		mapOldBillingItem          map[string]entities.BillItem
		discountName               string
	)

	proRatedBillItem,
		ratioOfProRatedBillingItem,
		normalBillItem,
		mapPeriodInfo,
		err = s.BillingScheduleService.CheckScheduleReturnProRatedItemAndMapPeriodInfo(
		ctx,
		db,
		orderItemData,
	)
	if err != nil {
		return
	}

	mapOldBillingItem, err = s.BillItemService.GetMapOldBillingItemForRecurringBilling(ctx, db, orderItemData, mapPeriodInfo)
	if err != nil {
		return
	}
	proRatedPrice, err = s.PriceService.IsValidPriceForUpdateRecurringBilling(
		ctx,
		db,
		orderItemData,
		proRatedBillItem,
		ratioOfProRatedBillingItem,
		normalBillItem,
		mapOldBillingItem,
	)
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		s.DiscountService.IsValidDiscountForRecurringBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			&discountName,
		),
		s.TaxService.IsValidTaxForRecurringBilling(
			ctx,
			db,
			orderItemData,
		),

		s.BillItemService.CreateUpdateBillItemForRecurringBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			proRatedPrice,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapPeriodInfo,
			mapOldBillingItem,
			discountName,
		),
	)

	return
}

func (s *BillingServiceForRecurringProduct) CreateBillItemForOrderWithdrawal(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {
	var (
		proRatedBillItem           utils.BillingItemData
		ratioOfProRatedBillingItem entities.BillingRatio
		normalBillItem             []utils.BillingItemData
		mapOldBillingItem          map[string]entities.BillItem
		mapPeriodInfo              map[string]entities.BillingSchedulePeriod
	)

	proRatedBillItem,
		ratioOfProRatedBillingItem,
		normalBillItem,
		mapPeriodInfo,
		err = s.BillingScheduleService.CheckScheduleReturnProRatedItemAndMapPeriodInfo(
		ctx,
		db,
		orderItemData,
	)
	if err != nil {
		return
	}

	mapOldBillingItem, err = s.BillItemService.GetMapOldBillingItemForRecurringBilling(ctx, db, orderItemData, mapPeriodInfo)
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		s.PriceService.IsValidPriceForCancelRecurringBilling(
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapOldBillingItem,
			mapPeriodInfo,
		),
		s.BillItemService.CreateCancelBillItemForRecurringBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapPeriodInfo,
			mapOldBillingItem,
		),
	)
	return
}

func (s *BillingServiceForRecurringProduct) CreateBillItemForOrderGraduate(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {

	var (
		proRatedBillItem           utils.BillingItemData
		ratioOfProRatedBillingItem entities.BillingRatio
		normalBillItem             []utils.BillingItemData
		mapOldBillingItem          map[string]entities.BillItem
		mapPeriodInfo              map[string]entities.BillingSchedulePeriod
	)

	proRatedBillItem,
		ratioOfProRatedBillingItem,
		normalBillItem,
		mapPeriodInfo,
		err = s.BillingScheduleService.CheckScheduleReturnProRatedItemAndMapPeriodInfo(
		ctx,
		db,
		orderItemData,
	)
	if err != nil {
		return
	}

	mapOldBillingItem, err = s.BillItemService.GetMapOldBillingItemForRecurringBilling(ctx, db, orderItemData, mapPeriodInfo)
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		s.PriceService.IsValidPriceForCancelRecurringBilling(
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapOldBillingItem,
			mapPeriodInfo,
		),
		s.BillItemService.CreateCancelBillItemForRecurringBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapPeriodInfo,
			mapOldBillingItem,
		),
	)

	return
}

func (s *BillingServiceForRecurringProduct) CreateBillItemForOrderLOA(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {

	var (
		proRatedBillItem           utils.BillingItemData
		ratioOfProRatedBillingItem entities.BillingRatio
		normalBillItem             []utils.BillingItemData
		mapOldBillingItem          map[string]entities.BillItem
		mapPeriodInfo              map[string]entities.BillingSchedulePeriod
	)

	proRatedBillItem,
		ratioOfProRatedBillingItem,
		normalBillItem,
		mapPeriodInfo,
		err = s.BillingScheduleService.CheckScheduleReturnProRatedItemAndMapPeriodInfo(
		ctx,
		db,
		orderItemData,
	)
	if err != nil {
		return
	}

	mapOldBillingItem, err = s.BillItemService.GetMapOldBillingItemForRecurringBilling(ctx, db, orderItemData, mapPeriodInfo)
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		s.PriceService.IsValidPriceForCancelRecurringBilling(
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapOldBillingItem,
			mapPeriodInfo,
		),
		s.BillItemService.CreateCancelBillItemForRecurringBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapPeriodInfo,
			mapOldBillingItem,
		),
	)

	return
}

func (s *BillingServiceForRecurringProduct) CreateBillItemForOrderCancel(
	ctx context.Context,
	db database.QueryExecer,
	orderItemData utils.OrderItemData) (err error) {

	var (
		proRatedBillItem           utils.BillingItemData
		ratioOfProRatedBillingItem entities.BillingRatio
		normalBillItem             []utils.BillingItemData
		mapOldBillingItem          map[string]entities.BillItem
		mapPeriodInfo              map[string]entities.BillingSchedulePeriod
	)

	proRatedBillItem,
		ratioOfProRatedBillingItem,
		normalBillItem,
		mapPeriodInfo,
		err = s.BillingScheduleService.CheckScheduleReturnProRatedItemAndMapPeriodInfo(
		ctx,
		db,
		orderItemData,
	)
	if err != nil {
		return
	}

	mapOldBillingItem, err = s.BillItemService.GetMapOldBillingItemForRecurringBilling(ctx, db, orderItemData, mapPeriodInfo)
	if err != nil {
		return
	}

	err = utils.GroupErrorFunc(
		s.PriceService.IsValidPriceForCancelRecurringBilling(
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapOldBillingItem,
			mapPeriodInfo,
		),
		s.BillItemService.CreateCancelBillItemForRecurringBilling(
			ctx,
			db,
			orderItemData,
			proRatedBillItem,
			ratioOfProRatedBillingItem,
			normalBillItem,
			mapPeriodInfo,
			mapOldBillingItem,
		),
	)

	return
}

func NewBillingServiceForRecurringProduct() *BillingServiceForRecurringProduct {
	return &BillingServiceForRecurringProduct{
		TaxService:             tax.NewTaxService(),
		PriceService:           price.NewPriceService(),
		BillItemService:        billItem.NewBillItemService(),
		DiscountService:        discount.NewDiscountService(),
		BillingScheduleService: billing_schedule.NewBillingScheduleService(),
	}
}
