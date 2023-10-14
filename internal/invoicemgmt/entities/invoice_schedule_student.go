package entities

import "github.com/jackc/pgtype"

type InvoiceScheduleStudent struct {
	InvoiceSchedulesStudentID pgtype.Text
	InvoiceScheduleHistoryID  pgtype.Text
	StudentID                 pgtype.Text
	ResourcePath              pgtype.Text
	ErrorDetails              pgtype.Text
	ActualErrorDetails        pgtype.Text
	CreatedAt                 pgtype.Timestamptz
}

func (e *InvoiceScheduleStudent) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_schedule_student_id",
			"invoice_schedule_history_id",
			"student_id",
			"resource_path",
			"error_details",
			"actual_error_details",
			"created_at",
		}, []interface{}{
			&e.InvoiceSchedulesStudentID,
			&e.InvoiceScheduleHistoryID,
			&e.StudentID,
			&e.ResourcePath,
			&e.ErrorDetails,
			&e.ActualErrorDetails,
			&e.CreatedAt,
		}
}

func (e *InvoiceScheduleStudent) TableName() string {
	return "invoice_schedule_student"
}
