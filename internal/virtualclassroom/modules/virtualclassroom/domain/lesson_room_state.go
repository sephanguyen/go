package domain

import (
	"errors"
	"time"
)

var (
	ErrLessonRoomStateNotFound = errors.New("lesson room state not found")
)

type LessonRoomState struct {
	ID                  string
	LessonID            string
	CurrentMaterial     *CurrentMaterial
	SpotlightedUser     string
	WhiteboardZoomState *WhiteboardZoomState
	Recording           *CompositeRecordingState
	CurrentPolling      *CurrentPolling
	SessionTime         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           *time.Time
}
