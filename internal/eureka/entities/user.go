package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

// UserGroup enum value in DB
var (
	UserGroupStudent             = "USER_GROUP_STUDENT"
	UserGroupAdmin               = "USER_GROUP_ADMIN"
	UserGroupTeacher             = "USER_GROUP_TEACHER"
	UserGroupParent              = "USER_GROUP_PARENT"
	UserGroupSchoolAdmin         = "USER_GROUP_SCHOOL_ADMIN"
	UserGroupOrganizationManager = "USER_GROUP_ORGANIZATION_MANAGER"
)

type User struct {
	UserID pgtype.Text
	Name   pgtype.Text
}

func (t *User) FieldMap() ([]string, []interface{}) {
	return []string{
			"user_id",
			"name",
		}, []interface{}{
			&t.UserID,
			&t.Name,
		}
}

func (t *User) TableName() string {
	return "users"
}

type Users []*User

func (u *Users) Add() database.Entity {
	e := &User{}
	*u = append(*u, e)

	return e
}
