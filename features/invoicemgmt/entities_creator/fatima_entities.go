package entitiescreator

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/common"
	invoiceConst "github.com/manabie-com/backend/features/invoicemgmt/constant"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/invoicemgmt/services/utils"
	paymentEntities "github.com/manabie-com/backend/internal/payment/entities"
	paymentRepo "github.com/manabie-com/backend/internal/payment/repositories"
	payment_pb "github.com/manabie-com/backend/pkg/manabuf/payment/v1"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

// CreateStudentProduct used InsertStudentProductStmt to insert student_product.
// stepState dependency:
//   - stepState.StudentID
//   - stepState.ProductID
//   - stepState.LocationID
//
// stepState assigned:
//   - stepState.StudentProductID
func (c *EntitiesCreator) CreateStudentProduct(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		randomStr := idutil.ULIDNow()
		studentProductID := fmt.Sprintf("student-product-id-%v", randomStr)

		studentProduct := &paymentEntities.StudentProduct{}
		database.AllNullEntity(studentProduct)

		now := time.Now()
		err := multierr.Combine(
			studentProduct.StudentProductID.Set(studentProductID),
			studentProduct.StudentID.Set(stepState.StudentID),
			studentProduct.ProductID.Set(stepState.ProductID),
			studentProduct.UpcomingBillingDate.Set(now),
			studentProduct.StartDate.Set(now),
			studentProduct.EndDate.Set(now),
			studentProduct.ProductStatus.Set("PRODUCT_STATUS"),
			studentProduct.ApprovalStatus.Set("APPROVAL_STATUS"),
			studentProduct.LocationID.Set(stepState.LocationID),
			studentProduct.CreatedAt.Set(now),
			studentProduct.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("studentProduct set: %w", err)
		}

		_, err = db.Exec(ctx,
			InsertStudentProductStmt,
			studentProduct.StudentProductID.String,
			studentProduct.StudentID.String,
			studentProduct.ProductID.String,
			studentProduct.ProductStatus.String,
			studentProduct.ApprovalStatus.String,
			studentProduct.LocationID.String,
		)
		if err != nil {
			return err
		}

		stepState.StudentProductID = studentProductID

		return nil
	}
}

// CreateBillingSchedule used InsertBillingScheduleStmt to insert billing_schedule.
// stepState assigned:
//   - stepState.BillingScheduleID
func (c *EntitiesCreator) CreateBillingSchedule(ctx context.Context, db database.QueryExecer, isArchived bool) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithBillingSchedule")
		defer span.End()

		var billingScheduleID string
		randomStr := idutil.ULIDNow()
		name := "BILLING-SCHEDULE-NAME-" + randomStr
		remarks := "BILLING-SCHEDULE-REMARKS-" + time.Now().Format("01-02-2006")

		billingSchedule := &paymentEntities.BillingSchedule{}
		database.AllNullEntity(billingSchedule)

		now := time.Now()
		err := multierr.Combine(
			billingSchedule.BillingScheduleID.Set(randomStr),
			billingSchedule.Name.Set(name),
			billingSchedule.Remarks.Set(remarks),
			billingSchedule.IsArchived.Set(isArchived),
			billingSchedule.CreatedAt.Set(now),
			billingSchedule.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("billingSchedule set: %w", err)
		}

		row := db.QueryRow(ctx, InsertBillingScheduleStmt,
			billingSchedule.BillingScheduleID.String,
			billingSchedule.Name.String,
			billingSchedule.Remarks.String,
			billingSchedule.IsArchived.Bool,
		)
		err = row.Scan(&billingScheduleID)
		if err != nil {
			return fmt.Errorf("error in creating billing schedule: %w", err)
		}

		stepState.BillingScheduleID = billingScheduleID
		return nil
	}
}

