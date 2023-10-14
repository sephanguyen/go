package entities

import "github.com/jackc/pgtype"

type BillItemAccountCategory struct {
	BillItemSequenceNumber pgtype.Int4
	AccountCategoryID      pgtype.Text
	ResourcePath           pgtype.Text
}

func (e *BillItemAccountCategory) FieldMap() ([]string, []interface{}) {
	return []string{
			"bill_item_sequence_number",
			"accounting_category_id",
			"resource_path",
		}, []interface{}{
			&e.BillItemSequenceNumber,
			&e.AccountCategoryID,
			&e.ResourcePath,
		}
}

func (e *BillItemAccountCategory) TableName() string {
	return "bill_item_account_category"
}
