package usermgmt

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	location_repo "github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"google.golang.org/grpc"
)

func (s *suite) generateAUserGroupPayloadWithRoleName(ctx context.Context, validity string, roleName string) (*pb.CreateUserGroupRequest, error) {
	req := &pb.CreateUserGroupRequest{
		UserGroupName: "UserGroupName",
	}

	roleNamesList := strings.Split(roleName, ", ")
	query := `SELECT role_id FROM role WHERE role_name = ANY($1) and deleted_at IS NULL ORDER BY created_at ASC LIMIT $2`
	rows, err := s.BobDBTrace.Query(ctx, query, roleNamesList, len(roleNamesList))
	if err != nil {
		return nil, fmt.Errorf("BobDBTrace.Query: Find Role IDs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		// get role ids and init payload
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}

		req.RoleWithLocations = append(
			req.RoleWithLocations,
			&pb.RoleWithLocations{
				RoleId:      roleID,
				LocationIds: []string{constants.ManabieOrgLocation},
			},
		)
	}

	// fix payload with given style
	strangeID := newID()
	switch validity {
	case "missing name":
		req.UserGroupName = ""

	case "missing role_ids":
		req.RoleWithLocations[0].LocationIds = []string{}

	case "invalid location_id":
		req.RoleWithLocations[0].LocationIds = []string{strangeID}

	case "invalid role_id":
		req.RoleWithLocations[0].RoleId = strangeID
	}

	return req, nil
}

func (s *suite) signedInAndCreateUserGroupWithValidityPayload(ctx context.Context, validity string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	createUserGroupReq, err := s.generateAUserGroupPayloadWithRoleName(ctx, validity, constant.RoleSchoolAdmin)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("generateAUserGroupPayload: %w", err)
	}

	resp, err := CreateUserGroup(ctx, s.BobDBTrace, s.UserMgmtConn, createUserGroupReq, nil)
	if err != nil {
		stepState.ResponseErr = err
		return StepStateToContext(ctx, stepState), nil
	}

	stepState.Request = createUserGroupReq
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGroupMustBeExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	query := `
		SELECT
			user_group.user_group_name,
			user_group.org_location_id,
			user_group.is_system,
			granted_role.granted_role_id,
			granted_role_access_path.location_id,
			granted_role.role_id
		
		FROM
			user_group
		
		LEFT JOIN granted_role ON
			granted_role.user_group_id = user_group.user_group_id
				AND
			granted_role.deleted_at IS NULL
		
		LEFT JOIN granted_role_access_path ON
			granted_role_access_path.granted_role_id = granted_role.granted_role_id
				AND
			granted_role_access_path.deleted_at IS NULL
		
		WHERE
			user_group.user_group_id = $1
				AND
			user_group.deleted_at IS NULL
	`
	userGroupID := stepState.Response.(*pb.CreateUserGroupResponse).UserGroupId
	rows, err := s.BobDBTrace.Query(ctx, query, userGroupID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("BobDBTrace.Query: Find Role IDs: %w", err)
	}
	defer rows.Close()

	// store found results for assertion later
	// https://stackoverflow.com/a/47544821
	userGroupNames := make(map[string]struct{})
	ugOrgLocationIDs := make(map[string]struct{})
	grantedRoleIds := make(map[string]struct{})
	locationIds := make(map[string]struct{})
	roleIds := make(map[string]struct{})

	userGroupName := ""
	ugOrgLocationID := ""
	grantedRoleID := ""
	locationID := ""
	roleID := ""
	isSystem := true

	for rows.Next() {
		if err := rows.Scan(&userGroupName, &ugOrgLocationID, &isSystem, &grantedRoleID, &locationID, &roleID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when scaning user group: %w", err)
		}

		// firstly we have to assert that user group was not created by system
		if isSystem {
			return StepStateToContext(ctx, stepState), fmt.Errorf("found user group `%s` was created by system", userGroupName)
		}
		userGroupNames[userGroupName] = struct{}{}
		ugOrgLocationIDs[ugOrgLocationID] = struct{}{}
		grantedRoleIds[grantedRoleID] = struct{}{}
		locationIds[locationID] = struct{}{}
		roleIds[roleID] = struct{}{}
	}

	// assertion nested factors
	req := stepState.Request.(*pb.CreateUserGroupRequest)
	if _, ok := userGroupNames[req.GetUserGroupName()]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroupName %s not found in DB", req.GetUserGroupName())
	}

	for _, roleWithLocations := range req.RoleWithLocations {
		if _, ok := roleIds[roleWithLocations.RoleId]; !ok {
			return StepStateToContext(ctx, stepState), fmt.Errorf("role of userGroup was not inserted correctly")
		}

		for _, locationID := range roleWithLocations.LocationIds {
			if _, ok := locationIds[locationID]; !ok {
				return StepStateToContext(ctx, stepState), fmt.Errorf("location of userGroup was not inserted correctly")
			}
		}
	}

	resourcePath := golibs.ResourcePathFromCtx(ctx)
	locationOrg, err := (&location_repo.LocationRepo{}).GetLocationOrg(ctx, s.BobDB, fmt.Sprint(resourcePath))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot find location org")
	}
	if _, ok := ugOrgLocationIDs[locationOrg.LocationID]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected org location but now found")
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGroupCreatedSuccessfully(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	resp, ok := stepState.Response.(*pb.CreateUserGroupResponse)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroup was not created: %v", stepState.ResponseErr)
	}
	if resp == nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("userGroup was not created: %v", stepState.ResponseErr)
	}

	return StepStateToContext(ctx, stepState), nil
}

func createUserGroupRequest(ctx context.Context, db database.QueryExecer, roleWithLocations []RoleWithLocation) (*pb.CreateUserGroupRequest, error) {
	req := &pb.CreateUserGroupRequest{
		UserGroupName: fmt.Sprintf("user_group+%s", idutil.ULIDNow()),
	}
	// roleWithLocationReq := &pb.RoleWithLocations{}
	for _, roleWithLocation := range roleWithLocations {
		if len(roleWithLocation.LocationIDs) == 0 {
			return nil, fmt.Errorf("granted location empty")
		}
		role, err := (&repository.RoleRepo{}).GetByName(ctx, db, database.Text(roleWithLocation.RoleName))
		if err != nil {
			return nil, err
		}
		req.RoleWithLocations = append(req.RoleWithLocations, &pb.RoleWithLocations{
			RoleId:      role.RoleID.String,
			LocationIds: roleWithLocation.LocationIDs,
		})
	}

	return req, nil
}

func createUserGroup(ctx context.Context, userConnection *grpc.ClientConn, req *pb.CreateUserGroupRequest) (*pb.CreateUserGroupResponse, error) {
	return pb.NewUserGroupMgmtServiceClient(userConnection).CreateUserGroup(ctx, req)
}
