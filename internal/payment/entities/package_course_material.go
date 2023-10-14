package entities

import "github.com/jackc/pgtype"

type PackageCourseMaterial struct {
	PackageID        pgtype.Text
	CourseID         pgtype.Text
	MaterialID       pgtype.Text
	AvailableFrom    pgtype.Timestamptz
	AvailableUntil   pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	IsAddedByDefault pgtype.Bool
	ResourcePath     pgtype.Text
}

func (e *PackageCourseMaterial) FieldMap() ([]string, []interface{}) {
	return []string{
			"package_id",
			"course_id",
			"material_id",
			"available_from",
			"available_until",
			"created_at",
			"is_added_by_default",
			"resource_path",
		}, []interface{}{
			&e.PackageID,
			&e.CourseID,
			&e.MaterialID,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.CreatedAt,
			&e.IsAddedByDefault,
			&e.ResourcePath,
		}
}

func (e *PackageCourseMaterial) TableName() string {
	return "package_course_material"
}
