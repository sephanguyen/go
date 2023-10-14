package entities

import "github.com/jackc/pgtype"

type Invoice struct {
	InvoiceID             pgtype.Text
	Type                  pgtype.Text
	Status                pgtype.Text
	StudentID             pgtype.Text
	SubTotal              pgtype.Numeric
	Total                 pgtype.Numeric
	ResourcePath          pgtype.Text
	CreatedAt             pgtype.Timestamptz
	UpdatedAt             pgtype.Timestamptz
	IsExported            pgtype.Bool
	OutstandingBalance    pgtype.Numeric
	AmountPaid            pgtype.Numeric
	AmountRefunded        pgtype.Numeric
	InvoiceSequenceNumber pgtype.Int4
	InvoiceReferenceID    pgtype.Text
	MigratedAt            pgtype.Timestamptz
	InvoiceReferenceID2   pgtype.Text
	DeletedAt             pgtype.Timestamptz
}

func (e *Invoice) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_id",
			"type",
			"status",
			"student_id",
			"sub_total",
			"total",
			"resource_path",
			"created_at",
			"updated_at",
			"is_exported",
			"outstanding_balance",
			"amount_paid",
			"amount_refunded",
			"invoice_sequence_number",
			"invoice_reference_id",
			"migrated_at",
			"invoice_reference_id2",
			"deleted_at",
		}, []interface{}{
			&e.InvoiceID,
			&e.Type,
			&e.Status,
			&e.StudentID,
			&e.SubTotal,
			&e.Total,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.IsExported,
			&e.OutstandingBalance,
			&e.AmountPaid,
			&e.AmountRefunded,
			&e.InvoiceSequenceNumber,
			&e.InvoiceReferenceID,
			&e.MigratedAt,
			&e.InvoiceReferenceID2,
			&e.DeletedAt,
		}
}

func (*Invoice) TableName() string {
	return "invoice"
}

type InvoicePaymentMap struct {
	Invoice       *Invoice
	Payment       *Payment
	UserBasicInfo *UserBasicInfo
}
