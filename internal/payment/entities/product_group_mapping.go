package entities

import (
	"github.com/jackc/pgtype"
)

type ProductGroupMapping struct {
	ProductGroupID pgtype.Text
	ProductID      pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	ResourcePath   pgtype.Text
}

func (e *ProductGroupMapping) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_group_id",
			"product_id",
			"updated_at",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.ProductGroupID,
			&e.ProductID,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *ProductGroupMapping) TableName() string {
	return "product_group_mapping"
}
