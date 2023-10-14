package accesscontrol

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
)

var testResourcePath = "-212348"

func (s *suite) addUserGroup(ctx context.Context, userGroup string) (string, error) {
	command := `INSERT INTO public.user_group
					(user_group_id, user_group_name, created_at, updated_at, deleted_at, resource_path, org_location_id, is_system)
					VALUES($1, $2, now(), now(), NULL, $3, $4, false)
	 				on conflict (user_group_id) do nothing;`
	userGroupId := s.newID()
	var userGroupIdText pgtype.Text
	var userGroupText pgtype.Text
	_ = userGroupText.Set(userGroup)
	_ = userGroupIdText.Set(userGroupId)
	_, err := connections.MasterMgmtDB.Exec(ctx, command, userGroupIdText, userGroupText, getTextResourcePath(), userGroupText)
	if err != nil {
		return "", fmt.Errorf("cannot insert user group: %v", err)
	}
	return userGroupId, nil
}

func addNewUserToDefaultUserGroupMember(ctx context.Context, userId string, userGroupId string) error {
	command := `INSERT INTO public.user_group_member
							(user_id, user_group_id, created_at, updated_at, deleted_at, resource_path)
							VALUES($1, $2, now(), now(), NULL, $3) on conflict (user_id, user_group_id) do nothing;`
	var userIdText pgtype.Text
	_ = userIdText.Set(userId)

	var userGroupIdText pgtype.Text
	_ = userGroupIdText.Set(userGroupId)
	_, err := connections.MasterMgmtDB.Exec(ctx, command, userIdText, userGroupIdText, getTextResourcePath())
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot insert user group member: %v", err))
	}
	return nil
}

func insertLocation(ctx context.Context, location string, accessPath string) error {
	stmt := `INSERT INTO public.locations
						(location_id, "name", created_at, updated_at, deleted_at, resource_path, access_path)
						VALUES($1, $2, now(), now(), NULL, $3, $4) on conflict (location_id) do nothing;
	`
	var locationText pgtype.Text
	_ = locationText.Set(location)
	var accessPathText pgtype.Text
	_ = accessPathText.Set(accessPath)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, locationText, locationText, getTextResourcePath(), accessPathText)
	if err != nil {
		return fmt.Errorf("cannot insert location: %v", err)
	}
	return nil
}

func addLocationToGrantedRoleAccessPath(ctx context.Context, locationId string, grantedRoleId string) error {
	stmt := `INSERT INTO granted_role_access_path (granted_role_id,location_id,created_at,updated_at,deleted_at,resource_path) VALUES
	 												($1,$2,now(),now(),NULL,$3) on conflict (granted_role_id, location_id) do nothing;`
	var locationText pgtype.Text
	_ = locationText.Set(locationId)
	var grantedRoleIdText pgtype.Text
	_ = grantedRoleIdText.Set(grantedRoleId)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, grantedRoleIdText, locationText, getTextResourcePath())
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot insert granted role access path: %v", err))
	}
	return nil
}

func (s *suite) insertGrantedRole(ctx context.Context, userGroupId string, roleId string, grantedRoleId string) (string, error) {
	stmt := `INSERT INTO granted_role (granted_role_id,user_group_id,role_id,created_at,updated_at,deleted_at,resource_path) VALUES
						($1,$2,$3,now(),now(),NULL,$4)
						on conflict (granted_role_id) do nothing;`
	var roleIdText pgtype.Text
	_ = roleIdText.Set(roleId)
	var userGroupIdText pgtype.Text
	_ = userGroupIdText.Set(userGroupId)

	var grantedRoleIdText pgtype.Text
	_ = grantedRoleIdText.Set(grantedRoleId)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, grantedRoleIdText, userGroupIdText, roleIdText, getTextResourcePath())
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot insert granted role access path: %v", err))
	}
	return grantedRoleId, nil
}

func getTextResourcePath() pgtype.Text {
	var resourcePathText pgtype.Text
	_ = resourcePathText.Set(testResourcePath)
	return resourcePathText
}

func (s *suite) insertPermission(ctx context.Context, permission string) (string, error) {
	stmt := `INSERT INTO "permission" (permission_id,permission_name,created_at,updated_at,deleted_at,resource_path) VALUES
									  ($1,$2,now(),now(),NULL,$3) on conflict (permission_id) do nothing;`
	id := s.newID()
	var permissionText pgtype.Text
	var idText pgtype.Text
	_ = permissionText.Set(permission)
	_ = idText.Set(id)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText, permissionText, getTextResourcePath())
	if err != nil {
		return "", fmt.Errorf("cannot insert permission: %v", err)
	}
	return id, nil
}

func (s *suite) insertRole(ctx context.Context, roleName string) (string, error) {
	stmt := `INSERT INTO "role" (role_id,role_name,created_at,updated_at,deleted_at,resource_path) VALUES
						($1,$2,now(),now(),NULL,$3)
						on conflict (role_id) do nothing;`
	id := s.newID()
	var roleNameText pgtype.Text
	var idText pgtype.Text
	_ = roleNameText.Set(roleName)
	_ = idText.Set(id)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText, roleNameText, getTextResourcePath())
	if err != nil {
		return "", fmt.Errorf("cannot insert role: %v", err)
	}
	return id, nil
}

func insertRolePermission(ctx context.Context, permissionId string, roleId string) error {
	stmt := `INSERT INTO permission_role (permission_id,role_id,created_at,updated_at,deleted_at,resource_path) VALUES
										($1,$2,now(),now(),NULL,$3)
										on conflict (permission_id, role_id) do nothing;`
	var permissionIdText pgtype.Text
	var roleIdText pgtype.Text
	_ = permissionIdText.Set(permissionId)
	_ = roleIdText.Set(roleId)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, permissionIdText, roleIdText, getTextResourcePath())
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot insert role permission: %v", err))
	}
	return nil
}
