package entities

import "github.com/jackc/pgtype"

type PackageCourseFee struct {
	PackageID        pgtype.Text
	CourseID         pgtype.Text
	FeeID            pgtype.Text
	AvailableFrom    pgtype.Timestamptz
	AvailableUntil   pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	IsAddedByDefault pgtype.Bool
	ResourcePath     pgtype.Text
}

func (e *PackageCourseFee) FieldMap() ([]string, []interface{}) {
	return []string{
			"package_id",
			"course_id",
			"fee_id",
			"available_from",
			"available_until",
			"created_at",
			"is_added_by_default",
			"resource_path",
		}, []interface{}{
			&e.PackageID,
			&e.CourseID,
			&e.FeeID,
			&e.AvailableFrom,
			&e.AvailableUntil,
			&e.CreatedAt,
			&e.IsAddedByDefault,
			&e.ResourcePath,
		}
}

func (e *PackageCourseFee) TableName() string {
	return "package_course_fee"
}
