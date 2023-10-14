package domain

import (
	"time"
)

type StudentEventLog struct {
	EventID            string
	EventType          string
	StudentID          string
	LearningMaterialID string
	Payload            map[string]any

	CreatedAt time.Time
}
