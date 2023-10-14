package entity

import (
	"time"

	"github.com/manabie-com/backend/internal/golibs/learnosity"
)

// SessionResponse represents the response returned from the Learnosity Data API.
type SessionResponse struct {
	UserID       string                   `json:"user_id"`
	ActivityID   string                   `json:"activity_id"`
	NumAttempted int                      `json:"num_attempted"`
	NumQuestions int                      `json:"num_questions"`
	SessionID    string                   `json:"session_id"`
	Score        int                      `json:"score"`
	MaxScore     int                      `json:"max_score"`
	Status       learnosity.SessionStatus `json:"status"`
	DtSaved      time.Time                `json:"dt_saved"`
	DtStarted    time.Time                `json:"dt_started"`
	DtCompleted  *time.Time               `json:"dt_completed"` // if status is Incomplete, then this field is null
}
