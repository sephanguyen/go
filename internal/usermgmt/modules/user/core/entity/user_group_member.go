package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type UserGroupMember struct {
	UserID       pgtype.Text
	UserGroupID  pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (ugm *UserGroupMember) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "user_group_id", "created_at", "updated_at", "deleted_at", "resource_path"}
	values = []interface{}{&ugm.UserID, &ugm.UserGroupID, &ugm.CreatedAt, &ugm.UpdatedAt, &ugm.DeletedAt, &ugm.ResourcePath}
	return
}

func (*UserGroupMember) TableName() string {
	return "user_group_member"
}

type UserGroupMembers []*UserGroupMember

func (u *UserGroupMembers) Add() database.Entity {
	e := &UserGroupMember{}
	*u = append(*u, e)
	return e
}
