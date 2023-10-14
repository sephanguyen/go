package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"google.golang.org/grpc/status"
)

func (s *suite) userGroupNeedToBeUpdated(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, err := s.generateAUserGroupPayloadWithRoleName(ctx, "valid", constant.RoleTeacher)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	resp, err := CreateUserGroup(ctx, s.BobDBTrace, s.UserMgmtConn, req, nil)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	stepState.ExistedUserGroupID = resp.GetUserGroupId()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateUserGroupWithValidPayloadWithRoleName(ctx context.Context, signedUser string, roleName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, err := generateUpdateUserGroupRequestWithRoleName(ctx, s.BobDBTrace, stepState.ExistedUserGroupID, roleName)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).UpdateUserGroup(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateUserGroupWithValidPayload(ctx context.Context, signedUser string) (context.Context, error) {
	return s.userUpdateUserGroupWithValidPayloadWithRoleName(ctx, signedUser, constant.RoleTeacher)
}

func (s *suite) updateUserGroupSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return nil, stepState.ResponseErr
	}

	if err := s.validUpdatedUserGroup(ctx); err != nil {
		return nil, err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createUserGroupWithRole(ctx context.Context, roleName string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, err := s.generateAUserGroupPayloadWithRoleName(ctx, "valid", roleName)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "s.generateAUserGroupPayloadWithRoleName")
	}

	createUserGroupResp, err := CreateUserGroup(ctx, s.BobDBTrace, s.UserMgmtConn, req, nil)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	ctx, err = s.generateACreateStaffProfile(ctx, "full field valid", "valid locations")
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("s.generateACreateStaffProfile: %w", err)
	}

	createStaffRequest := stepState.Request.(*pb.CreateStaffRequest)
	createStaffRequest.Staff.UserGroupIds = []string{createUserGroupResp.GetUserGroupId()}
	resp, err := pb.NewStaffServiceClient(s.UserMgmtConn).CreateStaff(contextWithToken(ctx), createStaffRequest)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "NewStaffServiceClient.CreateStaff")
	}

	stepState.ExistedUserGroupID = createUserGroupResp.GetUserGroupId()
	// assign user id for checking legacy user group later steps
	stepState.UserIDs = []string{resp.Staff.StaffId}
	stepState.Users = []*entity.LegacyUser{{
		ID:       database.Text(resp.Staff.StaffId),
		UserName: database.Text(resp.Staff.Username),
		Email:    database.Text(resp.Staff.Email),
	}}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) usersOfThatUserGroupHaveAUserGroupv1(ctx context.Context, legacyUserGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// check user group in users table
	query := `SELECT user_group FROM users WHERE user_id = ANY($1) AND deleted_at IS NULL`
	rows, err := s.BobDBTrace.Query(ctx, query, database.TextArray(stepState.UserIDs))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		userGroup := ""
		err := rows.Scan(&userGroup)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if legacyUserGroup != userGroup {
			return StepStateToContext(ctx, stepState), fmt.Errorf("after update user group legacy user group of users is expected %s but got %s", legacyUserGroup, userGroup)
		}
	}

	// check legacy user group
	query = `SELECT group_id FROM users_groups WHERE user_id = ANY($1) AND is_origin = TRUE AND status = $2`
	rows, err = s.BobDBTrace.Query(ctx, query, database.TextArray(stepState.UserIDs), entity.UserGroupStatusActive)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if rows.Err() != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		userGroup := ""
		err := rows.Scan(&userGroup)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if userGroup != legacyUserGroup {
			return StepStateToContext(ctx, stepState), fmt.Errorf("after update user group legacy user group of users is expected %s but got %s", legacyUserGroup, userGroup)
		}
	}
	return StepStateToContext(ctx, stepState), nil
}

func generateUpdateUserGroupRequest(ctx context.Context, bobDB database.QueryExecer, userGroupID string) (*pb.UpdateUserGroupRequest, error) {
	return generateUpdateUserGroupRequestWithRoleName(ctx, bobDB, userGroupID, constant.RoleTeacher)
}

