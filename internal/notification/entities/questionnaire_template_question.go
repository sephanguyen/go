package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type QuestionnaireTemplateQuestion struct {
	QuestionnaireTemplateQuestionID pgtype.Text
	QuestionnaireTemplateID         pgtype.Text
	OrderIndex                      pgtype.Int4
	Type                            pgtype.Text
	Title                           pgtype.Text
	Choices                         pgtype.TextArray
	IsRequired                      pgtype.Bool
	CreatedAt                       pgtype.Timestamptz
	UpdatedAt                       pgtype.Timestamptz
	DeletedAt                       pgtype.Timestamptz
}

func (e *QuestionnaireTemplateQuestion) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"questionnaire_template_question_id",
		"questionnaire_template_id",
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
		&e.QuestionnaireTemplateQuestionID,
		&e.QuestionnaireTemplateID,
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

func (*QuestionnaireTemplateQuestion) TableName() string {
	return "questionnaire_template_questions"
}

type QuestionnaireTemplateQuestions []*QuestionnaireTemplateQuestion

func (u *QuestionnaireTemplateQuestions) Add() database.Entity {
	e := &QuestionnaireTemplateQuestion{}
	*u = append(*u, e)

	return e
}
