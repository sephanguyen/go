package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Tag struct {
	TagID      pgtype.Text
	TagName    pgtype.Text
	CreatedAt  pgtype.Timestamptz
	UpdatedAt  pgtype.Timestamptz
	DeletedAt  pgtype.Timestamptz
	IsArchived pgtype.Bool
}

func (e *Tag) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"tag_id",
		"tag_name",
		"created_at",
		"updated_at",
		"deleted_at",
		"is_archived",
	}
	values = []interface{}{
		&e.TagID,
		&e.TagName,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
		&e.IsArchived,
	}
	return
}

func (*Tag) TableName() string {
	return "tags"
}

type Tags []*Tag

func (u *Tags) Add() database.Entity {
	e := &Tag{}
	*u = append(*u, e)

	return e
}
