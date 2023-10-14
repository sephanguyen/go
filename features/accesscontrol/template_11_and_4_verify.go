package accesscontrol

import (
	"context"
	"fmt"

	"github.com/jackc/pgtype"
)

func insertAcTestTemplate11And4TableAccessPath(ctx context.Context, locationId string, id string) error {
	stmt := `INSERT INTO public.ac_test_template_11_4_access_paths
						(ac_test_template_11_4_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
						VALUES($1, $2, $3, now(), now(), NULL, $4)
						on conflict (ac_test_template_11_4_id, location_id) do nothing;`
	var locationText pgtype.Text
	_ = locationText.Set(locationId)
	var idText pgtype.Text
	_ = idText.Set(id)

	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText, locationText, locationText, getTextResourcePath())
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot insert granted role access path: %v", err))
	}
	return nil
}

func insertAcTestTemplate11And4Table(ctx context.Context, id string, owners string, name string) error {
	stmt := `INSERT INTO public.ac_test_template_11_4
						(ac_test_template_11_4_id, created_at, updated_at, deleted_at, resource_path, owners, "name")
						VALUES($1, now(), now(), NULL,$2, $3, $4);`
	var idText pgtype.Text
	_ = idText.Set(id)
	var ownersText pgtype.Text
	_ = ownersText.Set(owners)
	var nameText pgtype.Text
	_ = nameText.Set(name)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText, getTextResourcePath(), ownersText, nameText)
	if err != nil {
		return err
	}
	return nil
}

func updateAcTestTemplate11And4Table(ctx context.Context, id string, name string) (int64, error) {
	stmt := `update public.ac_test_template_11_4 set updated_at = now(), "name" = $1 where ac_test_template_11_4_id = $2`
	var idText pgtype.Text
	_ = idText.Set(id)
	var nameText pgtype.Text
	_ = nameText.Set(name)
	updatedRecords, err := connections.MasterMgmtDB.Exec(ctx, stmt, nameText, idText)
	if err != nil {
		return 0, err
	}
	return updatedRecords.RowsAffected(), nil
}

func deleteAcTestTemplate11And4Table(ctx context.Context, id string) (int64, error) {
	stmt := `delete from public.ac_test_template_11_4 where ac_test_template_11_4_id = $1`
	var idText pgtype.Text
	_ = idText.Set(id)
	deletedRecords, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText)
	if err != nil {
		return 0, err
	}
	return deletedRecords.RowsAffected(), nil
}

func getAcTestTemplate11And4Data(ctx context.Context, data string) (string, error) {
	stmt := `select ac_test_template_11_4_id from ac_test_template_11_4 att where ac_test_template_11_4_id = $1;`
	var idText pgtype.Text
	_ = idText.Set(data)
	rows, err := connections.MasterMgmtDB.Query(ctx, stmt, idText)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("Get user error %v", err))
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		acTestTemplate1Id := ""
		err := rows.Scan(&acTestTemplate1Id)
		if err != nil {
			return "", fmt.Errorf("scan id fail. %v", err)
		}
		ids = append(ids, acTestTemplate1Id)
	}
	if len(ids) == 0 {
		return "", nil
	}
	return ids[0], nil
}

func (s *suite) addDataWithOwnerIsToTableAcTestTemplate11And4(ctx context.Context, data, owner string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := insertAcTestTemplate11And4Table(ctx, data, owner, data)
	stepState.ErrInfo = err
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addDataWithOwnerIsWithLocationToTableAcTestTemplate11And4(ctx context.Context, data, owner, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := insertAcTestTemplate11And4TableAccessPath(ctx, location, data)
	stepState.ErrInfo = err
	if err != nil {
		return ctx, err
	}
	err = insertAcTestTemplate11And4Table(ctx, data, owner, data)
	stepState.ErrInfo = err
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addPermissionAndToPermissionRole(ctx context.Context, _, _, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := insertRolePermission(ctx, stepState.PermissionReadId, stepState.RoleId)
	if err != nil {
		return ctx, err
	}
	err = insertRolePermission(ctx, stepState.PermissionWriteId, stepState.RoleId)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addRoleAndUserGroupToGrantedRole(ctx context.Context, role, _, grantedRole string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	_, err := s.insertGrantedRole(ctx, stepState.ExistedUserGroupID, stepState.RoleId, grantedRole)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addUserToUserGroup(ctx context.Context, user, _ string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := addNewUserToDefaultUserGroupMember(ctx, user, stepState.ExistedUserGroupID)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) assignLocationToGrantedRole(ctx context.Context, location, grantedRole string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := addLocationToGrantedRoleAccessPath(ctx, location, grantedRole)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) commandShouldReturnTheRecordsUserIsOwners(ctx context.Context, recordData string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.CurrentTestData != recordData {
		return ctx, fmt.Errorf("Error when check record data %s and %s", recordData, stepState.CurrentTestData)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createLocationWithAccessPathWithParent(ctx context.Context, location, accessPath, parent string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	err := insertLocation(ctx, location, accessPath)
	if err != nil {
		return ctx, err
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createPermissionWithName(ctx context.Context, permissionName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	permissionId, err := s.insertPermission(ctx, permissionName)
	if err != nil {
		return ctx, err
	}
	if permissionName == "accesscontrol.ac_test_template_11_4.read" {
		stepState.PermissionReadId = permissionId
	}
	if permissionName == "accesscontrol.ac_test_template_11_4.write" {
		stepState.PermissionWriteId = permissionId
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createRoleWithName(ctx context.Context, roleName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	roleId, err := s.insertRole(ctx, roleName)
	if err != nil {
		return ctx, err
	}
	stepState.RoleId = roleId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserGroupName(ctx context.Context, userGroupName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	userGroupId, err := s.addUserGroup(ctx, userGroupName)
	if err != nil {
		return ctx, err
	}
	stepState.ExistedUserGroupID = userGroupId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetDataFromTableAcTestTemplate11And4(ctx context.Context, data string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	data, err := getAcTestTemplate11And4Data(ctx, data)
	if err != nil {
		return ctx, err
	}
	stepState.CurrentTestData = data
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateDataWithNameToTableAcTestTemplate11And4(ctx context.Context, id, name string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rowAffected, err := updateAcTestTemplate11And4Table(ctx, id, name)
	if err != nil {
		return ctx, err
	}

	if rowAffected == 0 {
		stepState.ErrInfo = fmt.Errorf("The command can't update any record")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDeleteDataFromTableAcTestTemplate11And4(ctx context.Context, id string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	rowAffected, err := deleteAcTestTemplate11And4Table(ctx, id)
	if err != nil {
		return ctx, err
	}

	if rowAffected == 0 {
		stepState.ErrInfo = fmt.Errorf("The command can't delete any record")
	}

	return StepStateToContext(ctx, stepState), nil
}
