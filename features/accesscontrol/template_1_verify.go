package accesscontrol

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgtype"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
)

func insertTestTableAccessPath(ctx context.Context, locationId string, acTestTemplate1Id string) error {
	stmt := `INSERT INTO public.ac_test_template_1_access_paths
						(ac_test_template_1_id, location_id, access_path, created_at, updated_at, deleted_at, resource_path)
						VALUES($1, $2, $3, now(), now(), NULL, $4)
						on conflict (ac_test_template_1_id, location_id) do nothing;`
	var locationText pgtype.Text
	_ = locationText.Set(locationId)
	var acTestTemplate1IdText pgtype.Text
	_ = acTestTemplate1IdText.Set(acTestTemplate1Id)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, acTestTemplate1IdText, locationText, locationText, getTextResourcePath())
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot insert granted role access path: %v", err))
	}
	return nil
}

func insertTestTable(ctx context.Context, id string) error {
	stmt := `INSERT INTO public.ac_test_template_1
						(ac_test_template_1_id, created_at, updated_at, deleted_at, resource_path)
						VALUES($1, now(), now(), NULL,$2);`
	var idText pgtype.Text
	_ = idText.Set(id)
	_, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText, getTextResourcePath())
	if err != nil {
		return err
	}
	return nil
}

func deleteTestTable(ctx context.Context, id string) (int64, error) {
	stmt := `delete from public.ac_test_template_1 where ac_test_template_1_id = $1`
	var idText pgtype.Text
	_ = idText.Set(id)
	deletedRecords, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText)
	if err != nil {
		return 0, err
	}
	return deletedRecords.RowsAffected(), nil
}

func updateTestTable(ctx context.Context, id string) (int64, error) {
	stmt := `update public.ac_test_template_1 set updated_at = now() where ac_test_template_1_id = $1`
	var idText pgtype.Text
	_ = idText.Set(id)
	updatedRecords, err := connections.MasterMgmtDB.Exec(ctx, stmt, idText)
	if err != nil {
		return 0, err
	}
	return updatedRecords.RowsAffected(), nil
}

func (s *suite) loginAsUserIdAndGroup(ctx context.Context, userId string, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	userGroupId, err := s.addUserGroup(ctx, group)
	if err != nil {
		return ctx, err
	}
	addNewUserToDefaultUserGroupMember(ctx, userId, userGroupId)

	claim := interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: testResourcePath,
			DefaultRole:  group,
			UserGroup:    group,
			UserID:       userId,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, &claim)

	stepState.CurrentUserID = userId
	stepState.CurrentUserGroup = userGroupId
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) getTotalTestRecords(ctx context.Context, id string) (int, error) {
	stmt := `select ac_test_template_1_id from ac_test_template_1 att where ac_test_template_1_id = $1;`
	var idText pgtype.Text
	_ = idText.Set(id)
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
			return 0, fmt.Errorf("scan id fail. %v", err)
		}
		ids = append(ids, acTestTemplate1Id)
	}
	return len(ids), nil
}

func (s *suite) tableBWithLocationAndPermission(ctx context.Context, location, permission string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.CurrentTestId = "new data"
	err := insertTestTableAccessPath(ctx, location, stepState.CurrentTestId)
	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userAssignedAnd(ctx context.Context, location, permission string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.LocationID = location

	err := insertLocation(ctx, location, location)

	stepState.PermissionID, err = s.insertPermission(ctx, permission)
	stepState.RoleId, err = s.insertRole(ctx, permission)
	err = insertRolePermission(ctx, stepState.PermissionID, stepState.RoleId)
	stepState.GrantedRoleId, err = s.insertGrantedRole(ctx, stepState.CurrentUserGroup, stepState.RoleId, s.newID())
	err = addLocationToGrantedRoleAccessPath(ctx, location, stepState.GrantedRoleId)

	if err != nil {
		zapLogger.Fatal(fmt.Sprintf("cannot init auth info: %v", err))
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGetDataFromTableB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	totalRecords, err := s.getTotalTestRecords(ctx, stepState.CurrentTestId)

	if err != nil {
		return ctx, err
	}

	stepState.TotalRecords = totalRecords

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userShouldOnlyGetTheirAssigned(ctx context.Context, records string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	numRecords, err := strconv.Atoi(records)
	if err != nil {
		return ctx, err
	}

	if stepState.TotalRecords != numRecords {
		return ctx, fmt.Errorf("Number of record expected and actual is difference %d/%d", numRecords, stepState.TotalRecords)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userDataBelongToLocationIntoTableAcTestTemplate1(ctx context.Context, command string, data string, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	insertTestTableAccessPath(ctx, location, data)

	switch command {
	case "insert":
		stepState.ErrInfo = insertTestTable(ctx, data)
	case "delete":
		rowAffects, err := deleteTestTable(ctx, data)
		if err != nil {
			return ctx, err
		}
		if rowAffects == 0 {
			stepState.ErrInfo = fmt.Errorf("User can't delete record %d", rowAffects)
		}
	case "update":
		rowAffects, err := updateTestTable(ctx, data)
		if err != nil {
			return ctx, err
		}
		if rowAffects == 0 {
			stepState.ErrInfo = fmt.Errorf("User can't update record %d", rowAffects)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func invalidRLS(err error) bool {
	return strings.Contains(err.Error(), "violates row-level security policy")
}

func (s *suite) userDataIntoTableAcTestTemplate1(ctx context.Context, command string, data string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.ErrInfo = nil
	switch command {
	case "insert":
		err := insertTestTable(ctx, data)
		if !invalidRLS(err) {
			stepState.ErrInfo = err
		}
	case "delete":
		rowAffects, err := deleteTestTable(ctx, data)
		if err != nil && !invalidRLS(err) {
			return ctx, err
		}
		if rowAffects > 0 {
			stepState.ErrInfo = fmt.Errorf("User should not delete this record %d", rowAffects)
		}
	case "update":
		rowAffects, err := updateTestTable(ctx, data)
		if err != nil && !invalidRLS(err) {
			return ctx, err
		}
		if rowAffects > 0 {
			stepState.ErrInfo = fmt.Errorf("User should not update this record %d", rowAffects)
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnSuccess(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ErrInfo != nil {
		return ctx, fmt.Errorf("Error returnSuccess %v", stepState.ErrInfo)
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnFail(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ErrInfo != nil {
		return ctx, fmt.Errorf("Error returnFail %v", stepState.ErrInfo)
	}
	return StepStateToContext(ctx, stepState), nil
}
