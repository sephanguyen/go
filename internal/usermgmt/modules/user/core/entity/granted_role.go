package entity

import (
	"github.com/jackc/pgtype"
)

type GrantedRole struct {
	GrantedRoleID pgtype.Text
	UserGroupID   pgtype.Text
	RoleID        pgtype.Text
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
}

func (g *GrantedRole) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"granted_role_id", "user_group_id", "role_id", "created_at", "updated_at", "deleted_at", "resource_path"}
	values = []interface{}{&g.GrantedRoleID, &g.UserGroupID, &g.RoleID, &g.CreatedAt, &g.UpdatedAt, &g.DeletedAt, &g.ResourcePath}
	return
}

func (*GrantedRole) TableName() string {
	return "granted_role"
}
