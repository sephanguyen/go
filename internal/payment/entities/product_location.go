package entities

import "github.com/jackc/pgtype"

type ProductLocation struct {
	ProductID    pgtype.Text
	LocationID   pgtype.Text
	CreatedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (p *ProductLocation) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"product_id",
		"location_id",
		"created_at",
		"resource_path",
	}
	values = []interface{}{
		&p.ProductID,
		&p.LocationID,
		&p.CreatedAt,
		&p.ResourcePath,
	}
	return
}

func (p *ProductLocation) TableName() string {
	return "product_location"
}
