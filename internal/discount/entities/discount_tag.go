package entities

import "github.com/jackc/pgtype"

type DiscountTag struct {
	DiscountTagID   pgtype.Text
	DiscountTagName pgtype.Text
	Selectable      pgtype.Bool
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
	IsArchived      pgtype.Bool
	ResourcePath    pgtype.Text
}

func (e *DiscountTag) FieldMap() ([]string, []interface{}) {
	return []string{
			"discount_tag_id",
			"discount_tag_name",
			"selectable",
			"created_at",
			"updated_at",
			"is_archived",
			"resource_path",
		}, []interface{}{
			&e.DiscountTagID,
			&e.DiscountTagName,
			&e.Selectable,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.IsArchived,
			&e.ResourcePath,
		}
}

func (e *DiscountTag) TableName() string {
	return "discount_tag"
}
