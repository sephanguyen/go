package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type UserGroupV2 struct {
	UserGroupID   pgtype.Text
	UserGroupName pgtype.Text
	CreatedAt     pgtype.Timestamptz
	UpdatedAt     pgtype.Timestamptz
	DeletedAt     pgtype.Timestamptz
	ResourcePath  pgtype.Text
	OrgLocationID pgtype.Text
	IsSystem      pgtype.Bool
}

func (u *UserGroupV2) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_group_id", "user_group_name", "created_at", "updated_at", "deleted_at", "resource_path", "org_location_id", "is_system"}
	values = []interface{}{&u.UserGroupID, &u.UserGroupName, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.ResourcePath, &u.OrgLocationID, &u.IsSystem}
	return
}

func (*UserGroupV2) TableName() string {
	return "user_group"
}

type UserGroupV2s []*UserGroupV2

func (u *UserGroupV2s) Add() database.Entity {
	e := &UserGroupV2{}
	*u = append(*u, e)
	return e
}
