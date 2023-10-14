package export

type InvoiceScheduleExportData struct {
	InvoiceScheduleID string
	InvoiceDate       string
	IsArchived        bool
	Remarks           string
}

func (e *InvoiceScheduleExportData) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_schedule_id",
			"invoice_date",
			"is_archived",
			"remarks",
		}, []interface{}{
			&e.InvoiceScheduleID,
			&e.InvoiceDate,
			&e.IsArchived,
			&e.Remarks,
		}
}

func (e *InvoiceScheduleExportData) TableName() string {
	return "invoice_schedule"
}
