package dto

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type CourseBookDto struct {
	BookID    pgtype.Text `sql:"book_id"`
	CourseID  pgtype.Text `sql:"course_id"`
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *CourseBookDto) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id", "course_id", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.BookID, &c.CourseID, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*CourseBookDto) TableName() string {
	return "courses_books"
}

type CoursesBooksDtos []*CourseBookDto

func (u *CoursesBooksDtos) Add() database.Entity {
	e := &CourseBookDto{}
	*u = append(*u, e)

	return e
}

func NewCourseBookDtoFromEntity(courseID string, bookID string) (*CourseBookDto, error) {
	dto := &CourseBookDto{}
	err := multierr.Combine(
		dto.CourseID.Set(courseID),
		dto.BookID.Set(bookID),
	)
	return dto, err
}
