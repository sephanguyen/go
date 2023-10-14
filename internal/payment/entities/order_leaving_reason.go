package entities

import "github.com/jackc/pgtype"

type OrderLeavingReason struct {
	OrderID         pgtype.Text
	LeavingReasonID pgtype.Text
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	ResourcePath    pgtype.Text
}

func (e *OrderLeavingReason) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"order_id",
		"leaving_reason_id",
		"created_at",
		"updated_at",
		"resource_path",
	}
	values = []interface{}{
		&e.OrderID,
		&e.LeavingReasonID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.ResourcePath,
	}
	return
}

func (e *OrderLeavingReason) TableName() string {
	return "order_leaving_reason"
}
