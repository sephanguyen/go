package entities

import "github.com/jackc/pgtype"

type ProductAccountingCategory struct {
	ProductID            pgtype.Text
	AccountingCategoryID pgtype.Text
	CreatedAt            pgtype.Timestamptz
	ResourcePath         pgtype.Text
}

func (e *ProductAccountingCategory) FieldMap() ([]string, []interface{}) {
	return []string{
			"product_id",
			"accounting_category_id",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.ProductID,
			&e.AccountingCategoryID,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *ProductAccountingCategory) TableName() string {
	return "product_accounting_category"
}
