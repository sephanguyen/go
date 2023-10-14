package entities

import "github.com/jackc/pgtype"

type ProductDiscount struct {
	DiscountID   pgtype.Text
	ProductID    pgtype.Text
	CreatedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *ProductDiscount) FieldMap() ([]string, []interface{}) {
	return []string{
			"discount_id",
			"product_id",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.DiscountID,
			&e.ProductID,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *ProductDiscount) TableName() string {
	return "product_discount"
}
