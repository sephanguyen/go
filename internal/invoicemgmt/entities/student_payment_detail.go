package entities

import "github.com/jackc/pgtype"

const (
	PaymentMethodDirectDebit      = "DIRECT_DEBIT"
	PaymentMethodConvenienceStore = "CONVENIENCE_STORE"
	PaymentMethodCash             = "CASH"
	PaymentMethodBankTransfer     = "BANK_TRANSFER"
)

type StudentPaymentDetail struct {
	StudentPaymentDetailID pgtype.Text
	StudentID              pgtype.Text
	PayerName              pgtype.Text
	PayerPhoneNumber       pgtype.Text
	PaymentMethod          pgtype.Text
	CreatedAt              pgtype.Timestamptz
	UpdatedAt              pgtype.Timestamptz
	DeletedAt              pgtype.Timestamptz
	MigratedAt             pgtype.Timestamptz
	ResourcePath           pgtype.Text
}

func (e *StudentPaymentDetail) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_payment_detail_id",
			"student_id",
			"payer_name",
			"payer_phone_number",
			"payment_method",
			"created_at",
			"updated_at",
			"deleted_at",
			"migrated_at",
			"resource_path",
		}, []interface{}{
			&e.StudentPaymentDetailID,
			&e.StudentID,
			&e.PayerName,
			&e.PayerPhoneNumber,
			&e.PaymentMethod,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.MigratedAt,
			&e.ResourcePath,
		}
}

func (e *StudentPaymentDetail) TableName() string {
	return "student_payment_detail"
}

type StudentBillingDetailsMap struct {
	StudentPaymentDetail *StudentPaymentDetail
	BillingAddress       *BillingAddress
}

type StudentBankDetailsMap struct {
	StudentPaymentDetail *StudentPaymentDetail
	BankAccount          *BankAccount
}
