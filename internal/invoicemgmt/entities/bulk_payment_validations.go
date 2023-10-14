package entities

import "github.com/jackc/pgtype"

type BulkPaymentValidations struct {
	BulkPaymentValidationsID pgtype.Text
	PaymentMethod            pgtype.Text
	SuccessfulPayments       pgtype.Int4
	FailedPayments           pgtype.Int4
	PendingPayments          pgtype.Int4
	ResourcePath             pgtype.Text
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
	DeletedAt                pgtype.Timestamptz
	ValidationDate           pgtype.Timestamptz
}

func (e *BulkPaymentValidations) FieldMap() ([]string, []interface{}) {
	return []string{
			"bulk_payment_validations_id",
			"payment_method",
			"successful_payments",
			"failed_payments",
			"pending_payments",
			"resource_path",
			"created_at",
			"updated_at",
			"deleted_at",
			"validation_date",
		}, []interface{}{
			&e.BulkPaymentValidationsID,
			&e.PaymentMethod,
			&e.SuccessfulPayments,
			&e.FailedPayments,
			&e.PendingPayments,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ValidationDate,
		}
}

func (e *BulkPaymentValidations) TableName() string {
	return "bulk_payment_validations"
}
