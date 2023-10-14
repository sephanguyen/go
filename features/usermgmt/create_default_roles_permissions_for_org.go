package usermgmt

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb_ms "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"

	"github.com/pkg/errors"
)

func createNewOrganizationData() *pb_ms.CreateOrganizationRequest {
	return &pb_ms.CreateOrganizationRequest{
		Organization: &pb_ms.Organization{
			OrganizationId:   newID(),
			TenantId:         newID(),
			OrganizationName: "organization name test",
			DomainName:       strings.ToLower("domain-test" + idutil.ULIDNow()),
			LogoUrl:          "logo-url",
			CountryCode:      cpb.Country_COUNTRY_JP,
		},
	}
}

func (s *suite) genOrganizationInfo(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.Request = createNewOrganizationData()
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) signedInAndCreateOrg(ctx context.Context, group string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	ctx, err := s.signedAsAccount(ctx, group)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrapf(err, "s.signedAsAccount: %s", group)
	}

	req := stepState.Request.(*pb_ms.CreateOrganizationRequest)
	resp, err := pb_ms.NewOrganizationServiceClient(s.MasterMgmtConn).CreateOrganization(contextWithToken(ctx), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), errors.Wrap(err, "NewOrganizationServiceClient.CreateOrganization")
	}
	stepState.Response = resp

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) roleAndPermissionMustBeExistedInDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	org := stepState.Response.(*pb_ms.CreateOrganizationResponse).GetOrganization()

	var mapRoleWithPermission map[string](map[string]struct{})
	var err error

	retryAmount := 5
	sleepRetry := time.Duration(10)

	err = try.Do(func(attempt int) (bool, error) {
		mapRoleWithPermission, err = getRoleWithPermissionByOrgID(ctx, s.BobDB, org.GetOrganizationId())
		if err == nil {
			if err := checkExistCorrectRoleWithPermission(org.GetDomainName(), mapRoleWithPermission); err == nil {
				return false, nil
			}
		}
		retry := attempt < retryAmount
		if retry {
			time.Sleep(sleepRetry * time.Second)
			return true, err
		}
		return false, err
	})
	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("can not retrieve the list role has same amount with the defined")
	}
	return StepStateToContext(ctx, stepState), nil
}

func getRoleWithPermissionByOrgID(ctx context.Context, db database.Ext, orgID string) (map[string](map[string]struct{}), error) {
	query := `
	    SELECT r.role_name, p.permission_name
	    
	    FROM role r
	    
	    INNER JOIN permission_role pr
	      ON r.role_id = pr.role_id AND
	         r.resource_path = pr.resource_path
	    
	    INNER JOIN permission p
	      ON pr.permission_id = p.permission_id AND
	         pr.resource_path = p.resource_path
	    
	    WHERE
	      r.resource_path = (
	        SELECT resource_path
	        FROM organizations
	        WHERE organization_id = $1
	      )
	`
	rows, err := db.Query(ctx, query, orgID)
	if err != nil {
		return nil, errors.Wrapf(err, "s.BobDB.Query: %s", orgID)
	}
	defer rows.Close()

	roleName := ""
	permissionName := ""
	// find and role and map existed permission to role
	mapRoleWithPermission := map[string](map[string]struct{}){}
	for rows.Next() {
		if err := rows.Scan(&roleName, &permissionName); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}

		_, ok := mapRoleWithPermission[roleName]
		if !ok {
			mapRoleWithPermission[roleName] = map[string]struct{}{}
		}

		mapRoleWithPermission[roleName][permissionName] = struct{}{}
	}

	return mapRoleWithPermission, nil
}

func checkExistCorrectRoleWithPermission(orgID string, mapRoleWithPermission map[string](map[string]struct{})) error {
	for role, permissions := range port.RoleWithPermissionForOrg {
		if _, ok := mapRoleWithPermission[role]; !ok {
			return fmt.Errorf("`%s` role was not created for the %s org", role, orgID)
		}

		for _, permission := range permissions {
			if _, ok := mapRoleWithPermission[role][permission]; !ok {
				return fmt.Errorf("`%s` permission was not created for %s role of the %s org", permission, role, orgID)
			}
		}
	}
	return nil
}