// CreateBillingSchedulePeriod used InsertBillingSchedulePeriodStmt to insert billing_schedule_period.
// stepState dependency:
//   - stepState.BillingScheduleID
//
// stepState assigned:
//   - stepState.BillingSchedulePeriodID
func (c *EntitiesCreator) CreateBillingSchedulePeriod(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithBillingSchedulePeriod")
		defer span.End()

		var billingSchedulePeriodID string
		randomStr := idutil.ULIDNow()
		periodStartDate := time.Now()
		periodEndDate := periodStartDate.AddDate(0, 0, 28)
		periodBillingDate := periodStartDate.AddDate(0, -1, 15)

		now := time.Now()
		billingSchedulePeriod := &paymentEntities.BillingSchedulePeriod{}
		err := multierr.Combine(
			billingSchedulePeriod.BillingSchedulePeriodID.Set(randomStr),
			billingSchedulePeriod.Name.Set(fmt.Sprintf("billing-period-name-%s", randomStr)),
			billingSchedulePeriod.BillingScheduleID.Set(stepState.BillingScheduleID),
			billingSchedulePeriod.StartDate.Set(periodStartDate),
			billingSchedulePeriod.EndDate.Set(periodEndDate),
			billingSchedulePeriod.BillingDate.Set(periodBillingDate),
			billingSchedulePeriod.IsArchived.Set(false),
			billingSchedulePeriod.CreatedAt.Set(now),
			billingSchedulePeriod.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("billingSchedulePeriod set: %w", err)
		}

		row := db.QueryRow(
			ctx,
			InsertBillingSchedulePeriodStmt,
			billingSchedulePeriod.BillingSchedulePeriodID.String,
			billingSchedulePeriod.Name.String,
			billingSchedulePeriod.BillingScheduleID.String,
			billingSchedulePeriod.StartDate.Time,
			billingSchedulePeriod.EndDate.Time,
			billingSchedulePeriod.BillingDate.Time,
			billingSchedulePeriod.IsArchived.Bool,
		)
		err = row.Scan(&billingSchedulePeriodID)
		if err != nil {
			fmt.Printf("cannot insert billing schedule period, err: %s", err)
		}

		stepState.BillingSchedulePeriodID = billingSchedulePeriodID

		return err
	}
}

// CreateProduct used InsertProductStmt to insert product.
// stepState dependency:
//   - stepState.BillingScheduleID
//
// stepState assigned:
//   - stepState.ProductID
func (c *EntitiesCreator) CreateProduct(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithProduct")
		defer span.End()

		randomStr := idutil.ULIDNow()
		now := time.Now()

		product := &paymentEntities.Product{}
		err := multierr.Combine(
			product.ProductID.Set(randomStr),
			product.Name.Set(fmt.Sprintf("PRODUCT-NAME-%v", randomStr)),
			product.ProductType.Set("3"),
			product.TaxID.Set(stepState.TaxID),
			product.AvailableFrom.Set(time.Now().AddDate(1, 0, 0)),
			product.AvailableUntil.Set(time.Now().AddDate(2, 0, 0)),
			product.Remarks.Set(fmt.Sprintf("REMARKS-%v", randomStr)),
			product.CustomBillingPeriod.Set(now),
			product.BillingScheduleID.Set(stepState.BillingScheduleID),
			product.DisableProRatingFlag.Set(false),
			product.IsArchived.Set(false),
			product.CreatedAt.Set(now),
			product.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("product set: %w", err)
		}

		row := db.QueryRow(ctx, InsertProductStmt,
			product.ProductID.String,
			product.Name.String,
			product.ProductType.String,
			product.TaxID.String,
			product.AvailableFrom.Time,
			product.AvailableUntil.Time,
			product.Remarks.String,
			product.CustomBillingPeriod.Time,
			product.BillingScheduleID.String,
			product.DisableProRatingFlag.Bool,
			product.IsArchived.Bool,
		)

		var productID string
		err = row.Scan(&productID)
		if err != nil {
			return fmt.Errorf("cannot insert product, err %w", err)
		}

		stepState.ProductID = productID

		return nil
	}
}

