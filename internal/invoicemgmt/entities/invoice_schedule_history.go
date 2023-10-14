package entities

import "github.com/jackc/pgtype"

type InvoiceScheduleHistory struct {
	InvoiceScheduleHistoryID pgtype.Text
	InvoiceScheduleID        pgtype.Text
	NumberOfFailedInvoices   pgtype.Int4
	TotalStudents            pgtype.Int4
	ExecutionStartDate       pgtype.Timestamptz
	ExecutionEndDate         pgtype.Timestamptz
	ResourcePath             pgtype.Text
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
}

func (e *InvoiceScheduleHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_schedule_history_id",
			"invoice_schedule_id",
			"number_of_failed_invoices",
			"total_students",
			"execution_start_date",
			"execution_end_date",
			"resource_path",
			"created_at",
			"updated_at",
		}, []interface{}{
			&e.InvoiceScheduleHistoryID,
			&e.InvoiceScheduleID,
			&e.NumberOfFailedInvoices,
			&e.TotalStudents,
			&e.ExecutionStartDate,
			&e.ExecutionEndDate,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
		}
}

func (e *InvoiceScheduleHistory) TableName() string {
	return "invoice_schedule_history"
}
