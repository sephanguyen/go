package model

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type SystemNotificationRecipient struct {
	SystemNotificationRecipientID pgtype.Text
	SystemNotificationID          pgtype.Text
	UserID                        pgtype.Text
	CreatedAt                     pgtype.Timestamptz
	UpdatedAt                     pgtype.Timestamptz
	DeletedAt                     pgtype.Timestamptz
}

func (*SystemNotificationRecipient) TableName() string {
	return "system_notification_recipients"
}

func (e *SystemNotificationRecipient) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"system_notification_recipient_id",
		"system_notification_id",
		"user_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.SystemNotificationRecipientID,
		&e.SystemNotificationID,
		&e.UserID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

type SystemNotificationRecipients []*SystemNotificationRecipient

func (u *SystemNotificationRecipients) Add() database.Entity {
	e := &SystemNotificationRecipient{}
	*u = append(*u, e)

	return e
}
