package entity

import "github.com/jackc/pgtype"

type Permission struct {
	PermissionID   pgtype.Text
	PermissionName pgtype.Text
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
	ResourcePath   pgtype.Text
}

func (e *Permission) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"permission_id", "permission_name", "created_at", "updated_at", "deleted_at", "resource_path"}
	values = []interface{}{&e.PermissionID, &e.PermissionName, &e.CreatedAt, &e.UpdatedAt, &e.DeletedAt, &e.ResourcePath}
	return
}

func (*Permission) TableName() string {
	return "permission"
}
