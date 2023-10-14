package core

import "github.com/jackc/pgtype"

type GrantedPermission struct {
	UserID         pgtype.Text
	PermissionName pgtype.Text
	LocationID     pgtype.Text
	PermissionID   pgtype.Text
}

func (g *GrantedPermission) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"user_id", "permission_name", "location_id", "permission_id"}
	values = []interface{}{
		&g.UserID,
		&g.PermissionName,
		&g.LocationID,
		&g.PermissionID,
	}

	return
}

func (*GrantedPermission) TableName() string {
	return "granted_permissions"
}
