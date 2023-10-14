package model

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type SystemNotificationContent struct {
	SystemNotificationContentID pgtype.Text
	SystemNotificationID        pgtype.Text
	Language                    pgtype.Text
	Text                        pgtype.Text
	CreatedAt                   pgtype.Timestamptz
	UpdatedAt                   pgtype.Timestamptz
	DeletedAt                   pgtype.Timestamptz
}

func (*SystemNotificationContent) TableName() string {
	return "system_notification_contents"
}

func (e *SystemNotificationContent) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"system_notification_content_id",
		"system_notification_id",
		"language",
		"text",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.SystemNotificationContentID,
		&e.SystemNotificationID,
		&e.Language,
		&e.Text,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

type SystemNotificationContents []*SystemNotificationContent

func (u *SystemNotificationContents) Add() database.Entity {
	e := &SystemNotificationContent{}
	*u = append(*u, e)

	return e
}
