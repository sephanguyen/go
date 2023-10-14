package entities

import (
	"github.com/jackc/pgtype"
)

type InvoiceCreditNote struct {
	CreditNoteID             pgtype.Text
	InvoiceID                pgtype.Text
	CreditNoteSequenceNumber pgtype.Int4
	Reason                   pgtype.Text
	Price                    pgtype.Numeric
	ResourcePath             pgtype.Text
	CreatedAt                pgtype.Timestamptz
	UpdatedAt                pgtype.Timestamptz
}

func (e *InvoiceCreditNote) FieldMap() ([]string, []interface{}) {
	return []string{
			"credit_note_id",
			"invoice_id",
			"credit_note_sequence_number",
			"reason",
			"price",
			"resource_path",
			"created_at",
			"updated_at",
		}, []interface{}{
			&e.CreditNoteID,
			&e.InvoiceID,
			&e.CreditNoteSequenceNumber,
			&e.Reason,
			&e.Price,
			&e.ResourcePath,
			&e.CreatedAt,
			&e.UpdatedAt,
		}
}

func (*InvoiceCreditNote) TableName() string {
	return "invoice_credit_note"
}
