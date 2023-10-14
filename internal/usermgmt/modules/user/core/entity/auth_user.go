package entity

import (
	"github.com/jackc/pgtype"
)

type AuthUser struct {
	UserID        pgtype.Text
	UserGroup     pgtype.Text
	Email         pgtype.Text
	LoginEmail    pgtype.Text
	DeactivatedAt pgtype.Timestamptz
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
}

func (au *AuthUser) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "user_group", "email", "login_email", "deactivated_at", "created_at", "updated_at", "deleted_at", "resource_path"}
	values = []interface{}{&au.UserID, &au.UserGroup, &au.Email, &au.LoginEmail, &au.DeactivatedAt, &au.CreatedAt, &au.UpdatedAt, &au.DeletedAt, &au.ResourcePath}
	return
}

func (*AuthUser) TableName() string {
	return "users"
}
