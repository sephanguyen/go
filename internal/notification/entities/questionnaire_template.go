package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type QuestionnaireTemplate struct {
	QuestionnaireTemplateID pgtype.Text
	Name                    pgtype.Text
	ResubmitAllowed         pgtype.Bool
	ExpirationDate          pgtype.Timestamptz
	Type                    pgtype.Text
	CreatedAt               pgtype.Timestamptz
	UpdatedAt               pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (e *QuestionnaireTemplate) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"questionnaire_template_id",
		"name",
		"resubmit_allowed",
		"expiration_date",
		"type",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.QuestionnaireTemplateID,
		&e.Name,
		&e.ResubmitAllowed,
		&e.ExpirationDate,
		&e.Type,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*QuestionnaireTemplate) TableName() string {
	return "questionnaire_templates"
}

type QuestionnaireTemplates []*QuestionnaireTemplate

func (u *QuestionnaireTemplates) Add() database.Entity {
	e := &QuestionnaireTemplate{}
	*u = append(*u, e)

	return e
}
