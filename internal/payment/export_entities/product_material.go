package export

import "time"

type ProductMaterialExport struct {
	MaterialID           string
	Name                 string
	MaterialType         string
	TaxID                string
	ProductTag           string
	ProductPartnerID     string
	AvailableFrom        time.Time
	AvailableUntil       time.Time
	CustomBillingPeriod  time.Time
	CustomBillingDate    time.Time
	DisableProRatingFlag bool
	BillingScheduleID    string
	Remarks              string
	IsUnique             bool
	IsArchived           bool
}

func (e *ProductMaterialExport) FieldMap() ([]string, []interface{}) {
	return []string{
			"material_id",
			"name",
			"material_type",
			"tax_id",
			"product_tag",
			"product_partner_id",
			"available_from",
			"available_until",
			"custom_billing_period",
			"custom_billing_date",
			"disable_pro_rating_flag",
			"billing_schedule_id",
			"remarks",
			"is_unique",
			"is_archived",
		}, []interface{}{
			&e.MaterialID,
			&e.Name,
			&e.MaterialType,
			&e.TaxID,
			&e.ProductTag,
			&e.ProductPartnerID,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.CustomBillingPeriod,
			&e.CustomBillingDate,
			&e.DisableProRatingFlag,
			&e.BillingScheduleID,
			&e.Remarks,
			&e.IsUnique,
			&e.IsArchived,
		}
}

func (e *ProductMaterialExport) TableName() string {
	return "product_material"
}
