package utils

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/consts"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func SliceUsersGroupContains(elems []cpb.UserGroup, v cpb.UserGroup) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func GenSampleNatsNotification() *ypb.NatsCreateNotificationRequest {
	tracingID := idutil.ULIDNow()
	notificationTemplate := &ypb.NatsCreateNotificationRequest{
		ClientId:       "unit_test_client",
		SendingMethods: []string{consts.SendingMethodPushNotification},
		Target:         &ypb.NatsNotificationTarget{},
		TargetGroup: &ypb.NatsNotificationTargetGroup{
			UserGroupFilter: &ypb.NatsNotificationTargetGroup_UserGroupFilter{
				UserGroups: []string{consts.TargetUserGroupParent, consts.TargetUserGroupStudent},
			},
		},
		NotificationConfig: &ypb.NatsPushNotificationConfig{
			Mode:             consts.NotificationModeNotify,
			PermanentStorage: true,
			Notification: &ypb.NatsNotification{
				Title:   fmt.Sprintf("nats notify %v", tracingID),
				Message: "popup message",
				Content: "<h1>hello world</h1>",
			},
			Data: map[string]string{
				"custom_data_type": "unit test",
			},
		},
		SendTime:  &ypb.NatsNotificationSendTime{},
		TracingId: tracingID,
		SchoolId:  1,
	}

	return notificationTemplate
}
