package entities

import (
	"github.com/jackc/pgtype"
)

type UpcomingBillItem struct {
	OrderID                 pgtype.Text
	BillItemSequenceNumber  pgtype.Int4
	ProductID               pgtype.Text
	StudentProductID        pgtype.Text
	ProductDescription      pgtype.Text
	DiscountID              pgtype.Text
	TaxID                   pgtype.Text
	BillingSchedulePeriodID pgtype.Text
	BillingDate             pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
	IsGenerated             pgtype.Bool
	ExecuteNote             pgtype.Text
	ResourcePath            pgtype.Text
}

func (e *UpcomingBillItem) FieldMap() ([]string, []interface{}) {
	return []string{
			"order_id",
			"bill_item_sequence_number",
			"product_id",
			"student_product_id",
			"product_description",
			"discount_id",
			"tax_id",
			"billing_schedule_period_id",
			"billing_date",
			"created_at",
			"updated_at",
			"deleted_at",
			"is_generated",
			"execute_note",
			"resource_path",
		}, []interface{}{
			&e.OrderID,
			&e.BillItemSequenceNumber,
			&e.ProductID,
			&e.StudentProductID,
			&e.ProductDescription,
			&e.DiscountID,
			&e.TaxID,
			&e.BillingSchedulePeriodID,
			&e.BillingDate,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.IsGenerated,
			&e.ExecuteNote,
			&e.ResourcePath,
		}
}

func (e *UpcomingBillItem) TableName() string {
	return "upcoming_bill_item"
}
