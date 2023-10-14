package domain

import (
	"sort"
	"time"
)

// Session represents a unique attempt of an assessment by a student at a specific time.
type Session struct {
	ID string

	// The compound IDs is Identity.
	AssessmentID string // known as activity_id.
	UserID       string

	// Virtual props.
	CourseID           string
	LearningMaterialID string
	Submission         *Submission

	// Other information.
	MaxScore    int
	GradedScore int
	Status      SessionStatus

	CreatedAt   time.Time
	CompletedAt *time.Time
}

type SessionStatus string

const (
	SessionStatusNone       SessionStatus = "NONE"
	SessionStatusIncomplete SessionStatus = "INCOMPLETE"
	SessionStatusCompleted  SessionStatus = "COMPLETED"
)

func (s *Session) Validate() (err error) {
	if s.ID == "" {
		return ErrIDRequired
	}
	if s.AssessmentID == "" {
		return ErrAssessmentIDRequired
	}
	if s.UserID == "" {
		return ErrUserIDRequired
	}

	switch s.Status {
	case SessionStatusNone, SessionStatusIncomplete, SessionStatusCompleted:
	default:
		return ErrInvalidSessionStatus
	}
	return nil
}

type Sessions []Session

func (ss Sessions) SortDescByCompletedAt() {
	sort.Slice(ss, func(i, j int) bool {
		if ss[i].CompletedAt == nil || ss[j].CompletedAt == nil {
			return ss[i].CreatedAt.After(ss[j].CreatedAt)
		}
		return ss[i].CompletedAt.After(*ss[j].CompletedAt)
	})
}
