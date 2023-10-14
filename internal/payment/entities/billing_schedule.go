package entities

import "github.com/jackc/pgtype"

type BillingSchedule struct {
	BillingScheduleID pgtype.Text
	Name              pgtype.Text
	Remarks           pgtype.Text
	IsArchived        pgtype.Bool
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (p *BillingSchedule) FieldMap() ([]string, []interface{}) {
	return []string{
			"billing_schedule_id",
			"name",
			"remarks",
			"is_archived",
			"updated_at",
			"created_at",
			"resource_path",
		}, []interface{}{
			&p.BillingScheduleID,
			&p.Name,
			&p.Remarks,
			&p.IsArchived,
			&p.UpdatedAt,
			&p.CreatedAt,
			&p.ResourcePath,
		}
}

func (p *BillingSchedule) TableName() string {
	return "billing_schedule"
}
