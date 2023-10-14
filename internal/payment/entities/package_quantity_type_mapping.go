package entities

import "github.com/jackc/pgtype"

type PackageQuantityTypeMapping struct {
	PackageType  pgtype.Text
	QuantityType pgtype.Text
	CreatedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *PackageQuantityTypeMapping) FieldMap() ([]string, []interface{}) {
	return []string{
			"package_type",
			"quantity_type",
			"created_at",
			"resource_path",
		}, []interface{}{
			&e.PackageType,
			&e.QuantityType,
			&e.CreatedAt,
			&e.ResourcePath,
		}
}

func (e *PackageQuantityTypeMapping) TableName() string {
	return "package_quantity_type_mapping"
}
