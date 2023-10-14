package mappers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/subscribers/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func NatsNotificationToPb(data *ypb.NatsCreateNotificationRequest) (*cpb.Notification, error) {
	// Set status default for draft type
	scheduledAt := timestamppb.Now()
	status := cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT
	// Override status with scheduled type
	if data.SendTime.Type == consts.NotificationTypeScheduled {
		status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
		var err error
		scheduledAtTemp, err := time.Parse("2006-01-02T15:04:05Z07:00", data.SendTime.Time)
		scheduledAt = timestamppb.New(scheduledAtTemp)
		if err != nil {
			return nil, fmt.Errorf("invalid send time")
		}
	}

	// Inject send mode to data
	data.NotificationConfig.Data["mode"] = data.NotificationConfig.Mode
	data.NotificationConfig.Data["tracing_id"] = data.TracingId

	// Clear all properties before set new value
	jsonData, err := json.Marshal(data.NotificationConfig.Data)

	if err != nil {
		return nil, err
	}

	notiMessageStr, err := utils.GetRawNotificationNatsMessage(data.NotificationConfig.Notification)

	if err != nil {
		return nil, err
	}

	// Bind all values to notification
	notiCpb := &cpb.Notification{
		Type:        cpb.NotificationType_NOTIFICATION_TYPE_NATS_ASYNC,
		Event:       cpb.NotificationEvent_NOTIFICATION_EVENT_NONE,
		Status:      status,
		ScheduledAt: scheduledAt,
		ReceiverIds: data.Target.ReceivedUserIds,
		SchoolId:    data.SchoolId,
		Data:        string(jsonData),
		Message: &cpb.NotificationMessage{
			Title: data.NotificationConfig.Notification.Title,
			Content: &cpb.RichText{
				Raw:      notiMessageStr,
				Rendered: data.NotificationConfig.Notification.Content,
			},
		},
	}

	// Bind target group
	notiCpb.TargetGroup = &cpb.NotificationTargetGroup{
		CourseFilter: &cpb.NotificationTargetGroup_CourseFilter{
			Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			CourseIds: []string{},
		},
		GradeFilter: &cpb.NotificationTargetGroup_GradeFilter{
			Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			GradeIds: []string{},
		},
		UserGroupFilter: &cpb.NotificationTargetGroup_UserGroupFilter{
			UserGroups: []cpb.UserGroup{},
		},
		LocationFilter: &cpb.NotificationTargetGroup_LocationFilter{
			Type:        cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			LocationIds: []string{},
		},
		ClassFilter: &cpb.NotificationTargetGroup_ClassFilter{
			Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			ClassIds: []string{},
		},
		SchoolFilter: &cpb.NotificationTargetGroup_SchoolFilter{
			Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
			SchoolIds: []string{},
		},
	}

	if data.TargetGroup != nil && data.TargetGroup.UserGroupFilter != nil && len(data.TargetGroup.UserGroupFilter.UserGroups) > 0 {
		for _, userGroup := range data.TargetGroup.UserGroupFilter.UserGroups {
			switch userGroup {
			case consts.TargetUserGroupParent:
				notiCpb.TargetGroup.UserGroupFilter.UserGroups = append(notiCpb.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_PARENT)
			case consts.TargetUserGroupStudent:
				notiCpb.TargetGroup.UserGroupFilter.UserGroups = append(notiCpb.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_STUDENT)
			}
		}
	}
	if data.GetTarget() != nil {
		notiCpb.GenericReceiverIds = data.GetTarget().GenericUserIds
	}

	return notiCpb, nil
}
