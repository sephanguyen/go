package payloads

import "github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"

type UpsertSystemNotificationContentPayload struct {
	SystemNotificationID       string
	SystemNotificationContents []*dto.SystemNotificationContent
}