// CreateOrder used orderRepo.Create to insert order.
// stepState dependency:
//   - stepState.StudentID
//   - stepState.LocationID
//
// stepState assigned:
//   - stepState.OrderID
func (c *EntitiesCreator) CreateOrder(ctx context.Context, db database.QueryExecer, orderStatus string, isReviewed bool) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithOrder")
		defer span.End()

		randomStr := idutil.ULIDNow()
		orderID := "order-id-" + randomStr

		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			// orderRepo := &paymentRepo.OrderRepo{}
			order := &paymentEntities.Order{}
			database.AllNullEntity(order)

			err := multierr.Combine(
				order.OrderID.Set(orderID),
				order.StudentID.Set(stepState.StudentID),
				order.OrderComment.Set("order_comment"),
				order.OrderStatus.Set(orderStatus),
				order.StudentFullName.Set("Student Full Name"),
				order.OrderType.Set("order_type"),
				order.LocationID.Set(stepState.LocationID),
				order.IsReviewed.Set(isReviewed),
			)
			if err != nil {
				return false, fmt.Errorf("order set: %w", err)
			}

			_, err = db.Exec(ctx,
				InsertOrderStmt,
				order.OrderID.String,
				order.StudentID.String,
				order.OrderComment.String,
				order.OrderStatus.String,
				order.StudentFullName.String,
				order.OrderType.String,
				order.LocationID.String,
				order.IsReviewed.Bool,
			)
			if err == nil {
				stepState.OrderID = orderID
				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("order create err: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create order, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		return nil
	}
}

// CreateTax used InsertTaxStmt to insert tax.
// stepState assigned:
//   - stepState.TaxID
func (c *EntitiesCreator) CreateTax(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		ctx, span := interceptors.StartSpan(ctx, "EntitiesCreator.WithTax")
		defer span.End()

		randomStr := idutil.ULIDNow()
		now := time.Now()

		tax := &paymentEntities.Tax{}
		err := multierr.Combine(
			tax.TaxID.Set(randomStr),
			tax.Name.Set(fmt.Sprintf("tax-name-"+randomStr)),
			tax.TaxPercentage.Set(10),
			tax.TaxCategory.Set("TAX_CATEGORY_INCLUSIVE"),
			tax.DefaultFlag.Set(false),
			tax.IsArchived.Set(false),
			tax.CreatedAt.Set(now),
			tax.UpdatedAt.Set(now),
		)
		if err != nil {
			return fmt.Errorf("tax set: %w", err)
		}

		row := db.QueryRow(ctx, InsertTaxStmt,
			tax.TaxID.String,
			tax.Name.String,
			tax.TaxPercentage.Int,
			tax.TaxCategory.String,
			tax.DefaultFlag.Bool,
			tax.IsArchived.Bool,
		)

		var taxID string
		err = row.Scan(&taxID)
		if err != nil {
			return fmt.Errorf("cannot insert tax, err: %s", err)
		}

		stepState.TaxID = taxID
		return nil
	}
}

// CreateBillItem used InsertTaxStmt to insert tax.
// stepState dependency:
//   - stepState.ProductID
//   - stepState.BillingScheduleID
//   - stepState.StudentProductID
//
// stepState assigned:
//   - stepState.BillItemSequenceNumber
//   - stepState.BillItemSequenceNumbers
type BillingItemDescriptionToJSONB struct {
	ProductName string `json:"product_name"`
}

