package entity

import (
	"github.com/jackc/pgtype"
)

type AppleUser struct {
	ID        pgtype.Text
	UserID    pgtype.Text
	UpdatedAt pgtype.Timestamptz
	CreatedAt pgtype.Timestamptz
}

func (rcv *AppleUser) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"apple_user_id", "user_id", "updated_at", "created_at"}
	values = []interface{}{&rcv.ID, &rcv.UserID, &rcv.UpdatedAt, &rcv.CreatedAt}
	return
}

func (*AppleUser) TableName() string {
	return "apple_users"
}
