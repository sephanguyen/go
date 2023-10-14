package entities

import "github.com/jackc/pgtype"

type AccountingCategory struct {
	AccountingCategoryID pgtype.Text
	Name                 pgtype.Text
	Remarks              pgtype.Text
	IsArchived           pgtype.Bool
	UpdatedAt            pgtype.Timestamptz
	CreatedAt            pgtype.Timestamptz
	ResourcePath         pgtype.Text
}

func (e *AccountingCategory) FieldMap() ([]string, []interface{}) {
	return []string{
			"accounting_category_id",
			"name",
			"remarks",
			"is_archived",
			"updated_at",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.AccountingCategoryID,
			&e.Name,
			&e.Remarks,
			&e.IsArchived,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *AccountingCategory) TableName() string {
	return "accounting_category"
}
