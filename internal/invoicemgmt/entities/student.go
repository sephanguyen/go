package entities

import "github.com/jackc/pgtype"

type Student struct {
	StudentID    pgtype.Text
	CurrentGrade pgtype.Int2
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *Student) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"current_grade",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.StudentID,
			&e.CurrentGrade,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *Student) TableName() string {
	return "students"
}
