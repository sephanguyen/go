package entities

import "github.com/jackc/pgtype"

type Package struct {
	Product          `sql:"-"`
	PackageID        pgtype.Text
	PackageType      pgtype.Text
	MaxSlot          pgtype.Int4
	PackageStartDate pgtype.Timestamptz
	PackageEndDate   pgtype.Timestamptz
	ResourcePath     pgtype.Text
}

func (e *Package) FieldMap() ([]string, []interface{}) {
	return []string{
			"package_id",
			"package_type",
			"max_slot",
			"package_start_date",
			"package_end_date",
			"resource_path",
		}, []interface{}{
			&e.PackageID,
			&e.PackageType,
			&e.MaxSlot,
			&e.PackageStartDate,
			&e.PackageEndDate,
			&e.ResourcePath,
		}
}

func (e *Package) TableName() string {
	return "package"
}
