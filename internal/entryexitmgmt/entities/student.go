package entities

import "github.com/jackc/pgtype"

type Student struct {
	User `sql:"-"`

	ID           pgtype.Text `sql:"student_id,pk"`
	CurrentGrade pgtype.Int2 `sql:"current_grade"`
	SchoolID     pgtype.Int4 `sql:"school_id"`
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
	GradeID      pgtype.Text
}

// FieldMap return a map of field name and pointer to field
func (e *Student) FieldMap() ([]string, []interface{}) {
	return []string{
			"student_id",
			"current_grade",
			"school_id",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
			"grade_id",
		}, []interface{}{
			&e.ID,
			&e.CurrentGrade,
			&e.SchoolID,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
			&e.GradeID,
		}
}

// TableName returns "students"
func (e *Student) TableName() string {
	return "students"
}
