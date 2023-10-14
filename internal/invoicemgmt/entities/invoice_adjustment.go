package entities

import "github.com/jackc/pgtype"

type InvoiceAdjustment struct {
	InvoiceAdjustmentID             pgtype.Text
	InvoiceID                       pgtype.Text
	Description                     pgtype.Text
	Amount                          pgtype.Numeric
	StudentID                       pgtype.Text
	InvoiceAdjustmentSequenceNumber pgtype.Int4
	CreatedAt                       pgtype.Timestamptz
	UpdatedAt                       pgtype.Timestamptz
	DeletedAt                       pgtype.Timestamptz
	MigratedAt                      pgtype.Timestamptz
	ResourcePath                    pgtype.Text
}

func (e *InvoiceAdjustment) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_adjustment_id",
			"invoice_id",
			"description",
			"amount",
			"student_id",
			"invoice_adjustment_sequence_number",
			"created_at",
			"updated_at",
			"deleted_at",
			"migrated_at",
			"resource_path",
		}, []interface{}{
			&e.InvoiceAdjustmentID,
			&e.InvoiceID,
			&e.Description,
			&e.Amount,
			&e.StudentID,
			&e.InvoiceAdjustmentSequenceNumber,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.MigratedAt,
			&e.ResourcePath,
		}
}

func (e *InvoiceAdjustment) TableName() string {
	return "invoice_adjustment"
}

func (e *InvoiceAdjustment) PrimaryKey() string {
	return "invoice_adjustment_id"
}

func (e *InvoiceAdjustment) UpdateOnConflictQuery() string {
	return `
	description = EXCLUDED.description,
	amount = EXCLUDED.amount,
	updated_at = EXCLUDED.updated_at`
}
