package service

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/payment/entities"
	pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"go.uber.org/multierr"
)

func (s *InternalService) GenerateBillingItems(ctx context.Context, _ *pb.GenerateBillingItemsRequest) (resp *pb.GenerateBillingItemsResponse, err error) {
	var (
		upcomingBillItems []entities.UpcomingBillItem
	)
	resp = new(pb.GenerateBillingItemsResponse)
	resp.Successful = false
	upcomingBillItems, err = s.upcomingBillItemService.GetUpcomingBillItemsForGenerate(ctx, s.DB)
	if err != nil {
		return
	}
	successful := 0
	failed := 0
	for index := range upcomingBillItems {
		err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
			err = s.generateBillItem(ctx, tx, upcomingBillItems[index])
			if err != nil {
				_ = tx.Rollback(ctx)
			}
			return err
		})
		if err != nil {
			if err.Error() == "last billing item" {
				_ = s.SetLastUpcomingBillItem(ctx, s.DB, upcomingBillItems[index])
				continue
			}
			_ = s.addExecuteNoteForCurrentUpcomingBillItem(ctx, s.DB, upcomingBillItems[index], err)
			failed++
			continue
		}
		successful++
		_ = s.updateCurrentUpcomingBillItemStatus(ctx, s.DB, upcomingBillItems[index])
	}
	resp.Successful = true
	resp.Successed = int32(successful)
	resp.Failed = int32(failed)
	return
}

func (s *InternalService) generateBillItem(
	ctx context.Context,
	db database.QueryExecer,
	upcomingBillItem entities.UpcomingBillItem,
) (err error) {
	var (
		studentProduct              entities.StudentProduct
		billingSchedulePeriod       entities.BillingSchedulePeriod
		billingPeriod               entities.BillingSchedulePeriod
		billItems                   []entities.BillItem
		newBillItem                 *entities.BillItem
		latestBillingSchedulePeriod entities.BillingSchedulePeriod
	)
	studentProduct, err = s.studentProductService.GetStudentProductByStudentProductIDForUpdate(ctx, db, upcomingBillItem.StudentProductID.String)
	if err != nil {
		return
	}
	if studentProduct.UpdatedToStudentProductID.Status == pgtype.Present ||
		studentProduct.ProductStatus.String == pb.StudentProductStatus_CANCELLED.String() ||
		studentProduct.StudentProductLabel.String == pb.StudentProductLabel_PAUSED.String() {
		return
	}
	billItems, err = s.billItemService.GetRecurringBillItemsByOrderIDAndProductID(ctx, db, upcomingBillItem.OrderID.String, upcomingBillItem.ProductID.String)
	if err != nil {
		return
	}
	if len(billItems) == 0 {
		err = fmt.Errorf(
			"bill Item not found with order_id: %v and product_id: %v", upcomingBillItem.OrderID, upcomingBillItem.ProductID)
		return
	}
	latestBillItem := billItems[0]
	billingSchedulePeriod, err = s.billingScheduleService.GetBillingSchedulePeriodByID(ctx, db, upcomingBillItem.BillingSchedulePeriodID.String)
	if err != nil {
		return
	}
	latestBillingSchedulePeriod, err = s.billingScheduleService.GetLatestBillingSchedulePeriod(ctx, db, billingSchedulePeriod.BillingScheduleID.String)
	if err != nil {
		return
	}
	if billingSchedulePeriod.BillingSchedulePeriodID.String == latestBillingSchedulePeriod.BillingSchedulePeriodID.String {
		err = fmt.Errorf("last billing item")
		return
	}
	billingPeriod, err = s.billingScheduleService.GetNextBillingSchedulePeriod(ctx, db, billingSchedulePeriod.BillingScheduleID.String, billingSchedulePeriod.EndDate.Time)
	if err != nil {
		return
	}
	billItemDescription, _ := latestBillItem.GetBillingItemDescription()
	billItemDescription.BillingRatioDenominator = nil
	billItemDescription.BillingRatioNumerator = nil
	billItemDescription.BillingPeriodName = &billingPeriod.Name.String
	billItemDescription.DiscountName = nil
	newBillItem = new(entities.BillItem)
	err = multierr.Combine(
		newBillItem.OrderID.Set(latestBillItem.OrderID.String),
		newBillItem.StudentID.Set(latestBillItem.StudentID.String),
		newBillItem.ProductID.Set(latestBillItem.ProductID.String),
		newBillItem.StudentProductID.Set(latestBillItem.StudentProductID.String),
		newBillItem.BillStatus.Set(pb.BillingStatus_BILLING_STATUS_PENDING.String()),
		newBillItem.BillType.Set(pb.BillingType_BILLING_TYPE_UPCOMING_BILLING.String()),
		newBillItem.BillDate.Set(billingPeriod.BillingDate.Time),
		newBillItem.BillFrom.Set(billingPeriod.StartDate.Time),
		newBillItem.BillTo.Set(billingPeriod.EndDate.Time),
		newBillItem.BillSchedulePeriodID.Set(billingPeriod.BillingSchedulePeriodID.String),
		newBillItem.ProductDescription.Set(latestBillItem.ProductDescription.String),
		newBillItem.ProductPricing.Set(nil),
		newBillItem.DiscountID.Set(nil),
		newBillItem.DiscountAmountValue.Set(0),
		newBillItem.DiscountAmountType.Set(nil),
		newBillItem.DiscountAmount.Set(0),
		newBillItem.RawDiscountAmount.Set(0),
		newBillItem.TaxCategory.Set(nil),
		newBillItem.TaxPercentage.Set(nil),
		newBillItem.TaxID.Set(nil),
		newBillItem.TaxAmount.Set(nil),
		newBillItem.FinalPrice.Set(0),
		newBillItem.BillApprovalStatus.Set(nil),
		newBillItem.OldPrice.Set(0),
		newBillItem.PreviousBillItemStatus.Set(nil),
		newBillItem.PreviousBillItemSequenceNumber.Set(nil),
		newBillItem.AdjustmentPrice.Set(nil),
		newBillItem.IsLatestBillItem.Set(true),
		newBillItem.LocationID.Set(latestBillItem.LocationID.String),
		newBillItem.LocationName.Set(latestBillItem.LocationName.String),
		newBillItem.BillingRatioNumerator.Set(nil),
		newBillItem.BillingRatioDenominator.Set(nil),
	)
	if err != nil {
		return
	}
	err = s.calculatorPrice(ctx, db, billItems, upcomingBillItem, newBillItem, studentProduct, billItemDescription, billingPeriod)
	if err != nil {
		return
	}
	err = s.billItemService.CreateUpcomingBillItems(ctx, db, newBillItem)
	if err != nil {
		return
	}
	err = s.upcomingBillItemService.CreateUpcomingBillItem(ctx, db, *newBillItem)
	if err != nil {
		return
	}
	return
}

