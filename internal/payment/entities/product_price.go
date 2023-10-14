package entities

import "github.com/jackc/pgtype"

type ProductPrice struct {
	ProductPriceID          pgtype.Int4
	ProductID               pgtype.Text
	BillingSchedulePeriodID pgtype.Text
	Quantity                pgtype.Int4
	Price                   pgtype.Numeric
	PriceType               pgtype.Text
	CreatedAt               pgtype.Timestamptz
	ResourcePath            pgtype.Text
}

func (e *ProductPrice) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_price_id",
			"product_id",
			"billing_schedule_period_id",
			"quantity",
			"price",
			"price_type",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.ProductPriceID,
			&e.ProductID,
			&e.BillingSchedulePeriodID,
			&e.Quantity,
			&e.Price,
			&e.PriceType,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *ProductPrice) TableName() string {
	return "product_price"
}
