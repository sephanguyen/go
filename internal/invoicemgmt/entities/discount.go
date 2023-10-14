package entities

import "github.com/jackc/pgtype"

type Discount struct {
	DiscountID             pgtype.Text
	Name                   pgtype.Text
	DiscountType           pgtype.Text
	DiscountAmountType     pgtype.Text
	DiscountAmountValue    pgtype.Numeric
	RecurringValidDuration pgtype.Int4
	AvailableFrom          pgtype.Timestamptz
	AvailableUntil         pgtype.Timestamptz
	Remarks                pgtype.Text
	IsArchived             pgtype.Bool
	UpdatedAt              pgtype.Timestamptz
	CreatedAt              pgtype.Timestamptz
	ResourcePath           pgtype.Text
}

func (e *Discount) FieldMap() ([]string, []interface{}) {
	return []string{
			"discount_id",
			"name",
			"discount_type",
			"discount_amount_type",
			"discount_amount_value",
			"recurring_valid_duration",
			"available_from",
			"available_until",
			"remarks",
			"is_archived",
			"updated_at",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.DiscountID,
			&e.Name,
			&e.DiscountType,
			&e.DiscountAmountType,
			&e.DiscountAmountValue,
			&e.RecurringValidDuration,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.Remarks,
			&e.IsArchived,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *Discount) TableName() string {
	return "discount"
}
