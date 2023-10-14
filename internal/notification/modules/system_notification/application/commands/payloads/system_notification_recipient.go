package payloads

import "github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"

type UpsertSystemNotificationRecipientPayload struct {
	SystemNotificationID string
	Recipients           []*dto.SystemNotificationRecipient
}
