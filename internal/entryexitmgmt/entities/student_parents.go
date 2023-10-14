package entities

import "github.com/jackc/pgtype"

type StudentParent struct {
	StudentID    pgtype.Text
	ParentID     pgtype.Text
	Relationship pgtype.Text
	UpdatedAt    pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *StudentParent) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"student_id",
			"parent_id",
			"relationship",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.StudentID,
			&e.ParentID,
			&e.Relationship,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (*StudentParent) TableName() string {
	return "student_parents"
}
