package entities

import "github.com/jackc/pgtype"

type BulkPaymentRequest struct {
	BulkPaymentRequestID pgtype.Text
	PaymentMethod        pgtype.Text
	PaymentDueDateFrom   pgtype.Timestamptz // will not be used. we can delete it in the future
	PaymentDueDateTo     pgtype.Timestamptz // will not be used. we can delete it in the future
	ErrorDetails         pgtype.Text
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
	ResourcePath         pgtype.Text
}

func (e *BulkPaymentRequest) FieldMap() ([]string, []interface{}) {
	return []string{
			"bulk_payment_request_id",
			"payment_method",
			"payment_due_date_from",
			"payment_due_date_to",
			"error_details",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BulkPaymentRequestID,
			&e.PaymentMethod,
			&e.PaymentDueDateFrom,
			&e.PaymentDueDateTo,
			&e.ErrorDetails,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *BulkPaymentRequest) TableName() string {
	return "bulk_payment_request"
}
