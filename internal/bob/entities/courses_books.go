package entities

import (
	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type CoursesBooks struct {
	BookID    pgtype.Text `sql:"book_id"`
	CourseID  pgtype.Text `sql:"course_id"`
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *CoursesBooks) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id", "course_id", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.BookID, &c.CourseID, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*CoursesBooks) TableName() string {
	return "courses_books"
}

type CoursesBookss []*CoursesBooks

func (u *CoursesBookss) Add() database.Entity {
	e := &CoursesBooks{}
	*u = append(*u, e)

	return e
}
