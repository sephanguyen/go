package entities

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type UserDeviceToken struct {
	UserDeviceTokenID pgtype.Int4
	UserID            pgtype.Text
	DeviceToken       pgtype.Text
	AllowNotification pgtype.Bool
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
}

func (u *UserDeviceToken) FieldMap() (fields []string, values []interface{}) {
	fields = []string{
		"user_device_token_id",
		"user_id",
		"device_token",
		"allow_notification",
		"created_at",
		"updated_at"}
	values = []interface{}{&u.UserDeviceTokenID, &u.UserID, &u.DeviceToken, &u.AllowNotification, &u.CreatedAt, &u.UpdatedAt}
	return
}
func (*UserDeviceToken) TableName() string {
	return "user_device_tokens"
}

type UserDeviceTokens []*UserDeviceToken

func (udt *UserDeviceTokens) Add() database.Entity {
	e := &UserDeviceToken{}
	*udt = append(*udt, e)

	return e
}
