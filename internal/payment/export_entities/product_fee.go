package export

import "time"

type ProductFeeExport struct {
	FeeID                string
	Name                 string
	FeeType              string
	TaxID                string
	ProductTag           string
	ProductPartnerID     string
	AvailableFrom        time.Time
	AvailableUntil       time.Time
	CustomBillingPeriod  time.Time
	BillingScheduleID    string
	DisableProRatingFlag bool
	Remarks              string
	IsUnique             bool
	IsArchived           bool
}

func (e *ProductFeeExport) FieldMap() ([]string, []interface{}) {
	return []string{
			"fee_id",
			"name",
			"fee_type",
			"product_tag",
			"product_partner_id",
			"tax_id",
			"available_from",
			"available_until",
			"custom_billing_period",
			"billing_schedule_id",
			"disable_pro_rating_flag",
			"remarks",
			"is_unique",
			"is_archived",
		}, []interface{}{
			&e.FeeID,
			&e.Name,
			&e.FeeType,
			&e.TaxID,
			&e.ProductTag,
			&e.ProductPartnerID,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.CustomBillingPeriod,
			&e.BillingScheduleID,
			&e.DisableProRatingFlag,
			&e.Remarks,
			&e.IsUnique,
			&e.IsArchived,
		}
}

func (e *ProductFeeExport) TableName() string {
	return "product_fee"
}
