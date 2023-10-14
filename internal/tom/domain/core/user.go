package core

import (
	"github.com/jackc/pgtype"
)

type User struct {
	UserID            pgtype.Text
	UserGroup         pgtype.Text
	Name              pgtype.Text
	GivenName         pgtype.Text
	Country           pgtype.Text
	DeviceToken       pgtype.Text
	AllowNotification pgtype.Bool
	UpdatedAt         pgtype.Timestamptz
	CreatedAt         pgtype.Timestamptz
	DeletedAt         pgtype.Timestamptz
	ResourcePath      pgtype.Text
}

func (e *User) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"user_group",
			"name",
			"given_name",
			"country",
			"device_token",
			"allow_notification",
			"updated_at",
			"created_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.UserID,
			&e.UserGroup,
			&e.Name,
			&e.GivenName,
			&e.Country,
			&e.DeviceToken,
			&e.AllowNotification,
			&e.UpdatedAt,
			&e.CreatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (e *User) TableName() string {
	return "users"
}
