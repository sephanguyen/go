package entity

import (
	"github.com/jackc/pgtype"
)

type GrantedRoleAccessPath struct {
	GrantedRoleID pgtype.Text
	LocationID    pgtype.Text
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
}

func (g *GrantedRoleAccessPath) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"granted_role_id", "location_id", "created_at", "updated_at", "deleted_at", "resource_path"}
	values = []interface{}{&g.GrantedRoleID, &g.LocationID, &g.CreatedAt, &g.UpdatedAt, &g.DeletedAt, &g.ResourcePath}
	return
}

func (*GrantedRoleAccessPath) TableName() string {
	return "granted_role_access_path"
}
