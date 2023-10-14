package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Book struct {
	ID                         pgtype.Text `sql:"book_id,pk"`
	Name                       pgtype.Text
	Country                    pgtype.Text
	Subject                    pgtype.Text
	Grade                      pgtype.Int2
	SchoolID                   pgtype.Int4 `sql:"school_id"`
	UpdatedAt                  pgtype.Timestamptz
	CreatedAt                  pgtype.Timestamptz
	DeletedAt                  pgtype.Timestamptz
	CopiedFrom                 pgtype.Text
	CurrentChapterDisplayOrder pgtype.Int4
	BookType                   pgtype.Text
	IsV2                       pgtype.Bool
}

func (c *Book) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"book_id", "name", "country", "subject", "grade", "school_id", "updated_at", "created_at", "deleted_at", "copied_from", "current_chapter_display_order", "book_type", "is_v2"}
	values = []interface{}{&c.ID, &c.Name, &c.Country, &c.Subject, &c.Grade, &c.SchoolID, &c.UpdatedAt, &c.CreatedAt, &c.DeletedAt, &c.CopiedFrom, &c.CurrentChapterDisplayOrder, &c.BookType, &c.IsV2}
	return
}

func (*Book) TableName() string {
	return "books"
}

type Books []*Book

func (u *Books) Add() database.Entity {
	e := &Book{}
	*u = append(*u, e)

	return e
}
