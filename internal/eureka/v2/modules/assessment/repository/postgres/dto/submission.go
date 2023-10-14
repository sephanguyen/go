package dto

import (
	"time"

	"github.com/manabie-com/backend/internal/eureka/v2/modules/assessment/domain"
	"github.com/manabie-com/backend/internal/eureka/v2/pkg/errors"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type Submission struct {
	BaseEntity
	ID           pgtype.Text
	SessionID    pgtype.Text
	AssessmentID pgtype.Text
	StudentID    pgtype.Text
	CompletedAt  pgtype.Timestamptz

	AllocatedMarkerID pgtype.Text
	MarkedBy          pgtype.Text
	MarkedAt          pgtype.Timestamptz
	GradingStatus     pgtype.Text

	MaxScore    pgtype.Int4
	GradedScore pgtype.Int4
}

func (a *Submission) FieldMap() ([]string, []interface{}) {
	return []string{
			"id",
			"session_id",
			"assessment_id",
			"student_id",
			"grading_status",
			"allocated_marker_id",
			"marked_by",
			"marked_at",
			"max_score",
			"graded_score",
			"completed_at",
			"created_at",
			"updated_at",
			"deleted_at",
		}, []interface{}{
			&a.ID,
			&a.SessionID,
			&a.AssessmentID,
			&a.StudentID,
			&a.GradingStatus,
			&a.AllocatedMarkerID,
			&a.MarkedBy,
			&a.MarkedAt,
			&a.MaxScore,
			&a.GradedScore,
			&a.CompletedAt,
			&a.CreatedAt,
			&a.UpdatedAt,
			&a.DeletedAt,
		}
}

func (a *Submission) TableName() string {
	return "assessment_submission"
}

func (a *Submission) ToEntity() domain.Submission {
	sub := domain.Submission{
		ID:                a.ID.String,
		SessionID:         a.SessionID.String,
		AssessmentID:      a.AssessmentID.String,
		StudentID:         a.StudentID.String,
		AllocatedMarkerID: a.AllocatedMarkerID.String,
		GradingStatus:     toSubmissionGradingStatus(a.GradingStatus.String),
		MarkedBy:          a.MarkedBy.String,
		MaxScore:          int(a.MaxScore.Int),
		GradedScore:       int(a.GradedScore.Int),
		CreatedAt:         a.CreatedAt.Time,
		CompletedAt:       a.CompletedAt.Time,
	}
	if a.MarkedAt.Status == pgtype.Present {
		sub.MarkedAt = &a.MarkedAt.Time
	}
	return sub
}

func toSubmissionGradingStatus(s string) domain.GradingStatus {
	var ds domain.GradingStatus
	switch s {
	case "NOT_MARKED":
		ds = domain.GradingStatusNotMarked
	case "IN_PROGRESS":
		ds = domain.GradingStatusInProgress
	case "MARKED":
		ds = domain.GradingStatusMarked
	case "RETURNED":
		ds = domain.GradingStatusReturned
	default:
		ds = domain.GradingStatusNotMarked
	}
	return ds
}

func (a *Submission) FromEntity(now time.Time, d domain.Submission) error {
	database.AllNullEntity(a)

	errs := multierr.Combine(
		a.ID.Set(d.ID),
		a.SessionID.Set(d.SessionID),
		a.AssessmentID.Set(d.AssessmentID),
		a.StudentID.Set(d.StudentID),
		a.CompletedAt.Set(d.CompletedAt),
		a.GradingStatus.Set(string(d.GradingStatus)),
		a.MaxScore.Set(d.MaxScore),
		a.GradedScore.Set(d.GradedScore),
		a.CreatedAt.Set(now),
		a.UpdatedAt.Set(now),
	)

	if d.AllocatedMarkerID != "" {
		errs = multierr.Append(errs, a.AllocatedMarkerID.Set(d.AllocatedMarkerID))
	}

	if d.MarkedBy != "" {
		errs = multierr.Append(errs, a.MarkedBy.Set(d.MarkedBy))
	}

	if d.MarkedAt != nil {
		errs = multierr.Append(errs, a.MarkedAt.Set(d.MarkedAt))
	}

	if errs != nil {
		return errors.NewConversionError("multierr.Combine", errs)
	}

	return nil
}