func generateUpdateUserGroupRequestWithRoleName(ctx context.Context, bobDB database.QueryExecer, userGroupID string, roleName string) (*pb.UpdateUserGroupRequest, error) {
	req := &pb.UpdateUserGroupRequest{
		UserGroupId:   userGroupID,
		UserGroupName: fmt.Sprintf("updated_usergroup-%s", userGroupID),
	}

	roleNamesList := strings.Split(roleName, ", ")
	roleWithLocations, err := generateRoleWithLocationWithRoleName(ctx, bobDB, len(roleNamesList), roleNamesList)
	if err != nil {
		return nil, err
	}

	req.RoleWithLocations = roleWithLocations

	return req, nil
}

func generateRoleWithLocationWithRoleName(ctx context.Context, bobDB database.QueryExecer, limitRole int, roleNamesList []string) ([]*pb.RoleWithLocations, error) {
	query := `SELECT role_id FROM role WHERE role_name = ANY($1) and deleted_at IS NULL ORDER BY role_id DESC LIMIT $2`
	rows, err := bobDB.Query(ctx, query, roleNamesList, limitRole)
	if err != nil {
		return nil, fmt.Errorf("generateRoleWithLocation: find Role IDs: %w", err)
	}
	defer rows.Close()

	roleWithLocations := []*pb.RoleWithLocations{}
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roleWithLocations = append(
			roleWithLocations,
			&pb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: []string{constants.ManabieOrgLocation},
			},
		)
	}

	return roleWithLocations, nil
}

