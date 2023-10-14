package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type NotificationInternalUser struct {
	UserID    pgtype.Text
	IsSystem  pgtype.Bool
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
	DeletedAt pgtype.Timestamptz
}

type NotificationInternalUsers []*NotificationInternalUser

func (e *NotificationInternalUser) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_id",
		"is_system",
		"created_at",
		"updated_at",
		"deleted_at",
	}
	values = []interface{}{
		&e.UserID,
		&e.IsSystem,
		&e.CreatedAt,
		&e.UpdatedAt,
		&e.DeletedAt,
	}
	return
}

func (e *NotificationInternalUser) TableName() string { return "notification_internal_user" }

func (ss *NotificationInternalUsers) Add() database.Entity {
	e := &NotificationInternalUser{}
	*ss = append(*ss, e)

	return e
}
