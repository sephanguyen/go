package repo

import (
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"

	"github.com/jackc/pgtype"
)

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

func ToUserPermissionDomain(grantedPermissions []*GrantedPermission) *domain.UserPermissions {
	permissions := make([]string, len(grantedPermissions))
	locationIDs := make([]string, len(grantedPermissions))
	grantedLocations := map[string][]string{}

	for _, permission := range grantedPermissions {
		permissions = append(permissions, permission.PermissionName.String)
		locationIDs = append(locationIDs, permission.LocationID.String)
		grantedLocations[permission.PermissionName.String] = append(grantedLocations[permission.PermissionName.String], permission.LocationID.String)
	}
	userGroupPermission := domain.UserPermissions{
		Permissions:      golibs.Uniq(permissions),
		LocationIDs:      golibs.Uniq(locationIDs),
		GrantedLocations: grantedLocations,
	}
	return &userGroupPermission
}