func (s *InternalService) calculatorPrice(ctx context.Context,
	db database.QueryExecer,
	billItems []entities.BillItem,
	upcomingBillItem entities.UpcomingBillItem,
	billItem *entities.BillItem,
	studentProduct entities.StudentProduct,
	billingDescription *entities.BillingItemDescription,
	billingSchedulePeriod entities.BillingSchedulePeriod,
) (err error) {
	var (
		discount           entities.Discount
		tax                entities.Tax
		isEnrolledInOrg    bool
		priceType          = pb.ProductPriceType_DEFAULT_PRICE.String()
		rootStudentProduct entities.StudentProduct
	)
	discount, err = s.discountService.VerifyDiscountForGenerateUpcomingBillItem(ctx, db, billItems)
	if discount.DiscountID.Status == pgtype.Present {
		billingDescription.DiscountName = &discount.Name.String
	}
	if err != nil {
		return
	}
	if upcomingBillItem.TaxID.String != "" {
		tax, err = s.taxService.GetTaxByID(ctx, db, upcomingBillItem.TaxID.String)
		if err != nil {
			return
		}
	}
	productPrices, err := s.priceService.GetProductPricesByProductIDAndPriceType(ctx, db, studentProduct.ProductID.String, pb.ProductPriceType_ENROLLED_PRICE.String())
	if err != nil {
		return
	}
	if len(productPrices) > 0 {
		rootStudentProductCreatedAt := studentProduct.CreatedAt.Time
		if studentProduct.RootStudentProductID.Status == pgtype.Present {
			rootStudentProduct, err = s.studentProductService.GetStudentProductByStudentProductIDForUpdate(ctx, db, studentProduct.RootStudentProductID.String)
			if err != nil {
				return
			}
			rootStudentProductCreatedAt = rootStudentProduct.CreatedAt.Time
		}
		isEnrolledInOrg, err = s.studentService.CheckIsEnrolledInOrgByStudentIDAndTime(ctx, db, studentProduct.StudentID.String, rootStudentProductCreatedAt)
		if err != nil {
			return
		}
		if isEnrolledInOrg {
			priceType = pb.ProductPriceType_ENROLLED_PRICE.String()
		}
	}
	err = billItem.BillingItemDescription.Set(&billingDescription)
	if err != nil {
		return
	}
	err = s.priceService.CalculatorBillItemPrice(ctx, db, billItem, upcomingBillItem, tax, discount, priceType, billingDescription, billingSchedulePeriod)
	return
}

func (s *InternalService) addExecuteNoteForCurrentUpcomingBillItem(ctx context.Context, db database.QueryExecer, upcomingBillItem entities.UpcomingBillItem, err error) (importErr error) {
	importErr = s.upcomingBillItemService.AddExecuteNoteForCurrentUpcomingBillItem(ctx, db, upcomingBillItem, err)
	return
}

func (s *InternalService) updateCurrentUpcomingBillItemStatus(ctx context.Context, db database.QueryExecer, upcomingBillItem entities.UpcomingBillItem) (err error) {
	err = s.upcomingBillItemService.UpdateCurrentUpcomingBillItemStatus(ctx, db, upcomingBillItem)
	return
}

func (s *InternalService) SetLastUpcomingBillItem(ctx context.Context, db database.QueryExecer, upcomingBillItem entities.UpcomingBillItem) (err error) {
	err = s.upcomingBillItemService.SetLastUpcomingBillItem(ctx, db, upcomingBillItem)
	return
}
