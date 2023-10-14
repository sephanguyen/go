package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type Role struct {
	RoleID       pgtype.Text
	RoleName     pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
	IsSystem     pgtype.Bool
}

func (r *Role) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"role_id", "role_name", "created_at", "updated_at", "deleted_at", "resource_path", "is_system"}
	values = []interface{}{&r.RoleID, &r.RoleName, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt, &r.ResourcePath, &r.IsSystem}
	return
}

func (*Role) TableName() string {
	return "role"
}

type Roles []*Role

func (r *Roles) Add() database.Entity {
	e := &Role{}
	*r = append(*r, e)
	return e
}

func (r *Roles) ListRoleNames() []string {
	mapRoleNames := map[string]struct{}{}
	for _, role := range *r {
		mapRoleNames[role.RoleName.String] = struct{}{}
	}
	listRoleNames := []string{}
	for roleName := range mapRoleNames {
		listRoleNames = append(listRoleNames, roleName)
	}

	return listRoleNames
}
