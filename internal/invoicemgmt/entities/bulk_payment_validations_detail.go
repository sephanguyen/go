package entities

import (
	"github.com/jackc/pgtype"
)

type BulkPaymentValidationsDetail struct {
	BulkPaymentValidationsDetailID pgtype.Text
	BulkPaymentValidationsID       pgtype.Text
	InvoiceID                      pgtype.Text
	PaymentID                      pgtype.Text
	ValidatedResultCode            pgtype.Text
	PreviousResultCode             pgtype.Text
	PaymentStatus                  pgtype.Text
	CreatedAt                      pgtype.Timestamptz
	UpdatedAt                      pgtype.Timestamptz
	DeletedAt                      pgtype.Timestamptz
	ResourcePath                   pgtype.Text
}

func (e *BulkPaymentValidationsDetail) FieldMap() ([]string, []interface{}) {
	return []string{
			"bulk_payment_validations_detail_id",
			"bulk_payment_validations_id",
			"invoice_id",
			"payment_id",
			"validated_result_code",
			"previous_result_code",
			"payment_status",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BulkPaymentValidationsDetailID,
			&e.BulkPaymentValidationsID,
			&e.InvoiceID,
			&e.PaymentID,
			&e.ValidatedResultCode,
			&e.PreviousResultCode,
			&e.PaymentStatus,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *BulkPaymentValidationsDetail) TableName() string {
	return "bulk_payment_validations_detail"
}
