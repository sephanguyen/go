package domain

import (
	"time"
)

type FeedbackSession struct {
	// MUST BE UUID v4 instead of ULID
	ID           string
	SubmissionID string
	CreatedBy    string
	CreatedAt    time.Time
}
