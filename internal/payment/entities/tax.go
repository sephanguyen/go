package entities

import "github.com/jackc/pgtype"

type Tax struct {
	TaxID         pgtype.Text
	Name          pgtype.Text
	TaxPercentage pgtype.Int4
	TaxCategory   pgtype.Text
	DefaultFlag   pgtype.Bool
	IsArchived    pgtype.Bool
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
}

func (e *Tax) FieldMap() ([]string, []interface{}) {
	return []string{
			"tax_id",
			"name",
			"tax_percentage",
			"tax_category",
			"default_flag",
			"is_archived",
			"created_at",
			"updated_at",
			"resource_path",
		}, []interface{}{
			&e.TaxID,
			&e.Name,
			&e.TaxPercentage,
			&e.TaxCategory,
			&e.DefaultFlag,
			&e.IsArchived,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.ResourcePath,
		}
}

func (e *Tax) TableName() string {
	return "tax"
}
