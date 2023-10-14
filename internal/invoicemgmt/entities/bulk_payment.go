package entities

import "github.com/jackc/pgtype"

type BulkPayment struct {
	BulkPaymentID     pgtype.Text
	BulkPaymentStatus pgtype.Text
	PaymentMethod     pgtype.Text
	InvoiceStatus     pgtype.Text
	InvoiceType       pgtype.TextArray
	PaymentStatus     pgtype.TextArray
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (e *BulkPayment) FieldMap() ([]string, []interface{}) {
	return []string{
			"bulk_payment_id",
			"bulk_payment_status",
			"payment_method",
			"invoice_status",
			"invoice_type",
			"payment_status",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BulkPaymentID,
			&e.BulkPaymentStatus,
			&e.PaymentMethod,
			&e.InvoiceStatus,
			&e.InvoiceType,
			&e.PaymentStatus,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *BulkPayment) TableName() string {
	return "bulk_payment"
}
