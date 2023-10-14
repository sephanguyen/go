package entities

import "github.com/jackc/pgtype"

type Fee struct {
	Product      `sql:"-"`
	FeeID        pgtype.Text
	FeeType      pgtype.Text
	ResourcePath pgtype.Text
}

func (e *Fee) FieldMap() ([]string, []interface{}) {
	return []string{
			"fee_id",
			"fee_type",
			"resource_path",
		}, []interface{}{
			&e.FeeID,
			&e.FeeType,
			&e.ResourcePath,
		}
}

func (e *Fee) TableName() string {
	return "fee"
}
