package entities

import "github.com/jackc/pgtype"

type InvoiceSchedule struct {
	InvoiceScheduleID pgtype.Text
	InvoiceDate       pgtype.Timestamptz
	ScheduledDate     pgtype.Timestamptz
	Status            pgtype.Text
	IsArchived        pgtype.Bool
	Remarks           pgtype.Text
	UserID            pgtype.Text
	ResourcePath      pgtype.Text
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
}

func (e *InvoiceSchedule) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_schedule_id",
			"invoice_date",
			"scheduled_date",
			"status",
			"is_archived",
			"remarks",
			"user_id",
			"resource_path",
			"created_at",
			"updated_at",
		}, []interface{}{
			&e.InvoiceScheduleID,
			&e.InvoiceDate,
			&e.ScheduledDate,
			&e.Status,
			&e.IsArchived,
			&e.Remarks,
			&e.UserID,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
		}
}

func (e *InvoiceSchedule) TableName() string {
	return "invoice_schedule"
}
