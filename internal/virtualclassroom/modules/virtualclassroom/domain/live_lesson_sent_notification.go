package domain

import "time"

type LiveLessonSentNotification struct {
	SentNotificationID string
	LessonID           string
	SentAt             time.Time
	SentAtInterval     string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
