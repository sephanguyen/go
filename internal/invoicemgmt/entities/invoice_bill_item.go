package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type InvoiceBillItem struct {
	InvoiceBillItemID      pgtype.Text
	InvoiceID              pgtype.Text
	BillItemSequenceNumber pgtype.Int4
	PastBillingStatus      pgtype.Text
	CreatedAt              pgtype.Timestamptz
	ResourcePath           pgtype.Text
	MigratedAt             pgtype.Timestamptz
	UpdatedAt              pgtype.Timestamptz
	DeletedAt              pgtype.Timestamptz
}

func (e *InvoiceBillItem) FieldMap() ([]string, []interface{}) {
	return []string{
			"invoice_bill_item_id",
			"invoice_id",
			"bill_item_sequence_number",
			"past_billing_status",
			"created_at",
			"resource_path",
			"migrated_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.InvoiceBillItemID,
			&e.InvoiceID,
			&e.BillItemSequenceNumber,
			&e.PastBillingStatus,
			&e.CreatedAt,
			&e.ResourcePath,
			&e.MigratedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (*InvoiceBillItem) TableName() string {
	return "invoice_bill_item"
}

// InvoiceBillItems holds invoice bill items added by Add function
// Add function is called by ScanAll
type InvoiceBillItems []*InvoiceBillItem

func (i *InvoiceBillItems) Add() database.Entity {
	e := &InvoiceBillItem{}
	*i = append(*i, e)

	return e
}

// Convert InvoiceBillItems to array of type InvoiceBillItem
func (i InvoiceBillItems) ToArray() []*InvoiceBillItem {
	if len(i) == 0 {
		return []*InvoiceBillItem{}
	}
	return ([]*InvoiceBillItem)(i)
}