func (c *EntitiesCreator) CreateBillItem(ctx context.Context, db database.QueryExecer, status, billItemType string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		randomStr := idutil.ULIDNow()
		billingStartDate := time.Now()
		billingEndDate := billingStartDate.AddDate(0, 0, 28)
		billingDate := billingStartDate.AddDate(0, -1, 15)

		// Set the bill item entity based on its bounded context
		billItem := &paymentEntities.BillItem{}
		database.AllNullEntity(billItem)

		billingItemDescription := BillingItemDescriptionToJSONB{
			ProductName: fmt.Sprintf("test-product-%s", randomStr),
		}

		err := multierr.Combine(
			billItem.ProductID.Set(stepState.ProductID),
			billItem.ProductDescription.Set(fmt.Sprintf("PRODUCT-DESCRIPTION-%s", randomStr)),
			billItem.ProductPricing.Set(10),
			billItem.DiscountAmountType.Set(fmt.Sprintf("DISCOUNT-AMOUNT-%s", randomStr)),
			billItem.DiscountAmountValue.Set(int64(10)),
			billItem.TaxID.Set(stepState.TaxID),
			billItem.TaxCategory.Set(fmt.Sprintf("TAX-CATEGORY-%s", randomStr)),
			billItem.TaxPercentage.Set(10),
			billItem.OrderID.Set(stepState.OrderID),
			billItem.BillType.Set(billItemType),
			billItem.BillStatus.Set(status),
			billItem.BillDate.Set(billingDate),
			billItem.BillFrom.Set(billingStartDate),
			billItem.BillTo.Set(billingEndDate),
			billItem.BillingItemDescription.Set(database.JSONB(billingItemDescription)),
			billItem.BillSchedulePeriodID.Set(stepState.BillingSchedulePeriodID),
			billItem.DiscountAmount.Set(int64(10)),
			billItem.TaxAmount.Set(int64(10)),
			billItem.FinalPrice.Set(int64(10)),
			billItem.StudentID.Set(stepState.StudentID),
			billItem.BillApprovalStatus.Set(fmt.Sprintf("BILL-APPROVAL-STATUS-%s", randomStr)),
			billItem.LocationID.Set(stepState.LocationID),
			billItem.StudentProductID.Set(stepState.StudentProductID),
		)
		if err != nil {
			return fmt.Errorf("billItem set %v ", err)
		}
		// var count int64
		var exactAmount int64
		switch billItemType {
		case payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String():
			// adjustment price are set to random initially
			adjustmentPrice, err := rand.Int(rand.Reader, big.NewInt(99))
			if err != nil {
				return err
			}

			err = billItem.AdjustmentPrice.Set(adjustmentPrice.Int64())
			if err != nil {
				return err
			}

			exactAmount = billItem.AdjustmentPrice.Int.Int64()
		default:
			exactAmount = billItem.FinalPrice.Int.Int64()
		}

		billItemRepo := paymentRepo.BillItemRepo{}

		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			billItemSequenceNumber, err := billItemRepo.Create(ctx, db, billItem)
			if err == nil {
				stepState.BillItemSequenceNumber = billItemSequenceNumber.Int
				stepState.BillItemSequenceNumbers = append(stepState.BillItemSequenceNumbers, billItemSequenceNumber.Int)
				stepState.InvoiceTotal = exactAmount

				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("billItemRepo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create bill item, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		return nil
	}
}

func (c *EntitiesCreator) CreateDiscount(ctx context.Context, db database.QueryExecer) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		discountID := idutil.ULIDNow()
		now := time.Now().UTC()

		discount := &paymentEntities.Discount{}
		database.AllNullEntity(discount)

		_ = multierr.Combine(
			discount.DiscountID.Set(discountID),
			discount.Name.Set(fmt.Sprintf("DISCOUNT-%s", discountID)),
			discount.DiscountType.Set(fmt.Sprintf("DISCOUNT-TYPE-%s", discountID)),
			discount.DiscountAmountType.Set(fmt.Sprintf("DISCOUNT-AMOUNT-TYPE-%s", discountID)),
			discount.DiscountAmountValue.Set(10),
			discount.RecurringValidDuration.Set(1),
			discount.AvailableFrom.Set(now),
			discount.AvailableUntil.Set(now.Add(24*time.Hour)),
			discount.Remarks.Set(fmt.Sprintf("DISCOUNT-REMARKS-%s", discountID)),
			discount.IsArchived.Set(false),
			discount.UpdatedAt.Set(now),
			discount.CreatedAt.Set(now),
		)

		cmdTag, err := database.InsertExcept(ctx, discount, []string{"resource_path"}, db.Exec)
		if err != nil {
			return fmt.Errorf("err insert Discount: %w", err)
		}

		if cmdTag.RowsAffected() != 1 {
			return fmt.Errorf("err insert Discount: %d RowsAffected", cmdTag.RowsAffected())
		}

		stepState.DiscountID = discountID

		return nil
	}
}

