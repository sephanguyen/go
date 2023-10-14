package entities

import (
	"github.com/jackc/pgtype"
)

type Group struct {
	ID          pgtype.Text
	Name        pgtype.Text
	Description pgtype.Text
	Privileges  pgtype.JSONB
	CreatedAt   pgtype.Timestamptz
	UpdatedAt   pgtype.Timestamptz
}

func (rcv *Group) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"group_id", "name", "description", "privileges", "created_at", "updated_at"}
	values = []interface{}{&rcv.ID, &rcv.Name, &rcv.Description, &rcv.Privileges, &rcv.CreatedAt, &rcv.UpdatedAt}
	return
}

func (*Group) TableName() string {
	return "groups"
}
