package entities

import "github.com/jackc/pgtype"

type Order struct {
	OrderID                 pgtype.Text
	StudentID               pgtype.Text
	StudentFullName         pgtype.Text
	LocationID              pgtype.Text
	OrderSequenceNumber     pgtype.Int4
	OrderComment            pgtype.Text
	OrderStatus             pgtype.Text
	OrderType               pgtype.Text
	UpdatedAt               pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	ResourcePath            pgtype.Text
	IsReviewed              pgtype.Bool
	WithdrawalEffectiveDate pgtype.Timestamptz
}

func (e *Order) FieldMap() ([]string, []interface{}) {
	return []string{
			"order_id",
			"student_id",
			"student_full_name",
			"location_id",
			"order_comment",
			"order_status",
			"order_type",
			"updated_at",
			"created_at",
			"order_sequence_number",
			"resource_path",
			"is_reviewed",
			"withdrawal_effective_date",
		}, []interface{}{
			&e.OrderID,
			&e.StudentID,
			&e.StudentFullName,
			&e.LocationID,
			&e.OrderComment,
			&e.OrderStatus,
			&e.OrderType,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.OrderSequenceNumber,
			&e.ResourcePath,
			&e.IsReviewed,
			&e.WithdrawalEffectiveDate,
		}
}

func (e *Order) TableName() string {
	return "order"
}
