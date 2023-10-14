package entities

import "github.com/jackc/pgtype"

type NewCustomerCodeHistory struct {
	NewCustomerCodeHistoryID pgtype.Text
	NewCustomerCode          pgtype.Text
	StudentID                pgtype.Text
	BankAccountNumber        pgtype.Text
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
	DeletedAt                pgtype.Timestamptz
	ResourcePath             pgtype.Text
}

func (e *NewCustomerCodeHistory) FieldMap() ([]string, []interface{}) {
	return []string{
			"new_customer_code_history_id",
			"new_customer_code",
			"student_id",
			"bank_account_number",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.NewCustomerCodeHistoryID,
			&e.NewCustomerCode,
			&e.StudentID,
			&e.BankAccountNumber,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *NewCustomerCodeHistory) TableName() string {
	return "new_customer_code_history"
}
