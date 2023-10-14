package repo

import (
	"github.com/jackc/pgtype"
)

type CourseSubject struct {
	CourseID  pgtype.Varchar
	SubjectID pgtype.Varchar

	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *CourseSubject) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"course_id",
		"subject_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&c.CourseID,
		&c.SubjectID,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.DeletedAt,
	}
	return
}

func (*CourseSubject) TableName() string {
	return "course_subject"
}
