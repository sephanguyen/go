package entities

import (
	"github.com/jackc/pgtype"
)

type ProductGroup struct {
	ProductGroupID pgtype.Text
	GroupName      pgtype.Text
	GroupTag       pgtype.Text
	DiscountType   pgtype.Text
	IsArchived     pgtype.Bool
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	ResourcePath   pgtype.Text
}

func (e *ProductGroup) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_group_id",
			"group_name",
			"group_tag",
			"discount_type",
			"is_archived",
			"updated_at",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.ProductGroupID,
			&e.GroupName,
			&e.GroupTag,
			&e.DiscountType,
			&e.IsArchived,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *ProductGroup) TableName() string {
	return "product_group"
}
