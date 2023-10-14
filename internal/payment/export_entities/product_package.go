package export

import (
	"time"
)

type ProductPackageExport struct {
	PackageID            string
	Name                 string
	PackageType          string
	TaxID                string
	ProductTag           string
	ProductPartnerID     string
	AvailableFrom        time.Time
	AvailableUntil       time.Time
	MaxSlot              int32
	CustomBillingPeriod  time.Time
	BillingScheduleID    string
	DisableProRatingFlag bool
	PackageStartDate     time.Time
	PackageEndDate       time.Time
	Remarks              string
	IsArchived           bool
	IsUnique             bool
}

func (e *ProductPackageExport) FieldMap() ([]string, []interface{}) {
	return []string{
			"package_id",
			"name",
			"package_type",
			"tax_id",
			"product_tag",
			"product_partner_id",
			"available_from",
			"available_until",
			"max_slot",
			"custom_billing_period",
			"billing_schedule_id",
			"disable_pro_rating_flag",
			"package_start_date",
			"package_end_date",
			"remarks",
			"is_archived",
			"is_unique",
		}, []interface{}{
			&e.PackageID,
			&e.Name,
			&e.PackageType,
			&e.TaxID,
			&e.ProductTag,
			&e.ProductPartnerID,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.MaxSlot,
			&e.CustomBillingPeriod,
			&e.BillingScheduleID,
			&e.DisableProRatingFlag,
			&e.PackageStartDate,
			&e.PackageEndDate,
			&e.Remarks,
			&e.IsArchived,
			&e.IsUnique,
		}
}

func (e *ProductPackageExport) TableName() string {
	return "product_package"
}
