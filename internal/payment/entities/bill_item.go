package entities

import (
	"time"

	"github.com/jackc/pgtype"
)

type BillItem struct {
	BillItemSequenceNumber         pgtype.Int4
	StudentID                      pgtype.Text
	StudentProductID               pgtype.Text
	OrderID                        pgtype.Text
	BillType                       pgtype.Text
	BillStatus                     pgtype.Text
	BillDate                       pgtype.Timestamptz
	BillFrom                       pgtype.Timestamptz
	BillTo                         pgtype.Timestamptz
	BillingItemDescription         pgtype.JSONB
	BillSchedulePeriodID           pgtype.Text
	ProductID                      pgtype.Text
	ProductDescription             pgtype.Text
	ProductPricing                 pgtype.Int4
	DiscountAmountType             pgtype.Text
	DiscountAmountValue            pgtype.Numeric
	DiscountAmount                 pgtype.Numeric
	RawDiscountAmount              pgtype.Numeric
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
	PreviousBillItemSequenceNumber pgtype.Int4
	PreviousBillItemStatus         pgtype.Text
	AdjustmentPrice                pgtype.Numeric
	IsLatestBillItem               pgtype.Bool
	Price                          pgtype.Numeric
	OldPrice                       pgtype.Numeric
	BillingRatioNumerator          pgtype.Int4
	BillingRatioDenominator        pgtype.Int4
	DiscountID                     pgtype.Text
	Reference                      pgtype.Text
}

func (e *BillItem) FieldMap() ([]string, []interface{}) {
	return []string{
			"order_id",
			"student_id",
			"product_id",
			"student_product_id",
			"bill_type",
			"billing_status",
			"billing_date",
			"billing_from",
			"billing_to",
			"billing_schedule_period_id",
			"product_description",
			"product_pricing",
			"discount_amount_type",
			"discount_amount_value",
			"discount_amount",
			"raw_discount_amount",
			"tax_id",
			"tax_category",
			"tax_percentage",
			"tax_amount",
			"final_price",
			"updated_at",
			"created_at",
			"bill_item_sequence_number",
			"billing_item_description",
			"resource_path",
			"billing_approval_status",
			"location_id",
			"location_name",
			"previous_bill_item_sequence_number",
			"previous_bill_item_status",
			"adjustment_price",
			"is_latest_bill_item",
			"price",
			"old_price",
			"billing_ratio_numerator",
			"billing_ratio_denominator",
			"discount_id",
			"reference",
		}, []interface{}{
			&e.OrderID,
			&e.StudentID,
			&e.ProductID,
			&e.StudentProductID,
			&e.BillType,
			&e.BillStatus,
			&e.BillDate,
			&e.BillFrom,
			&e.BillTo,
			&e.BillSchedulePeriodID,
			&e.ProductDescription,
			&e.ProductPricing,
			&e.DiscountAmountType,
			&e.DiscountAmountValue,
			&e.DiscountAmount,
			&e.RawDiscountAmount,
			&e.TaxID,
			&e.TaxCategory,
			&e.TaxPercentage,
			&e.TaxAmount,
			&e.FinalPrice,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.BillItemSequenceNumber,
			&e.BillingItemDescription,
			&e.ResourcePath,
			&e.BillApprovalStatus,
			&e.LocationID,
			&e.LocationName,
			&e.PreviousBillItemSequenceNumber,
			&e.PreviousBillItemStatus,
			&e.AdjustmentPrice,
			&e.IsLatestBillItem,
			&e.Price,
			&e.OldPrice,
			&e.BillingRatioNumerator,
			&e.BillingRatioDenominator,
			&e.DiscountID,
			&e.Reference,
		}
}

type BillingItemDescription struct {
	ProductID               string        `json:"product_id"`
	ProductName             string        `json:"product_name"`
	ProductType             string        `json:"product_type"`
	PackageType             *string       `json:"package_type"`
	MaterialType            *string       `json:"material_type"`
	FeeType                 *string       `json:"fee_type"`
	QuantityType            *string       `json:"quantity_type"`
	BillingPeriod           *time.Time    `json:"billing_period"`
	BillingRatio            *string       `json:"billing_ratio"`
	CourseItems             []*CourseItem `json:"course_items"`
	BillingPeriodName       *string       `json:"billing_period_name"`
	BillingScheduleName     *string       `json:"billing_schedule_name"`
	BillingRatioNumerator   *int32        `json:"billing_ratio_numerator"`
	BillingRatioDenominator *int32        `json:"billing_ratio_denominator"`
	DiscountName            *string       `json:"discount_name"`
	GradeID                 string        `json:"grade_id"`
	GradeName               string        `json:"grade_name"`
}

type CourseItem struct {
	CourseID   string `json:"course_id"`
	CourseName string `json:"course_name"`
	Weight     *int32 `json:"weight"`
	Slot       *int32 `json:"slot"`
}

func (e *BillItem) GetBillingItemDescription() (*BillingItemDescription, error) {
	pp := &BillingItemDescription{}
	err := e.BillingItemDescription.AssignTo(pp)
	return pp, err
}

func (e *BillItem) TableName() string {
	return "bill_item"
}
