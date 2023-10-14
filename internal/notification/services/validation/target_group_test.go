package validation

import (
	"fmt"
	"math/rand"
	"testing"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

	"github.com/bxcodec/faker/v3"
	"github.com/stretchr/testify/assert"
)

func Test_ValidateTargetGroup(t *testing.T) {
	t.Parallel()
	faker.SetRandomMapAndSliceMinSize(1)
	type tcase struct {
		name  string
		setup func(noti *cpb.Notification)
		err   error
	}
	tcases := []tcase{
		{
			name: "nil target group",
			setup: func(noti *cpb.Notification) {
				noti.TargetGroup = nil
			},
			err: fmt.Errorf("request Notification.TargetGroup is null"),
		},
		{
			name: "nil course",
			setup: func(noti *cpb.Notification) {
				noti.TargetGroup.CourseFilter = nil
			},
			err: fmt.Errorf("request Notification.TargetGroup.CourseFilter is null"),
		},
		{
			name: "nil grade",
			setup: func(noti *cpb.Notification) {
				noti.TargetGroup.GradeFilter = nil
			},
			err: fmt.Errorf("request Notification.TargetGroup.GradeFilter is null"),
		},
		{
			name: "nil location",
			setup: func(noti *cpb.Notification) {
				noti.TargetGroup.LocationFilter = nil
			},
			err: fmt.Errorf("request Notification.TargetGroup.LocationFilter is null"),
		},
		{
			name: "nil class",
			setup: func(noti *cpb.Notification) {
				noti.TargetGroup.ClassFilter = nil
			},
			err: fmt.Errorf("request Notification.TargetGroup.ClassFilter is null"),
		},
		{
			name: "nil user group",
			setup: func(noti *cpb.Notification) {
				noti.TargetGroup.UserGroupFilter = nil
			},
			err: fmt.Errorf("request Notification.TargetGroup.UserGroupFilter is null"),
		},
		{
			name: "empty user group selection",
			setup: func(noti *cpb.Notification) {
				noti.TargetGroup.UserGroupFilter.UserGroups = nil
			},
			err: fmt.Errorf("request Notification.TargetGroup.UserGroupFilter.UserGroup is empty"),
		},
		{
			name: "nil receiver_ids and empty target group",
			setup: func(noti *cpb.Notification) {
				noti.Status = cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED
				noti.ReceiverIds = nil
				noti.GenericReceiverIds = nil
				noti.TargetGroup.CourseFilter = emptyCourseFilter()
				noti.TargetGroup.GradeFilter = emptyGradeFilter()
				noti.TargetGroup.LocationFilter = emptyLocationFilter()
				noti.TargetGroup.ClassFilter = emptyClassFilter()
			},
			err: fmt.Errorf("request scheduled Notification.TargetGroup is empty"),
		},
	}
	for _, tcas := range tcases {
		t.Run(tcas.name, func(t *testing.T) {
			noti := &cpb.Notification{}
			assert.NoError(t, faker.FakeData(noti))
			tcas.setup(noti)
			err := ValidateTargetGroup(noti)
			assert.Equal(t, tcas.err, err)
		})
	}

}

func emptyCourseFilter() *cpb.NotificationTargetGroup_CourseFilter {
	if rand.Intn(2) == 0 {
		return &cpb.NotificationTargetGroup_CourseFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	}
	return &cpb.NotificationTargetGroup_CourseFilter{
		Type:      cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		CourseIds: nil,
	}
}

func emptyGradeFilter() *cpb.NotificationTargetGroup_GradeFilter {
	if rand.Intn(2) == 0 {
		return &cpb.NotificationTargetGroup_GradeFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	}
	return &cpb.NotificationTargetGroup_GradeFilter{
		Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		GradeIds: nil,
	}
}

func emptyLocationFilter() *cpb.NotificationTargetGroup_LocationFilter {
	return &cpb.NotificationTargetGroup_LocationFilter{
		Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
	}
}

func emptyClassFilter() *cpb.NotificationTargetGroup_ClassFilter {
	if rand.Intn(2) == 0 {
		return &cpb.NotificationTargetGroup_ClassFilter{
			Type: cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE,
		}
	}
	return &cpb.NotificationTargetGroup_ClassFilter{
		Type:     cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST,
		ClassIds: nil,
	}
}
