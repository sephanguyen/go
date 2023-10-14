package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type BookChapter struct {
	BookID    pgtype.Text `sql:"book_id"`
	ChapterID pgtype.Text `sql:"chapter_id"`
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

func (c *BookChapter) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id", "chapter_id", "updated_at", "created_at", "deleted_at"}
	values = []interface{}{&c.BookID, &c.ChapterID, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt}
	return
}

func (*BookChapter) TableName() string {
	return "books_chapters"
}

type BookChapters []*BookChapter

func (u *BookChapters) Add() database.Entity {
	e := &BookChapter{}
	*u = append(*u, e)

	return e
}
