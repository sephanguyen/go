package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type QuestionTagType struct {
	BaseEntity
	QuestionTagTypeID pgtype.Text `sql:"question_tag_type_id"`
	Name              pgtype.Text `sql:"name"`
}

// FieldMap return a map of field name and pointer to field
func (e *QuestionTagType) FieldMap() ([]string, []interface{}) {
	names := []string{
		"question_tag_type_id",
		"name",
		"updated_at",
		"created_at",
		"deleted_at",
	}
	return names, []interface{}{
		&e.QuestionTagTypeID,
		&e.Name,
		&e.UpdatedAt,
		&e.CreatedAt,
		&e.DeletedAt,
	}
}

// TableName returns "question_tag_type" table name
func (e *QuestionTagType) TableName() string {
	return "question_tag_type"
}

type QuestionTagTypes []*QuestionTagType

func (ts *QuestionTagTypes) Add() database.Entity {
	t := &QuestionTagType{}
	*ts = append(*ts, t)

	return t
}

func (ts QuestionTagTypes) Get() []*QuestionTagType {
	return []*QuestionTagType(ts)
}
