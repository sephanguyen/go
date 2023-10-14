package entities

import "github.com/jackc/pgtype"

type PaymentInvoiceMap struct {
	Payment *Payment
	Invoice *Invoice
}

type PaymentInvoiceUserMap struct {
	Payment       *Payment
	Invoice       *Invoice
	UserBasicInfo *UserBasicInfo
}

type FilePaymentInvoiceMap struct {
	BulkPaymentRequestFilePayment *BulkPaymentRequestFilePayment
	Payment                       *Payment
	Invoice                       *Invoice
}

type BankPaymentInvoiceMap struct {
	Payment     *Payment
	Invoice     *Invoice
	PartnerBank *PartnerBank
}
type BankFilePaymentInvoiceMap struct {
	BulkPaymentRequestFilePayment *BulkPaymentRequestFilePayment
	PaymentInvoiceMap             []*BankPaymentInvoiceMap
}

type BulkPaymentRequestFilePayment struct {
	BulkPaymentRequestFilePaymentID pgtype.Text
	BulkPaymentRequestFileID        pgtype.Text
	PaymentID                       pgtype.Text
	CreatedAt                       pgtype.Timestamptz
	UpdatedAt                       pgtype.Timestamptz
	DeletedAt                       pgtype.Timestamptz
	ResourcePath                    pgtype.Text
}

func (e *BulkPaymentRequestFilePayment) FieldMap() ([]string, []interface{}) {
	return []string{
			"bulk_payment_request_file_payment_id",
			"bulk_payment_request_file_id",
			"payment_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.BulkPaymentRequestFilePaymentID,
			&e.BulkPaymentRequestFileID,
			&e.PaymentID,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *BulkPaymentRequestFilePayment) TableName() string {
	return "bulk_payment_request_file_payment"
}
