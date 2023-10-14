package repo

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
)

type LessonTeacher struct {
	LessonID    pgtype.Text
	TeacherID   pgtype.Text
	TeacherName pgtype.Text
	CreatedAt   pgtype.Timestamptz
	DeletedAt   pgtype.Timestamptz
}

func (l *LessonTeacher) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"lesson_id",
		"teacher_id",
		"teacher_name",
		"created_at",
		"deleted_at",
	}
	values = []interface{}{
		&l.LessonID,
		&l.TeacherID,
		&l.TeacherName,
		&l.CreatedAt,
		&l.DeletedAt,
	}
	return
}

func (*LessonTeacher) TableName() string {
	return "lessons_teachers"
}

type LessonTeachers []*LessonTeacher

func (u *LessonTeachers) Add() database.Entity {
	e := &LessonTeacher{}
	*u = append(*u, e)

	return e
}

func (l *LessonTeacher) ToLessonTeacherEntity() *domain.LessonTeacher {
	lt := &domain.LessonTeacher{
		TeacherID: l.TeacherID.String,
		Name:      l.TeacherName.String,
	}

	return lt
}
