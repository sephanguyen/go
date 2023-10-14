package entities

import "github.com/jackc/pgtype"

type User struct {
	UserID pgtype.Text
	Name   pgtype.Text
	Group  pgtype.Text
}

func (e *User) Columns() []string {
	return []string{
		"user_id",
		"name",
		"user_group",
	}
}

func (e *User) FieldMap() ([]string, []interface{}) {
	return e.Columns(), []interface{}{
		&e.UserID,
		&e.Name,
		&e.Group,
	}
}

func (e *User) TableName() string {
	return "users"
}
