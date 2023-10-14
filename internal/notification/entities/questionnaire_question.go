package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type QuestionnaireQuestion struct {
	QuestionnaireQuestionID pgtype.Text
	QuestionnaireID         pgtype.Text
	OrderIndex              pgtype.Int4
	Type                    pgtype.Text
	Title                   pgtype.Text
	Choices                 pgtype.TextArray
	IsRequired              pgtype.Bool
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (e *QuestionnaireQuestion) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"questionnaire_question_id",
		"questionnaire_id",
		"order_index",
		"type",
		"title",
		"choices",
		"is_required",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.QuestionnaireQuestionID,
		&e.QuestionnaireID,
		&e.OrderIndex,
		&e.Type,
		&e.Title,
		&e.Choices,
		&e.IsRequired,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*QuestionnaireQuestion) TableName() string {
	return "questionnaire_questions"
}

type QuestionnaireQuestions []*QuestionnaireQuestion

func (u *QuestionnaireQuestions) Add() database.Entity {
	e := &QuestionnaireQuestion{}
	*u = append(*u, e)

	return e
}
