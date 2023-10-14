package entities

import "github.com/jackc/pgtype"

type OrderItem struct {
	OrderID          pgtype.Text
	ProductID        pgtype.Text
	OrderItemID      pgtype.Text
	DiscountID       pgtype.Text
	StartDate        pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	StudentProductID pgtype.Text
	ProductName      pgtype.Text
	ResourcePath     pgtype.Text
	EffectiveDate    pgtype.Timestamptz
	CancellationDate pgtype.Timestamptz
	EndDate          pgtype.Timestamptz
}

func (e *OrderItem) FieldMap() ([]string, []interface{}) {
	return []string{
			"order_id",
			"product_id",
			"order_item_id",
			"discount_id",
			"start_date",
			"created_at",
			"student_product_id",
			"product_name",
			"effective_date",
			"cancellation_date",
			"end_date",
			"resource_path",
		}, []interface{}{
			&e.OrderID,
			&e.ProductID,
			&e.OrderItemID,
			&e.DiscountID,
			&e.StartDate,
			&e.CreatedAt,
			&e.StudentProductID,
			&e.ProductName,
			&e.EffectiveDate,
			&e.CancellationDate,
			&e.EndDate,
			&e.ResourcePath,
		}
}

func (e *OrderItem) TableName() string {
	return "order_item"
}
