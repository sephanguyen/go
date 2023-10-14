package entities

import "github.com/jackc/pgtype"

type PermissionRole struct {
	PermissionID pgtype.Text
	RoleID       pgtype.Text
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
	DeletedAt    pgtype.Timestamptz
	ResourcePath pgtype.Text
}

func (e *PermissionRole) FieldMap() (fields []string, values []interface{}) {
	return []string{
			"permission_id",
			"role_id",
			"created_at",
			"updated_at",
			"deleted_at",
			"resource_path",
		}, []interface{}{
			&e.PermissionID,
			&e.RoleID,
			&e.CreatedAt,
			&e.UpdatedAt,
			&e.DeletedAt,
			&e.ResourcePath,
		}
}

func (*PermissionRole) TableName() string {
	return "permission_role"
}
