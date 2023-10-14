package model

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type SystemNotification struct {
	SystemNotificationID pgtype.Text
	ReferenceID          pgtype.Text
	URL                  pgtype.Text
	ValidFrom            pgtype.Timestamptz
	ValidTo              pgtype.Timestamptz
	Status               pgtype.Text
	CreatedAt            pgtype.Timestamptz
	UpdatedAt            pgtype.Timestamptz
	DeletedAt            pgtype.Timestamptz
}

func (*SystemNotification) TableName() string {
	return "system_notifications"
}

func (e *SystemNotification) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"system_notification_id",
		"reference_id",
		"url",
		"valid_from",
		"valid_to",
		"status",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.SystemNotificationID,
		&e.ReferenceID,
		&e.URL,
		&e.ValidFrom,
		&e.ValidTo,
		&e.Status,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

type SystemNotifications []*SystemNotification

func (u *SystemNotifications) Add() database.Entity {
	e := &SystemNotification{}
	*u = append(*u, e)

	return e
}
