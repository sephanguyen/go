package entities

import (
	"github.com/jackc/pgtype"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type Chapter struct {
	ID                       pgtype.Text `sql:"chapter_id,pk"`
	Name                     pgtype.Text
	Country                  pgtype.Text
	Subject                  pgtype.Text
	Grade                    pgtype.Int2
	DisplayOrder             pgtype.Int2
	SchoolID                 pgtype.Int4 `sql:"school_id"`
	UpdatedAt                pgtype.Timestamptz
	CreatedAt                pgtype.Timestamptz
	DeletedAt                pgtype.Timestamptz
	CopiedFrom               pgtype.Text
	CurrentTopicDisplayOrder pgtype.Int4
}

func (rcv *Chapter) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"chapter_id", "name", "country", "subject", "grade", "display_order", "school_id", "updated_at", "created_at", "deleted_at", "copied_from", "current_topic_display_order"}
	values = []interface{}{&rcv.ID, &rcv.Name, &rcv.Country, &rcv.Subject, &rcv.Grade, &rcv.DisplayOrder, &rcv.SchoolID, &rcv.UpdatedAt, &rcv.CreatedAt, &rcv.DeletedAt, &rcv.CopiedFrom, &rcv.CurrentTopicDisplayOrder}
	return
}

func (*Chapter) TableName() string {
	return "chapters"
}

type Chapters []*Chapter

func (u *Chapters) Add() database.Entity {
	e := &Chapter{}
	*u = append(*u, e)

	return e
}

type CopiedChapter struct {
	ID         pgtype.Text
	CopyFromID pgtype.Text
}

func (rcv *CopiedChapter) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"chapter_id", "copied_from"}
	values = []interface{}{&rcv.ID, &rcv.CopyFromID}
	return
}

func (*CopiedChapter) TableName() string {
	return "chapters"
}

type CopiedChapters []*CopiedChapter

func (u *CopiedChapters) Add() database.Entity {
	e := &CopiedChapter{}
	*u = append(*u, e)

	return e
}
