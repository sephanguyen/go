package consts

import cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"

const (
	SendingMethodPushNotification = "push_notification"
	NotificationTypeImmediate     = "immediate"
	NotificationTypeScheduled     = "schedule"
	NotificationModeMute          = "mute"
	NotificationModeNotify        = "notify"

	TargetUserGroupParent  = "USER_GROUP_PARENT"
	TargetUserGroupStudent = "USER_GROUP_STUDENT"

	ErrorRecodeNotFound = "recode not found"

	LocalEnv   = "local"
	StagingEnv = "stag"
	UATEnv     = "uat"
	ProdEnv    = "prod"

	NotificationWritePermission = "communication.notification.write"
	NotificationReadPermission  = "communication.notification.read"
	NotificationOwnerPermission = "communication.notification.owner"

	UserReadPermission = "user.user.read"

	AllowTagCSVHeaders = "tag_id|tag_name|is_archived"

	DefaultOrder    = "default"
	AscendingOrder  = "asc"
	DescendingOrder = "desc"

	TargetGroupSelectTypeNone = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_NONE
	TargetGroupSelectTypeAll  = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_ALL
	TargetGroupSelectTypeList = cpb.NotificationTargetGroupSelect_NOTIFICATION_TARGET_GROUP_SELECT_LIST

	// nolint
	DateTimeCSVFormat = "2006/01/02, 15:04:05"

	ServiceName = "notificationmgmt"
)
