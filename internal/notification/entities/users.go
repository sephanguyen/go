package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type User struct {
	UserID    pgtype.Text
	Name      pgtype.Text
	FirstName pgtype.Text
	LastName  pgtype.Text
	DeletedAt pgtype.Timestamptz
}

func (u *User) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_id",
		"name",
		"first_name",
		"last_name",
		"deleted_at"}
	values = []interface{}{&u.UserID, &u.Name, &u.FirstName, &u.LastName, &u.DeletedAt}
	return
}
func (*User) TableName() string {
	return "users"
}

type Users []*User

func (udt *Users) Add() database.Entity {
	e := &User{}
	*udt = append(*udt, e)

	return e
}
