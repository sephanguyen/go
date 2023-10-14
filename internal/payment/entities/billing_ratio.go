package entities

import "github.com/jackc/pgtype"

type BillingRatio struct {
	BillingRatioID          pgtype.Text
	StartDate               pgtype.Timestamptz
	EndDate                 pgtype.Timestamptz
	BillingSchedulePeriodID pgtype.Text
	BillingRatioNumerator   pgtype.Int4
	BillingRatioDenominator pgtype.Int4
	IsArchived              pgtype.Bool
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	ResourcePath            pgtype.Text
}

func (e *BillingRatio) FieldMap() ([]string, []interface{}) {
	return []string{
			"billing_ratio_id",
			"start_date",
			"end_date",
			"billing_schedule_period_id",
			"billing_ratio_numerator",
			"billing_ratio_denominator",
			"is_archived",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&e.BillingRatioID,
			&e.StartDate,
			&e.EndDate,
			&e.BillingSchedulePeriodID,
			&e.BillingRatioNumerator,
			&e.BillingRatioDenominator,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ResourcePath,
		}
}

func (e *BillingRatio) TableName() string {
	return "billing_ratio"
}
