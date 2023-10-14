package entities

import "github.com/jackc/pgtype"

type StudentPackageAccessPath struct {
	StudentPackageID pgtype.Text
	StudentID        pgtype.Text
	CourseID         pgtype.Text
	LocationID       pgtype.Text
	AccessPath       pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func (e *StudentPackageAccessPath) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_package_id",
			"student_id",
			"course_id",
			"location_id",
			"access_path",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.StudentPackageID,
			&e.StudentID,
			&e.CourseID,
			&e.LocationID,
			&e.AccessPath,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *StudentPackageAccessPath) TableName() string {
	return "student_package_access_path"
}
