package entities

import (
	user_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
)

type User struct {
	ID                pgtype.Text `sql:"user_id,pk"`
	Group             pgtype.Text `sql:"user_group"`
	FullName          pgtype.Text `sql:"name"`
	GivenName         pgtype.Text
	Country           pgtype.Text
	DeviceToken       pgtype.Text
	AllowNotification pgtype.Bool `sql:",notnull"`
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
			&e.ID,
			&e.Group,
			&e.FullName,
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

func (e *User) GetName() string {
	usermgmtUser := user_entities.LegacyUser{
		GivenName: e.GivenName,
		FullName:  e.FullName,
	}

	return usermgmtUser.GetName()
}
