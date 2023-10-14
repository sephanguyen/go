package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type QuestionnaireUserAnswer struct {
	AnswerID                pgtype.Text
	UserNotificationID      pgtype.Text
	QuestionnaireQuestionID pgtype.Text
	UserID                  pgtype.Text
	TargetID                pgtype.Text
	Answer                  pgtype.Text
	SubmittedAt             pgtype.Timestamptz
	DeletedAt               pgtype.Timestamptz
}

func (e *QuestionnaireUserAnswer) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"answer_id",
		"user_notification_id",
		"questionnaire_question_id",
		"user_id",
		"target_id",
		"answer",
		"submitted_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.AnswerID,
		&e.UserNotificationID,
		&e.QuestionnaireQuestionID,
		&e.UserID,
		&e.TargetID,
		&e.Answer,
		&e.SubmittedAt,
		&e.DeletedAt,
	}
	return
}

func (*QuestionnaireUserAnswer) TableName() string {
	return "questionnaire_user_answers"
}

type QuestionnaireUserAnswers []*QuestionnaireUserAnswer

func (u *QuestionnaireUserAnswers) Add() database.Entity {
	e := &QuestionnaireUserAnswer{}
	*u = append(*u, e)

	return e
}
