package entity

import (
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgtype"
)

type UserAccessPath struct {
	UserID       pgtype.Text
	LocationID   pgtype.Text
	AccessPath   pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

type UserAccessPaths []*UserAccessPath

func (uap *UserAccessPath) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "location_id", "access_path", "created_at", "updated_at", "deleted_at", "resource_path"}
	values = []interface{}{&uap.UserID, &uap.LocationID, &uap.AccessPath, &uap.CreatedAt, &uap.UpdatedAt, &uap.DeletedAt, &uap.ResourcePath}
	return
}

func (*UserAccessPath) TableName() string {
	return "user_access_paths"
}

func (u *UserAccessPaths) Add() database.Entity {
	e := &UserAccessPath{}
	*u = append(*u, e)

	return e
}
