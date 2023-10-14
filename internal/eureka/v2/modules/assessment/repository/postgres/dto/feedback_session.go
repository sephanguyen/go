package dto

import (
	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"

	"github.com/jackc/pgtype"
)

type FeedbackSession struct {
	BaseEntity
	ID           pgtype.Text
	SubmissionID pgtype.Text
	CreatedBy    pgtype.Text
}

func (a *FeedbackSession) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"submission_id",
			"created_by",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&a.ID,
			&a.SubmissionID,
			&a.CreatedBy,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		}
}

func (a *FeedbackSession) TableName() string {
	return "feedback_session"
}

func (a *FeedbackSession) ToEntity() domain.FeedbackSession {
	sub := domain.FeedbackSession{
		ID:           a.ID.String,
		SubmissionID: a.SubmissionID.String,
		CreatedBy:    a.CreatedBy.String,
		CreatedAt:    a.CreatedAt.Time,
	}
	return sub
}
