package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type AssessmentSession struct {
	BaseEntity
	SessionID    pgtype.Text
	UserID       pgtype.Text
	AssessmentID pgtype.Text
}

func (e *AssessmentSession) FieldMap() ([]string, []interface{}) {
	return []string{
			"session_id",
			"user_id",
			"assessment_id",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&e.SessionID,
			&e.UserID,
			&e.AssessmentID,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
		}
}

func (e *AssessmentSession) TableName() string {
	return "assessment_session"
}

type AssessmentSessions []*AssessmentSession

func (u *AssessmentSessions) Add() database.Entity {
	e := &AssessmentSession{}
	*u = append(*u, e)
	return e
}
