package dto

import "time"

type SystemNotification struct {
	SystemNotificationID string
	ReferenceID          string
	Content              []*SystemNotificationContent
	URL                  string
	ValidFrom            time.Time
	Recipients           []*SystemNotificationRecipient
	Status               string
	IsDeleted            bool
}