func (c *EntitiesCreator) CreateMigratedBillItem(ctx context.Context, db database.QueryExecer, finalPrice float64, invoiceReference string) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		randomStr := idutil.ULIDNow()
		billingStartDate := time.Now()
		billingEndDate := billingStartDate.AddDate(0, 0, 28)
		billingDate := billingStartDate.AddDate(0, -1, 15)

		// Set the bill item entity based on its bounded context
		billItem := &paymentEntities.BillItem{}
		database.AllNullEntity(billItem)
		now := time.Now()
		err := multierr.Combine(
			billItem.ProductID.Set(stepState.ProductID),
			billItem.ProductDescription.Set(fmt.Sprintf("PRODUCT-DESCRIPTION-%s", randomStr)),
			billItem.ProductPricing.Set(10),
			billItem.DiscountAmountType.Set(fmt.Sprintf("DISCOUNT-AMOUNT-%s", randomStr)),
			billItem.DiscountAmountValue.Set(int64(10)),
			billItem.TaxID.Set(stepState.TaxID),
			billItem.TaxCategory.Set(fmt.Sprintf("TAX-CATEGORY-%s", randomStr)),
			billItem.TaxPercentage.Set(10),
			billItem.OrderID.Set(stepState.OrderID),
			billItem.BillType.Set(payment_pb.BillingType_BILLING_TYPE_BILLED_AT_ORDER.String()),
			billItem.BillStatus.Set(payment_pb.BillingStatus_BILLING_STATUS_BILLED.String()),
			billItem.BillDate.Set(billingDate),
			billItem.BillFrom.Set(billingStartDate),
			billItem.BillTo.Set(billingEndDate),
			billItem.BillSchedulePeriodID.Set(stepState.BillingSchedulePeriodID),
			billItem.DiscountAmount.Set(int64(10)),
			billItem.TaxAmount.Set(int64(10)),
			billItem.FinalPrice.Set(finalPrice),
			billItem.StudentID.Set(stepState.StudentID),
			billItem.BillApprovalStatus.Set(fmt.Sprintf("BILL-APPROVAL-STATUS-%s", randomStr)),
			billItem.Reference.Set(invoiceReference),
			billItem.LocationID.Set(stepState.LocationID),
			billItem.StudentProductID.Set(stepState.StudentProductID),
			billItem.UpdatedAt.Set(now),
			billItem.CreatedAt.Set(now),
		)

		if err != nil {
			return fmt.Errorf("billItem set %v ", err)
		}
		var billItemSequenceNumber pgtype.Int4

		if err := utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			err = database.InsertReturningAndExcept(ctx, billItem, db, []string{"bill_item_sequence_number", "resource_path"}, "bill_item_sequence_number", &billItemSequenceNumber)
			if err == nil {
				stepState.BillItemSequenceNumber = billItemSequenceNumber.Int
				stepState.BillItemSequenceNumbers = append(stepState.BillItemSequenceNumbers, billItemSequenceNumber.Int)
				stepState.InvoiceReferenceID = invoiceReference
				stepState.BillItemTotalFloat = finalPrice

				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("billItemRepo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create bill item, err %v", err)
		}, MaxRetry); err != nil {
			return err
		}

		return nil
	}
}

