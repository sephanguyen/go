package domain

import (
	"sort"
	"time"
)

type Submission struct {
	ID                string
	SessionID         string
	AssessmentID      string
	StudentID         string
	AllocatedMarkerID string

	GradingStatus GradingStatus

	MaxScore    int
	GradedScore int
	MarkedBy    string
	MarkedAt    *time.Time

	FeedBackSessionID string
	FeedBackBy        string

	CreatedAt   time.Time
	CompletedAt time.Time
}

type GradingStatus string

const (
	// GradingStatusNone is not stored in DB, but just a state of an uncompleted session
	GradingStatusNone       GradingStatus = "NONE"
	GradingStatusNotMarked  GradingStatus = "NOT_MARKED"
	GradingStatusInProgress GradingStatus = "IN_PROGRESS"
	GradingStatusMarked     GradingStatus = "MARKED"
	GradingStatusReturned   GradingStatus = "RETURNED"
)

type Submissions []Submission

func (subs Submissions) SortDescByCompletedAt() {
	sort.Slice(subs, func(i, j int) bool {
		return subs[i].CompletedAt.After(subs[j].CompletedAt)
	})
}

func (sub *Submission) Validate() error {
	if sub.ID == "" {
		return ErrSubmissionIDRequired
	}
	if sub.AssessmentID == "" {
		return ErrAssessmentIDRequired
	}
	if sub.StudentID == "" {
		return ErrStudentIDRequired
	}
	if sub.CompletedAt.IsZero() {
		return ErrCompletedAtRequired
	}

	switch sub.GradingStatus {
	case GradingStatusNotMarked, GradingStatusInProgress, GradingStatusMarked, GradingStatusReturned:
	default:
		return ErrInvalidGradingStatus
	}

	return nil
}
