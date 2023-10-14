package core

import "github.com/jackc/pgtype"

type UserDeviceToken struct {
	ID                pgtype.Int4
	UserID            pgtype.Text
	UserName          pgtype.Text
	Token             pgtype.Text
	AllowNotification pgtype.Bool
	CreatedAt         pgtype.Timestamptz
	UpdatedAt         pgtype.Timestamptz
}

func (u *UserDeviceToken) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_device_token_id", "user_id", "user_name", "token", "allow_notification", "created_at", "updated_at"}
	values = []interface{}{&u.ID, &u.UserID, &u.UserName, &u.Token, &u.AllowNotification, &u.CreatedAt, &u.UpdatedAt}
	return
}
func (*UserDeviceToken) TableName() string {
	return "user_device_tokens"
}
