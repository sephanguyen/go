package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type LiveLessonSentNotification struct {
	SentNotificationID pgtype.Text
	LessonID           pgtype.Text
	SentAt             pgtype.Timestamptz
	SentAtInterval     pgtype.Text
	CreatedAt          pgtype.Timestamptz
	UpdatedAt          pgtype.Timestamptz
}

func NewLiveLessonSentNotificationEntity(l *domain.LiveLessonSentNotification) (*LiveLessonSentNotification, error) {
	dto := &LiveLessonSentNotification{}
	if err := multierr.Combine(
		dto.SentNotificationID.Set(l.SentNotificationID),
		dto.LessonID.Set(l.LessonID),
		dto.SentAt.Set(l.SentAt),
		dto.SentAtInterval.Set(l.SentAtInterval),
		dto.CreatedAt.Set(l.CreatedAt),
		dto.UpdatedAt.Set(l.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from live lesson sent notification domain to entity: %w", err)
	}

	return dto, nil
}

func (l *LiveLessonSentNotification) FieldMap() ([]string, []interface{}) {
	return []string{
			"sent_notification_id",
			"lesson_id",
			"sent_at",
			"sent_at_interval",
			"created_at",
			"updated_at",
		}, []interface{}{
			&l.SentNotificationID,
			&l.LessonID,
			&l.SentAt,
			&l.SentAtInterval,
			&l.CreatedAt,
			&l.UpdatedAt,
		}
}

func (l *LiveLessonSentNotification) TableName() string {
	return "live_lesson_sent_notifications"
}

func (l *LiveLessonSentNotification) ToLiveLessonSentNotificationDomain() *domain.LiveLessonSentNotification {
	return &domain.LiveLessonSentNotification{
		SentNotificationID: l.SentNotificationID.String,
		LessonID:           l.LessonID.String,
		SentAt:             l.SentAt.Time,
		SentAtInterval:     l.SentAtInterval.String,
		CreatedAt:          l.CreatedAt.Time,
		UpdatedAt:          l.UpdatedAt.Time,
	}
}
