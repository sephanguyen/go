package notificationmgmt

import "github.com/jackc/pgtype"

type Notification struct {
	NotificationID     pgtype.Text
	TargetGroups       pgtype.JSONB
	ReceiverIDs        pgtype.TextArray
	GenericReceiverIDs pgtype.TextArray
}

type UserNotification struct {
	UserNotificationID pgtype.Text
	NotificationID     pgtype.Text
	UserID             pgtype.Text
	ParentID           pgtype.Text
	StudentID          pgtype.Text
	ParentName         pgtype.Text
	StudentName        pgtype.Text
}
