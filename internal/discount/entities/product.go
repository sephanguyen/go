package entities

import (
	"github.com/jackc/pgtype"
)

type Product struct {
	ProductID            pgtype.Text
	Name                 pgtype.Text
	ProductType          pgtype.Text
	TaxID                pgtype.Text
	ProductTag           pgtype.Text
	ProductPartnerID     pgtype.Text
	AvailableFrom        pgtype.Timestamptz
	AvailableUntil       pgtype.Timestamptz
	CustomBillingPeriod  pgtype.Timestamptz
	BillingScheduleID    pgtype.Text
	DisableProRatingFlag pgtype.Bool
	Remarks              pgtype.Text
	IsArchived           pgtype.Bool
	IsUnique             pgtype.Bool
	UpdatedAt            pgtype.Timestamptz
	CreatedAt            pgtype.Timestamptz
	ResourcePath         pgtype.Text
}

func (e *Product) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_id",
			"name",
			"product_type",
			"tax_id",
			"product_tag",
			"product_partner_id",
			"available_from",
			"available_until",
			"remarks",
			"custom_billing_period",
			"billing_schedule_id",
			"disable_pro_rating_flag",
			"is_archived",
			"is_unique",
			"updated_at",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.ProductID,
			&e.Name,
			&e.ProductType,
			&e.TaxID,
			&e.ProductTag,
			&e.ProductPartnerID,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.Remarks,
			&e.CustomBillingPeriod,
			&e.BillingScheduleID,
			&e.DisableProRatingFlag,
			&e.IsArchived,
			&e.IsUnique,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *Product) TableName() string {
	return "product"
}
