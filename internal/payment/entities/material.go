package entities

import "github.com/jackc/pgtype"

type MaterialType int

const (
	MaterialTypeOneTime MaterialType = iota
	MaterialTypeRecurring
)

type Material struct {
	Product           `sql:"-"`
	MaterialID        pgtype.Text
	MaterialType      pgtype.Text
	CustomBillingDate pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (e *Material) FieldMap() ([]string, []interface{}) {
	return []string{
			"material_id",
			"material_type",
			"custom_billing_date",
			"resource_path",
		}, []interface{}{
			&e.MaterialID,
			&e.MaterialType,
			&e.CustomBillingDate,
			&e.ResourcePath,
		}
}

func (e *Material) TableName() string {
	return "material"
}
