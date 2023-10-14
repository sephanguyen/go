package entities

import (
	"github.com/jackc/pgtype"
)

type InvoiceBillItemMap struct {
	InvoiceID              pgtype.Text
	BillItemSequenceNumber pgtype.Int4
	BillingItemDescription pgtype.JSONB
	FinalPrice             pgtype.Numeric
	AdjustmentPrice        pgtype.Numeric
	BillType               pgtype.Text
}

type BillItem struct {
	BillItemSequenceNumber         pgtype.Int4
	StudentID                      pgtype.Text
	OrderID                        pgtype.Text
	BillType                       pgtype.Text
	BillStatus                     pgtype.Text
	BillDate                       pgtype.Timestamptz
	BillFrom                       pgtype.Timestamptz
	BillTo                         pgtype.Timestamptz
	BillingItemDescription         pgtype.JSONB
	BillSchedulePeriodID           pgtype.Text
	ProductDescription             pgtype.Text
	ProductPricing                 pgtype.Int4
	DiscountAmountType             pgtype.Text
	DiscountAmountValue            pgtype.Numeric
	DiscountAmount                 pgtype.Numeric
	TaxID                          pgtype.Text
	TaxCategory                    pgtype.Text
	TaxAmount                      pgtype.Numeric
	TaxPercentage                  pgtype.Int4
	FinalPrice                     pgtype.Numeric
	UpdatedAt                      pgtype.Timestamptz
	CreatedAt                      pgtype.Timestamptz
	ResourcePath                   pgtype.Text
	BillApprovalStatus             pgtype.Text
	LocationID                     pgtype.Text
	LocationName                   pgtype.Text
	ProductID                      pgtype.Text
	StudentProductID               pgtype.Text
	PreviousBillItemSequenceNumber pgtype.Int4
	PreviousBillItemStatus         pgtype.Text
	AdjustmentPrice                pgtype.Numeric
	IsLatestBillItem               pgtype.Bool
	Price                          pgtype.Numeric
	OldPrice                       pgtype.Numeric
	BillingRatioNumerator          pgtype.Int4
	BillingRatioDenominator        pgtype.Int4
	DiscountID                     pgtype.Text
	IsReviewed                     pgtype.Bool
	RawDiscountAmount              pgtype.Numeric
	Reference                      pgtype.Text
}

func (e *BillItem) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_description",
			"product_pricing",
			"discount_amount_type",
			"discount_amount_value",
			"tax_id",
			"tax_category",
			"tax_percentage",
			"resource_path",
			"order_id",
			"bill_type",
			"billing_status",
			"billing_date",
			"billing_from",
			"billing_to",
			"billing_schedule_period_id",
			"bill_item_sequence_number",
			"discount_amount",
			"tax_amount",
			"final_price",
			"student_id",
			"billing_approval_status",
			"billing_item_description",
			"created_at",
			"updated_at",
			"location_id",
			"location_name",
			"product_id",
			"student_product_id",
			"previous_bill_item_sequence_number",
			"previous_bill_item_status",
			"adjustment_price",
			"is_latest_bill_item",
			"price",
			"old_price",
			"billing_ratio_numerator",
			"billing_ratio_denominator",
			"discount_id",
			"is_reviewed",
			"raw_discount_amount",
			"reference",
		}, []interface{}{
			&e.ProductDescription,
			&e.ProductPricing,
			&e.DiscountAmountType,
			&e.DiscountAmountValue,
			&e.TaxID,
			&e.TaxCategory,
			&e.TaxPercentage,
			&e.ResourcePath,
			&e.OrderID,
			&e.BillType,
			&e.BillStatus,
			&e.BillDate,
			&e.BillFrom,
			&e.BillTo,
			&e.BillSchedulePeriodID,
			&e.BillItemSequenceNumber,
			&e.DiscountAmount,
			&e.TaxAmount,
			&e.FinalPrice,
			&e.StudentID,
			&e.BillApprovalStatus,
			&e.BillingItemDescription,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.LocationID,
			&e.LocationName,
			&e.ProductID,
			&e.StudentProductID,
			&e.PreviousBillItemSequenceNumber,
			&e.PreviousBillItemStatus,
			&e.AdjustmentPrice,
			&e.IsLatestBillItem,
			&e.Price,
			&e.OldPrice,
			&e.BillingRatioNumerator,
			&e.BillingRatioDenominator,
			&e.DiscountID,
			&e.IsReviewed,
			&e.RawDiscountAmount,
			&e.Reference,
		}
}

func (e *BillItem) TableName() string {
	return "bill_item"
}