func (s *suite) userCanNotUpdateUserGroupAndReceiveStatusCodeError(ctx context.Context, signedUser, expectedCode string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr == nil {
		return ctx, errors.New("expected response has err but actual is nil")
	}

	stt, ok := status.FromError(stepState.ResponseErr)
	if !ok {
		return ctx, fmt.Errorf("returned error is not status.Status, err: %s", stepState.ResponseErr.Error())
	}

	if stt.Code().String() != expectedCode {
		return ctx, fmt.Errorf("expecting %s, got %s status code, message: %s", expectedCode, stt.Code().String(), stt.Message())
	}
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateUserGroupWithInvalidArgument(ctx context.Context, signedUser, invalidDataType string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, err := generateUpdateUserGroupRequest(ctx, s.BobDBTrace, stepState.ExistedUserGroupID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch invalidDataType {
	case "user group id empty":
		req.UserGroupId = ""
	case "user group is not existed":
		req.UserGroupId = "non-existed-user-group-id"
	case "missing user group name":
		req.UserGroupName = ""
	case "role id empty":
		req.RoleWithLocations[0].RoleId = ""
	case "location id empty":
		req.RoleWithLocations[0].LocationIds[0] = ""
	case "role is not existed":
		req.RoleWithLocations[0].RoleId = "non-existed-role-id"
	case "location is not existed":
		req.RoleWithLocations[0].LocationIds[0] = "non-existed-location-id"
	case "role missing location":
		req.RoleWithLocations[0].LocationIds = nil
	}

	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).UpdateUserGroup(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userUpdateUserGroupWithoutArgument(ctx context.Context, signedUser, withoutArgument string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	req, err := generateUpdateUserGroupRequest(ctx, s.BobDBTrace, stepState.ExistedUserGroupID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	if withoutArgument == "role with location" {
		req.RoleWithLocations = []*pb.RoleWithLocations{}
	}
	stepState.Request = req
	stepState.Response, stepState.ResponseErr = pb.NewUserGroupMgmtServiceClient(s.UserMgmtConn).UpdateUserGroup(ctx, req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) validUpdatedUserGroup(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)

	query := `
		SELECT
			user_group.user_group_name,
			granted_role.granted_role_id,
			granted_role.role_id,
			granted_role.deleted_at,
			granted_role_access_path.location_id,
			granted_role_access_path.deleted_at
		
		FROM
			user_group
		
		LEFT JOIN granted_role ON
			granted_role.user_group_id = user_group.user_group_id
		
		LEFT JOIN granted_role_access_path ON
			granted_role_access_path.granted_role_id = granted_role.granted_role_id

		WHERE
			user_group.user_group_id = $1
			AND
			user_group.deleted_at IS NULL
	`
	userGroupID := stepState.Request.(*pb.UpdateUserGroupRequest).UserGroupId
	rows, err := s.BobDBTrace.Query(ctx, query, userGroupID)
	if err != nil {
		return fmt.Errorf("BobDBTrace.Query: Find Role IDs: %w", err)
	}
	defer rows.Close()

	userGroupNames := make(map[string]struct{})
	grantedRoles := []*entity.GrantedRole{}
	mapGrantedRoleWithAccessPath := make(map[string][]*entity.GrantedRoleAccessPath)
	var (
		userGroupName                  string
		grantedRoleID                  string
		grantedRoleRoleID              string
		grantedRoleDeletedAt           *time.Time
		grantedRolAccessPathLocationID string
		grantedRolAccessPathDeletedAt  *time.Time
	)
	for rows.Next() {
		if err := rows.Scan(&userGroupName, &grantedRoleID, &grantedRoleRoleID, &grantedRoleDeletedAt, &grantedRolAccessPathLocationID, &grantedRolAccessPathDeletedAt); err != nil {
			return err
		}

		userGroupNames[userGroupName] = struct{}{}
		grantedRole := &entity.GrantedRole{
			GrantedRoleID: database.Text(grantedRoleID),
			UserGroupID:   database.Text(userGroupID),
			RoleID:        database.Text(grantedRoleRoleID),
		}
		if grantedRoleDeletedAt != nil {
			_ = grantedRole.DeletedAt.Set(grantedRoleDeletedAt)
		}
		grantedRoles = append(grantedRoles, grantedRole)

		grantedRoleAccessPath := &entity.GrantedRoleAccessPath{
			GrantedRoleID: database.Text(grantedRoleID),
			LocationID:    database.Text(grantedRolAccessPathLocationID),
		}
		if grantedRolAccessPathDeletedAt != nil {
			_ = grantedRoleAccessPath.DeletedAt.Set(grantedRolAccessPathDeletedAt)
		}
		mapGrantedRoleWithAccessPath[grantedRoleID] = append(mapGrantedRoleWithAccessPath[grantedRoleID], grantedRoleAccessPath)
	}

	// assertion nested factors
	req := stepState.Request.(*pb.UpdateUserGroupRequest)
	if _, ok := userGroupNames[req.GetUserGroupName()]; !ok {
		return fmt.Errorf("expected userGroupName %s but actual %s", req.GetUserGroupName(), userGroupNames)
	}

	mapRoleWithLocation := mapRoleWithLocation(req.RoleWithLocations)
	for _, role := range grantedRoles {
		// check soft deleted granted role and granted role access path
		if _, ok := mapRoleWithLocation[role.RoleID.String]; !ok {
			if role.DeletedAt.Status != pgtype.Present {
				return fmt.Errorf("expected grantedRole.DeletedAt not null")
			}

			grantedRoleAccessPaths := mapGrantedRoleWithAccessPath[role.GrantedRoleID.String]
			for _, grantedRoleAccessPath := range grantedRoleAccessPaths {
				if grantedRoleAccessPath.DeletedAt.Status != pgtype.Present {
					return fmt.Errorf("expected grantedRoleAccessPath.DeletedAt not null")
				}
			}
		} else { // check granted role and granted role access path inserted
			if role.DeletedAt.Status != pgtype.Undefined {
				return fmt.Errorf("expected grantedRole.DeletedAt null")
			}

			grantedRoleAccessPaths := mapGrantedRoleWithAccessPath[role.GrantedRoleID.String]
			for _, grantedRoleAccessPath := range grantedRoleAccessPaths {
				if grantedRoleAccessPath.DeletedAt.Status != pgtype.Undefined {
					return fmt.Errorf("expected grantedRoleAccessPath.DeletedAt null")
				}
			}
		}
	}

	return nil
}

func mapRoleWithLocation(req []*pb.RoleWithLocations) map[string][]string {
	roleWithLocation := map[string][]string{}
	for _, val := range req {
		roleWithLocation[val.RoleId] = val.LocationIds
	}

	return roleWithLocation
}
