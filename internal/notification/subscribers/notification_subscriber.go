package subscribers

import (
	nservices "github.com/manabie-com/backend/internal/notification/services"
)

type NotificationSubscriber struct {
	notificationService *nservices.NotificationModifierService
}

func NewNotificationSubscriber(svc *nservices.NotificationModifierService) *NotificationSubscriber {
	return &NotificationSubscriber{
		notificationService: svc,
	}
}
