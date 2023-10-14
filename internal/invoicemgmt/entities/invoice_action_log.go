package entities

import "github.com/jackc/pgtype"

type InvoiceActionLog struct {
	InvoiceActionID          pgtype.Text
	InvoiceID                pgtype.Text
	UserID                   pgtype.Text
	Action                   pgtype.Text
	ActionDetail             pgtype.Text
	PaymentSequenceNumber    pgtype.Int4
	ActionComment            pgtype.Text
	BulkPaymentValidationsID pgtype.Text
	ResourcePath             pgtype.Text
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
}

func (e *InvoiceActionLog) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_action_id",
			"invoice_id",
			"user_id",
			"action",
			"action_detail",
			"payment_sequence_number",
			"action_comment",
			"bulk_payment_validations_id",
			"resource_path",
			"created_at",
			"updated_at",
		}, []interface{}{
			&e.InvoiceActionID,
			&e.InvoiceID,
			&e.UserID,
			&e.Action,
			&e.ActionDetail,
			&e.PaymentSequenceNumber,
			&e.ActionComment,
			&e.BulkPaymentValidationsID,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
		}
}

func (*InvoiceActionLog) TableName() string {
	return "invoice_action_log"
}
