package mappers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/subscribers/utils"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	LAYOUT_TIME_FORMAT = "2006-01-02T15:04:05Z07:00"
)

func Test_NatsNotificationToPb(t *testing.T) {
	checkNotificationResponseData := func(t *testing.T, req *ypb.NatsCreateNotificationRequest, res *cpb.Notification) {
		// sendtime
		if req.SendTime.Type == consts.NotificationTypeScheduled {
			scheduledAtTemp, err := time.Parse(LAYOUT_TIME_FORMAT, req.SendTime.Time)
			assert.Nil(t, err)
			scheduledAt := timestamppb.New(scheduledAtTemp)
			assert.Equal(t, cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED, res.Status)
			assert.Equal(t, scheduledAt, res.ScheduledAt)
		} else {
			assert.Equal(t, cpb.NotificationStatus_NOTIFICATION_STATUS_DRAFT, res.Status)
		}

		// data payload
		req.NotificationConfig.Data["mode"] = req.NotificationConfig.Mode
		req.NotificationConfig.Data["tracing_id"] = req.TracingId
		jsonData, err := json.Marshal(req.NotificationConfig.Data)
		assert.Nil(t, err)
		assert.Equal(t, string(jsonData), res.Data)

		// message
		assert.Equal(t, req.NotificationConfig.Notification.Title, res.Message.Title)
		assert.Equal(t, req.NotificationConfig.Notification.Content, res.Message.Content.Rendered)
		notiMessageStr, err := utils.GetRawNotificationNatsMessage(req.NotificationConfig.Notification)
		assert.Nil(t, err)
		assert.Equal(t, notiMessageStr, res.Message.Content.GetRaw())

		// target group
		assert.Equal(t, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, res.TargetGroup.CourseFilter.Type)
		assert.Equal(t, []string{}, res.TargetGroup.CourseFilter.CourseIds)
		assert.Equal(t, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, res.TargetGroup.GradeFilter.Type)
		assert.Equal(t, []string{}, res.TargetGroup.GradeFilter.GradeIds)
		assert.Equal(t, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, res.TargetGroup.ClassFilter.Type)
		assert.Equal(t, []string{}, res.TargetGroup.ClassFilter.ClassIds)
		assert.Equal(t, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, res.TargetGroup.LocationFilter.Type)
		assert.Equal(t, []string{}, res.TargetGroup.LocationFilter.LocationIds)
		assert.Equal(t, cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE, res.TargetGroup.SchoolFilter.Type)
		assert.Equal(t, []string{}, res.TargetGroup.SchoolFilter.SchoolIds)

		assert.Equal(t, len(req.TargetGroup.UserGroupFilter.UserGroups), len(res.TargetGroup.UserGroupFilter.UserGroups))

		if req.TargetGroup != nil && req.TargetGroup.UserGroupFilter != nil && len(req.TargetGroup.UserGroupFilter.UserGroups) > 0 {
			for _, user_group := range req.TargetGroup.UserGroupFilter.UserGroups {
				switch user_group {
				case consts.TargetUserGroupParent:
					utils.SliceUsersGroupContains(res.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_PARENT)
				case consts.TargetUserGroupStudent:
					utils.SliceUsersGroupContains(res.TargetGroup.UserGroupFilter.UserGroups, cpb.UserGroup_USER_GROUP_STUDENT)
				}
			}
		}
	}

	t.Run("happy case immediate", func(t *testing.T) {
		t.Parallel()
		reqNoti := utils.GenSampleNatsNotification()
		reqNoti.SendTime.Type = consts.NotificationTypeImmediate
		res, err := NatsNotificationToPb(reqNoti)
		assert.Nil(t, err)
		checkNotificationResponseData(t, reqNoti, res)
	})

	t.Run("happy case schedule", func(t *testing.T) {
		t.Parallel()
		reqNoti := utils.GenSampleNatsNotification()
		reqNoti.SendTime.Type = consts.NotificationTypeScheduled
		reqNoti.SendTime.Time = time.Now().Add(1 * time.Minute).Format(LAYOUT_TIME_FORMAT)
		res, err := NatsNotificationToPb(reqNoti)
		assert.Nil(t, err)
		checkNotificationResponseData(t, reqNoti, res)
	})
}
