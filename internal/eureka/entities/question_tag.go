package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type QuestionTag struct {
	QuestionTagID     pgtype.Text
	Name              pgtype.Text
	QuestionTagTypeID pgtype.Text
	BaseEntity
}

func (q *QuestionTag) FieldMap() ([]string, []interface{}) {
	fields := []string{"question_tag_id", "name", "question_tag_type_id", "updated_at", "created_at", "deleted_at"}
	values := []interface{}{&q.QuestionTagID, &q.Name, &q.QuestionTagTypeID, &q.UpdatedAt, &q.CreatedAt, &q.DeletedAt}
	return fields, values
}

func (q *QuestionTag) TableName() string {
	return "question_tag"
}

type QuestionTags []*QuestionTag

func (q *QuestionTags) Add() database.Entity {
	e := &QuestionTag{}
	*q = append(*q, e)

	return e
}

func (q QuestionTags) Get() []*QuestionTag {
	return []*QuestionTag(q)
}
