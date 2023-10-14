package entities

import "github.com/jackc/pgtype"

type PaymentBankMap struct {
	Payment *Payment
	Bank    *Bank
}
type Payment struct {
	PaymentID             pgtype.Text
	InvoiceID             pgtype.Text
	PaymentMethod         pgtype.Text
	PaymentDueDate        pgtype.Timestamptz
	PaymentExpiryDate     pgtype.Timestamptz
	PaymentDate           pgtype.Timestamptz
	PaymentStatus         pgtype.Text
	ResourcePath          pgtype.Text
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	PaymentSequenceNumber pgtype.Int4
	StudentID             pgtype.Text
	IsExported            pgtype.Bool
	ResultCode            pgtype.Text
	Amount                pgtype.Numeric
	BulkPaymentID         pgtype.Text
	PaymentReferenceID    pgtype.Text
	MigratedAt            pgtype.Timestamptz
	ValidatedDate         pgtype.Timestamptz
	DeletedAt             pgtype.Timestamptz
	ReceiptDate           pgtype.Timestamptz
}

func (e *Payment) FieldMap() ([]string, []interface{}) {
	return []string{
			"payment_id",
			"invoice_id",
			"payment_method",
			"payment_due_date",
			"payment_expiry_date",
			"payment_date",
			"payment_status",
			"resource_path",
			"created_at",
			"updated_at",
			"payment_sequence_number",
			"student_id",
			"is_exported",
			"result_code",
			"amount",
			"bulk_payment_id",
			"payment_reference_id",
			"migrated_at",
			"validated_date",
			"deleted_at",
			"receipt_date",
		}, []interface{}{
			&e.PaymentID,
			&e.InvoiceID,
			&e.PaymentMethod,
			&e.PaymentDueDate,
			&e.PaymentExpiryDate,
			&e.PaymentDate,
			&e.PaymentStatus,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.PaymentSequenceNumber,
			&e.StudentID,
			&e.IsExported,
			&e.ResultCode,
			&e.Amount,
			&e.BulkPaymentID,
			&e.PaymentReferenceID,
			&e.MigratedAt,
			&e.ValidatedDate,
			&e.DeletedAt,
			&e.ReceiptDate,
		}
}

func (*Payment) TableName() string {
	return "payment"
}
