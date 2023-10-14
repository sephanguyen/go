package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

const (
	UserGroupStatusActive   = "USER_GROUP_STATUS_ACTIVE"
	UserGroupStatusInActive = "USER_GROUP_STATUS_INACTIVE"
)

type UserGroup struct {
	UserID       pgtype.Text
	GroupID      pgtype.Text
	IsOrigin     pgtype.Bool
	Status       pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (rcv *UserGroup) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "group_id", "is_origin", "status", "created_at", "updated_at", "resource_path"}
	values = []interface{}{&rcv.UserID, &rcv.GroupID, &rcv.IsOrigin, &rcv.Status, &rcv.CreatedAt, &rcv.UpdatedAt, &rcv.ResourcePath}
	return
}

func (*UserGroup) TableName() string {
	return "users_groups"
}

// UserGroups type alias for working with database helper
type UserGroups []*UserGroup

// Add appends another UserGroup to itself
func (u *UserGroups) Add() database.Entity {
	e := &UserGroup{}
	*u = append(*u, e)

	return e
}
