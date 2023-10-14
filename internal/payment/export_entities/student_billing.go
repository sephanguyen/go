package export

import "time"

type StudentBillingExport struct {
	StudentName     string
	StudentID       string
	Grade           string
	Location        string
	CreatedDate     time.Time
	Status          string
	BillingItemName string
	Courses         string
	DiscountName    string
	DiscountAmount  float32
	TaxAmount       float32
	BillingAmount   float32
}

func (e *StudentBillingExport) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_name",
			"student_id",
			"grade",
			"location",
			"created_date",
			"status",
			"billing_item_name",
			"courses",
			"discount_name",
			"discount_amount",
			"tax_amount",
			"billing_amount",
		}, []interface{}{
			&e.StudentName,
			&e.StudentID,
			&e.Grade,
			&e.Location,
			&e.CreatedDate,
			&e.Status,
			&e.BillingItemName,
			&e.Courses,
			&e.DiscountName,
			&e.DiscountAmount,
			&e.TaxAmount,
			&e.BillingAmount,
		}
}

func (e *StudentBillingExport) TableName() string {
	return "student_billing"
}
