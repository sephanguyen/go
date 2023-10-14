package entities

import "github.com/jackc/pgtype"

type BillingSchedulePeriod struct {
	BillingSchedulePeriodID pgtype.Text
	Name                    pgtype.Text
	BillingScheduleID       pgtype.Text
	StartDate               pgtype.Timestamptz
	EndDate                 pgtype.Timestamptz
	BillingDate             pgtype.Timestamptz
	Remarks                 pgtype.Text
	IsArchived              pgtype.Bool
	UpdatedAt               pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	ResourcePath            pgtype.Text
}

func (e *BillingSchedulePeriod) FieldMap() ([]string, []interface{}) {
	return []string{
			"billing_schedule_period_id",
			"name",
			"billing_schedule_id",
			"start_date",
			"end_date",
			"billing_date",
			"remarks",
			"is_archived",
			"updated_at",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.BillingSchedulePeriodID,
			&e.Name,
			&e.BillingScheduleID,
			&e.StartDate,
			&e.EndDate,
			&e.BillingDate,
			&e.Remarks,
			&e.IsArchived,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *BillingSchedulePeriod) TableName() string {
	return "billing_schedule_period"
}
