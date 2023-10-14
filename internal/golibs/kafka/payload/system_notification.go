package payload

import (
	"time"
)

type SystemNotificationStatus string

const (
	SystemNotificationStatusNew  SystemNotificationStatus = "SYSTEM_NOTIFICATION_STATUS_NEW"
	SystemNotificationStatusDone SystemNotificationStatus = "SYSTEM_NOTIFICATION_STATUS_DONE"
)

type UpsertSystemNotification struct {
	ReferenceID string                        `json:"reference_id"`
	Content     []SystemNotificationContent   `json:"content"`
	URL         string                        `json:"url"`
	ValidFrom   time.Time                     `json:"valid_from"`
	Recipients  []SystemNotificationRecipient `json:"recipients"`
	IsDeleted   bool                          `json:"is_deleted"`
	Status      SystemNotificationStatus      `json:"status"`
}

type SystemNotificationRecipient struct {
	UserID string `json:"user_id"`
}

type SystemNotificationContent struct {
	Language string `json:"language"`
	Text     string `json:"text"`
}
