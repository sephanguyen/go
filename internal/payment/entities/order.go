package entities

import (
	"github.com/jackc/pgtype"
)

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
	LOAStartDate            pgtype.Timestamptz
	LOAEndDate              pgtype.Timestamptz
	Background              pgtype.Text
	FutureMeasures          pgtype.Text
	VersionNumber           pgtype.Int4
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
			"is_reviewed",
			"withdrawal_effective_date",
			"loa_start_date",
			"loa_end_date",
			"background",
			"future_measures",
			"version_number",
			"order_sequence_number",
			"resource_path",
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
			&e.IsReviewed,
			&e.WithdrawalEffectiveDate,
			&e.LOAStartDate,
			&e.LOAEndDate,
			&e.Background,
			&e.FutureMeasures,
			&e.VersionNumber,
			&e.OrderSequenceNumber,
			&e.ResourcePath,
		}
}

func (e *Order) TableName() string {
	return "order"
}
