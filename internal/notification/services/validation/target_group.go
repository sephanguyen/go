package validation

import (
	"fmt"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func ValidateTargetGroup(noti *cpb.Notification) error {
	if noti.TargetGroup == nil {
		return fmt.Errorf("request Notification.TargetGroup is null")
	}

	if noti.TargetGroup.CourseFilter == nil {
		return fmt.Errorf("request Notification.TargetGroup.CourseFilter is null")
	}

	if noti.TargetGroup.GradeFilter == nil {
		return fmt.Errorf("request Notification.TargetGroup.GradeFilter is null")
	}

	if noti.TargetGroup.LocationFilter == nil {
		return fmt.Errorf("request Notification.TargetGroup.LocationFilter is null")
	}
	locationFilter := noti.TargetGroup.LocationFilter
	if locationFilter.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST && len(locationFilter.LocationIds) == 0 {
		return fmt.Errorf("request Notification.TargetGroup.LocationFilter.Type is ListSelected, but locations list is empty")
	}

	if noti.TargetGroup.ClassFilter == nil {
		return fmt.Errorf("request Notification.TargetGroup.ClassFilter is null")
	}

	if noti.TargetGroup.UserGroupFilter == nil {
		return fmt.Errorf("request Notification.TargetGroup.UserGroupFilter is null")
	}

	if len(noti.TargetGroup.UserGroupFilter.UserGroups) == 0 {
		return fmt.Errorf("request Notification.TargetGroup.UserGroupFilter.UserGroup is empty")
	}

	switch noti.Status {
	case cpb.NotificationStatus_NOTIFICATION_STATUS_SCHEDULED:
		// case no receiver ids
		if len(noti.ReceiverIds) == 0 && len(noti.GenericReceiverIds) == 0 {
			if CheckCourseFilterEmpty(noti.TargetGroup.CourseFilter) &&
				CheckGradeFilterEmpty(noti.TargetGroup.GradeFilter) &&
				CheckLocationFilterEmpty(noti.TargetGroup.LocationFilter) &&
				CheckClassFilterEmpty(noti.TargetGroup.ClassFilter) {
				return fmt.Errorf("request scheduled Notification.TargetGroup is empty")
			}
		}
	}

	return nil
}

func CheckCourseFilterEmpty(f *cpb.NotificationTargetGroup_CourseFilter) bool {
	return f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE ||
		(f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST && len(f.CourseIds) == 0)
}

func CheckGradeFilterEmpty(f *cpb.NotificationTargetGroup_GradeFilter) bool {
	return f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE ||
		(f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST && len(f.GradeIds) == 0)
}

func CheckLocationFilterEmpty(f *cpb.NotificationTargetGroup_LocationFilter) bool {
	return f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE ||
		(f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST && len(f.LocationIds) == 0)
}

func CheckClassFilterEmpty(f *cpb.NotificationTargetGroup_ClassFilter) bool {
	return f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE ||
		(f.Type == cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST && len(f.ClassIds) == 0)
}
