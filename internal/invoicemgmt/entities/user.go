package entities

import (
	user_entities "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

	"github.com/jackc/pgtype"
)

type User struct {
	UserID         pgtype.Text
	Name           pgtype.Text
	GivenName      pgtype.Text
	Group          pgtype.Text
	ResourcePath   pgtype.Text
	UserExternalID pgtype.Text
}

func (e *User) Columns() []string {
	return []string{
		"user_id",
		"name",
		"given_name",
		"user_group",
		"resource_path",
		"user_external_id",
	}
}

func (e *User) FieldMap() ([]string, []interface{}) {
	return e.Columns(), []interface{}{
		&e.UserID,
		&e.Name,
		&e.GivenName,
		&e.Group,
		&e.ResourcePath,
		&e.UserExternalID,
	}
}

func (e *User) TableName() string {
	return "users"
}

func (e *User) GetName() string {
	usermgmtUser := user_entities.LegacyUser{
		GivenName: e.GivenName,
		FullName:  e.Name,
	}

	return usermgmtUser.GetName()
}
