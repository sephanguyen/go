package entities

import "github.com/jackc/pgtype"

type StudentPackageClass struct {
	StudentPackageID pgtype.Text
	StudentID        pgtype.Text
	CourseID         pgtype.Text
	LocationID       pgtype.Text
	ClassID          pgtype.Text
	CreatedAt        pgtype.Timestamptz
	UpdatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
}

func (e *StudentPackageClass) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_package_id",
			"student_id",
			"course_id",
			"location_id",
			"class_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.StudentPackageID,
			&e.StudentID,
			&e.CourseID,
			&e.LocationID,
			&e.ClassID,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *StudentPackageClass) TableName() string {
	return "student_package_class"
}
