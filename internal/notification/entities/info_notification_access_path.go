package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type InfoNotificationAccessPath struct {
	NotificationID pgtype.Text
	LocationID     pgtype.Text
	AccessPath     pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
	CreatedUserID  pgtype.Text
}

type InfoNotificationAccessPaths []*InfoNotificationAccessPath

func (e *InfoNotificationAccessPath) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"notification_id",
		"location_id",
		"access_path",
		"created_at",
		"updated_at",
		"deleted_at",
		"created_user_id",
	}
	values = []interface{}{
		&e.NotificationID,
		&e.LocationID,
		&e.AccessPath,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
		&e.CreatedUserID,
	}
	return
}

func (e *InfoNotificationAccessPath) TableName() string { return "info_notifications_access_paths" }

func (ss *InfoNotificationAccessPaths) Add() database.Entity {
	e := &InfoNotificationAccessPath{}
	*ss = append(*ss, e)

	return e
}
