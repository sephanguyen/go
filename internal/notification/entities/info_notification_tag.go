package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type InfoNotificationTag struct {
	NotificationTagID pgtype.Text
	NotificationID    pgtype.Text
	TagID             pgtype.Text
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
}

func (e *InfoNotificationTag) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"notification_tag_id",
		"notification_id",
		"tag_id",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.NotificationTagID,
		&e.NotificationID,
		&e.TagID,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (*InfoNotificationTag) TableName() string {
	return "info_notifications_tags"
}

type InfoNotificationsTags []*InfoNotificationTag

func (u *InfoNotificationsTags) Add() database.Entity {
	e := &InfoNotificationTag{}
	*u = append(*u, e)

	return e
}
