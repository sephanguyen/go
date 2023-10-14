package support

import (
	"github.com/jackc/pgtype"
)

type UserGroupMember struct {
	UserID      pgtype.Text
	UserGroupID pgtype.Text
}

func (g *UserGroupMember) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "user_group_id"}
	values = []interface{}{
		&g.UserID,
		&g.UserGroupID,
	}

	return
}

func (*UserGroupMember) TableName() string {
	return "user_group_member"
}
