package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type AssessmentSession struct {
	BaseEntity
	SessionID    pgtype.Text
	AssessmentID pgtype.Text
	UserID       pgtype.Text

	Status pgtype.Text
}

func (a *AssessmentSession) FieldMap() ([]string, []interface{}) {
	return []string{
			"session_id",
			"assessment_id",
			"user_id",
			"status",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&a.SessionID,
			&a.AssessmentID,
			&a.UserID,
			&a.Status,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		}
}

func (a *AssessmentSession) TableName() string {
	return "assessment_session"
}

func (a *AssessmentSession) ToEntity() (session domain.Session, err error) {
	session = domain.Session{
		ID:           a.SessionID.String,
		AssessmentID: a.AssessmentID.String,
		UserID:       a.UserID.String,
		Status:       domain.SessionStatus(a.Status.String),
		CreatedAt:    a.CreatedAt.Time,
	}
	err = session.Validate()
	return session, err
}

func (a *AssessmentSession) FromEntity(now time.Time, d domain.Session) error {
	database.AllNullEntity(a)

	errs := multierr.Combine(
		a.SessionID.Set(d.ID),
		a.AssessmentID.Set(d.AssessmentID),
		a.UserID.Set(d.UserID),
		a.CreatedAt.Set(now),
		a.UpdatedAt.Set(now),
		a.Status.Set(d.Status),
	)

	if errs != nil {
		return errors.NewConversionError("multierr.Combine", errs)
	}

	return nil
}
