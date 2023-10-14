package core

import (
	"github.com/jackc/pgtype"
)

type OnlineUser struct {
	ID           pgtype.Text
	UserID       pgtype.Text
	NodeName     pgtype.Text
	LastActiveAt pgtype.Timestamptz
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
}

func (e *OnlineUser) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"online_user_id", "user_id", "node_name", "last_active_at", "created_at", "updated_at"}
	values = []interface{}{&e.ID, &e.UserID, &e.NodeName, &e.LastActiveAt, &e.CreatedAt, &e.UpdatedAt}
	return
}

func (*OnlineUser) TableName() string {
	return "online_users"
}
