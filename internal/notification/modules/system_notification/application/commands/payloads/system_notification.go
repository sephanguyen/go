package payloads

import "github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"

type UpsertSystemNotificationPayload struct {
	SystemNotification *dto.SystemNotification
}

type SoftDeleteSystemNotificationPayload struct {
	ReferenceID string
}

type SetSystemNotificationStatusPayload struct {
	SystemNotificationID string
	Status               string
}
