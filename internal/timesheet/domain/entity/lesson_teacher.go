package entity

import (
	"github.com/jackc/pgtype"
)

type LessonTeacher struct {
	LessonID  pgtype.Text
	TeacherID pgtype.Text
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (l *LessonTeacher) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"lesson_id",
		"teacher_id",
		"created_at",
		"deleted_at",
	}
	values = []interface{}{
		&l.LessonID,
		&l.TeacherID,
		&l.CreatedAt,
		&l.DeletedAt,
	}
	return
}

func (*LessonTeacher) TableName() string {
	return "lessons_teachers"
}
