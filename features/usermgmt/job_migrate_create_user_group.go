package usermgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/cmd/server/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/auth"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/configurations"

	"github.com/pkg/errors"
)

var queryUserGroupAndGrantedPermission = `
	SELECT
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
			user_group.user_group_name = $1
				AND
			user_group.deleted_at IS NULL
				AND
			user_group.resource_path = $2
		ORDER BY user_group.created_at DESC
	`

func (s *suite) someRolesAndLocationsToCreateUserGroup(ctx context.Context, role string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) systemRunJobToMigrateCreateUserGroupWithUserGroupNameRolesAndOrganization(ctx context.Context, userGroupName, role, organization string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var (
		organizationID string
		locationID     string
	)

	if organization == ManabieSchool {
		organizationID = fmt.Sprint(constants.ManabieSchool)
		locationID = fmt.Sprint(constants.ManabieOrgLocation)
	}

	usermgmt.RunMigrateCreateUserGroup(ctx, &configurations.Config{
		Common:     s.Cfg.Common,
		PostgresV2: s.Cfg.PostgresV2,
	}, userGroupName, role, locationID, organizationID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) userGroupCreateSuccessfullyWithUserGroupNameRolesAndOrganization(ctx context.Context, userGroupName, role, organization string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		organizationID string
		locationID     string
	)

	if organization == ManabieSchool {
		organizationID = fmt.Sprint(constants.ManabieSchool)
		locationID = fmt.Sprint(constants.ManabieOrgLocation)
	}

	ctx = auth.InjectFakeJwtToken(ctx, organizationID)
	existedRole, err := (&repository.RoleRepo{}).GetByName(ctx, s.BobDBTrace, database.Text(role))
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("get role by name %w", err)
	}

	locationOrg, err := (&repo.LocationRepo{}).GetLocationOrg(ctx, s.BobDBTrace, organizationID)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "get location org failed")
	}

	rows, err := s.BobDBTrace.Query(ctx, queryUserGroupAndGrantedPermission, userGroupName, organizationID)
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("BobDBTrace.Query: Find Role IDs: %w", err)
	}
	defer rows.Close()

	// store found results for assertion later
	ugOrgLocationIDs := make(map[string]struct{})
	grantedRoleIds := make(map[string]struct{})
	locationIds := make(map[string]struct{})
	roleIds := make(map[string]struct{})

	for rows.Next() {
		var (
			ugOrgLocationID             string
			grantedRoleID               string
			grantedAccessPathLocationID string
			roleID                      string
			isSystem                    bool
		)

		if err := rows.Scan(&ugOrgLocationID, &isSystem, &grantedRoleID, &grantedAccessPathLocationID, &roleID); err != nil {
			return StepStateToContext(ctx, stepState), fmt.Errorf("error when scaning user group: %w", err)
		}

		// make sure user group (newly created) isn't system user_group
		if !isSystem {
			return StepStateToContext(ctx, stepState), fmt.Errorf("expect user group `%s` was created by system", userGroupName)
		}
		ugOrgLocationIDs[ugOrgLocationID] = struct{}{}
		grantedRoleIds[grantedRoleID] = struct{}{}
		locationIds[grantedAccessPathLocationID] = struct{}{}
		roleIds[roleID] = struct{}{}
	}

	if _, ok := roleIds[existedRole.RoleID.String]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("role of userGroup was not inserted correctly")
	}

	if _, ok := ugOrgLocationIDs[locationOrg.LocationID]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("org location of userGroup was not inserted correctly")
	}

	if _, ok := locationIds[locationID]; !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("location of userGroup was not inserted correctly")
	}

	return StepStateToContext(ctx, stepState), nil
}
