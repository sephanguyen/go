package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Questionnaire struct {
	QuestionnaireID         pgtype.Text
	QuestionnaireTemplateID pgtype.Text
	ResubmitAllowed         pgtype.Bool
	ExpirationDate          pgtype.Timestamptz
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (e *Questionnaire) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"questionnaire_id",
		"questionnaire_template_id",
		"resubmit_allowed",
		"expiration_date",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.QuestionnaireID,
		&e.QuestionnaireTemplateID,
		&e.ResubmitAllowed,
		&e.ExpirationDate,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*Questionnaire) TableName() string {
	return "questionnaires"
}

type Questionnaires []*Questionnaire

func (u *Questionnaires) Add() database.Entity {
	e := &Questionnaire{}
	*u = append(*u, e)

	return e
}