// CreateBillItem used InsertTaxStmt to insert tax.
// stepState dependency:
//   - stepState.ProductID
//   - stepState.BillingScheduleID
//   - stepState.StudentProductID
//
// stepState assigned:
//   - stepState.BillItemSequenceNumber
//   - stepState.BillItemSequenceNumbers
func (c *EntitiesCreator) CreateBillItemV2(ctx context.Context, db database.QueryExecer, status, billItemType string, amount float64) InsertEntityFunction {
	return func(stepState *common.StepState) error {
		randomStr := idutil.ULIDNow()
		billingStartDate := time.Now()
		billingEndDate := billingStartDate.AddDate(0, 0, 28)
		billingDate := billingStartDate.AddDate(0, -1, 15)

		// Set the bill item entity based on its bounded context
		billItem := &paymentEntities.BillItem{}
		database.AllNullEntity(billItem)

		err := multierr.Combine(
			billItem.ProductID.Set(stepState.ProductID),
			billItem.ProductDescription.Set(fmt.Sprintf("PRODUCT-DESCRIPTION-%s", randomStr)),
			billItem.ProductPricing.Set(10),
			billItem.DiscountAmountType.Set(fmt.Sprintf("DISCOUNT-AMOUNT-%s", randomStr)),
			billItem.DiscountAmountValue.Set(int64(10)),
			billItem.TaxID.Set(stepState.TaxID),
			billItem.TaxCategory.Set(fmt.Sprintf("TAX-CATEGORY-%s", randomStr)),
			billItem.TaxPercentage.Set(10),
			billItem.OrderID.Set(stepState.OrderID),
			billItem.BillType.Set(billItemType),
			billItem.BillStatus.Set(status),
			billItem.BillDate.Set(billingDate),
			billItem.BillFrom.Set(billingStartDate),
			billItem.BillTo.Set(billingEndDate),
			billItem.BillSchedulePeriodID.Set(stepState.BillingSchedulePeriodID),
			billItem.DiscountAmount.Set(int64(10)),
			billItem.TaxAmount.Set(int64(10)),
			billItem.FinalPrice.Set(amount),
			billItem.StudentID.Set(stepState.StudentID),
			billItem.BillApprovalStatus.Set(fmt.Sprintf("BILL-APPROVAL-STATUS-%s", randomStr)),
			billItem.LocationID.Set(stepState.LocationID),
			billItem.StudentProductID.Set(stepState.StudentProductID),
		)
		if err != nil {
			return fmt.Errorf("billItem set %v ", err)
		}
		// var count int64
		var exactAmount int64
		switch billItemType {
		case payment_pb.BillingType_BILLING_TYPE_ADJUSTMENT_BILLING.String():
			// adjustment price are set to random initially
			adjustmentPrice, err := rand.Int(rand.Reader, big.NewInt(99))
			if err != nil {
				return err
			}

			err = billItem.AdjustmentPrice.Set(adjustmentPrice.Int64())
			if err != nil {
				return err
			}

			exactAmount = billItem.AdjustmentPrice.Int.Int64()
		default:
			exactAmount = billItem.FinalPrice.Int.Int64()
		}

		billItemRepo := paymentRepo.BillItemRepo{}

		err = utils.DoWithMaxRetry(func(attempt int) (bool, error) {
			billItemSequenceNumber, err := billItemRepo.Create(ctx, db, billItem)
			if err == nil {
				stepState.BillItemSequenceNumber = billItemSequenceNumber.Int
				stepState.BillItemSequenceNumbers = append(stepState.BillItemSequenceNumbers, billItemSequenceNumber.Int)
				stepState.InvoiceTotal = exactAmount

				return false, nil
			}

			if err != nil && !strings.Contains(err.Error(), "(SQLSTATE 23505)") {
				return false, fmt.Errorf("billItemRepo.Create: %w", err)
			}

			time.Sleep(invoiceConst.DuplicateSleepDuration)
			return attempt < MaxRetry, fmt.Errorf("cannot create bill item, err %v", err)
		}, MaxRetry)

		if err != nil {
			return err
		}

		return nil
	}
}
