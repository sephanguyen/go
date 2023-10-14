package queries

import (
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
)

type RetrieveSystemNotificationPayload struct {
	UserID   string
	Limit    uint32
	Offset   int64
	Language string
	Status   string
	Keyword  string
}

type RetrieveSystemNotificationResponse struct {
	TotalCount          uint32
	TotalForStatus      map[string]uint32
	SystemNotifications []*dto.SystemNotification
	Error               error
}
