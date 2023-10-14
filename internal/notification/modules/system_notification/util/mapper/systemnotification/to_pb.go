package systemnotification

import (
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/domain/dto"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToSystemNotificationPb(systemNotifications []*dto.SystemNotification) []*npb.RetrieveSystemNotificationsResponse_SystemNotification {
	systemNotificationsPb := make([]*npb.RetrieveSystemNotificationsResponse_SystemNotification, 0, len(systemNotifications))
	for _, sn := range systemNotifications {
		systemNotification := &npb.RetrieveSystemNotificationsResponse_SystemNotification{
			SystemNotificationId: sn.SystemNotificationID,
			Content:              ToSystemNotificationContentPb(sn.Content),
			Url:                  sn.URL,
			ValidFrom:            timestamppb.New(sn.ValidFrom),
			Status:               npb.SystemNotificationStatus(npb.SystemNotificationStatus_value[sn.Status]),
		}
		systemNotificationsPb = append(systemNotificationsPb, systemNotification)
	}
	return systemNotificationsPb
}

func ToSystemNotificationContentPb(systemNotificationContent []*dto.SystemNotificationContent) []*npb.RetrieveSystemNotificationsResponse_SystemNotificationContent {
	ret := make([]*npb.RetrieveSystemNotificationsResponse_SystemNotificationContent, 0, len(systemNotificationContent))
	for _, snc := range systemNotificationContent {
		ret = append(ret, &npb.RetrieveSystemNotificationsResponse_SystemNotificationContent{
			Language: snc.Language,
			Text:     snc.Text,
		})
	}
	return ret
}
