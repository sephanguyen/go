package entities

import "github.com/jackc/pgtype"

type Student struct {
	StudentID        pgtype.Text
	CurrentGrade     pgtype.Int2
	EnrollmentStatus pgtype.Text
	UpdatedAt        pgtype.Timestamptz
	CreatedAt        pgtype.Timestamptz
	DeletedAt        pgtype.Timestamptz
	GradeID          pgtype.Text
	ResourcePath     pgtype.Text
}

func (e *Student) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"current_grade",
			"enrollment_status",
			"updated_at",
			"created_at",
			"deleted_at",
			"grade_id",
			"resource_path",
		}, []interface{}{
			&e.StudentID,
			&e.CurrentGrade,
			&e.EnrollmentStatus,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.GradeID,
			&e.ResourcePath,
		}
}

func (e *Student) TableName() string {
	return "students"
}
