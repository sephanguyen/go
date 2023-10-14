package entities

import (
	"github.com/jackc/pgtype"
)

type ProductGroup struct {
	ProductGroupID pgtype.Text
	GroupName      pgtype.Text
	GroupTag       pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	ResourcePath   pgtype.Text
	DiscountType   pgtype.Text
	IsArchived     pgtype.Bool
}

func (e *ProductGroup) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_group_id",
			"group_name",
			"group_tag",
			"updated_at",
			"created_at",
			"resource_path",
			"discount_type",
			"is_archived",
		}, []interface{}{
			&e.ProductGroupID,
			&e.GroupName,
			&e.GroupTag,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.ResourcePath,
			&e.DiscountType,
			&e.IsArchived,
		}
}

func (e *ProductGroup) TableName() string {
	return "product_group"
}
