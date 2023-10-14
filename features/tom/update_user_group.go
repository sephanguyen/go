package tom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	constants_lib "github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"
)

func (s *suite) updateUserGroupWithRoleNamesAndLocations(ctx context.Context, roleName string, locationType string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	manabieRp := strconv.Itoa(constants_lib.ManabieSchool)
	ctx = contextWithResourcePath(ctx, manabieRp)
	roleNames := []string{roleName}

	locationID, _, err := s.CommonSuite.CreateLocationWithDB(ctx, manabieRp, locationType, constants_lib.ManabieOrgLocation, ManabieOrgLocationType)
	if err != nil {
		return ctx, err
	}
	grantedLocations := []string{locationID}
	s.CommonSuite.LocationIDs = grantedLocations

	stmt := "SELECT role_id FROM role WHERE deleted_at IS NULL AND role_name = ANY($1) LIMIT $2"
	rows, err := s.CommonSuite.BobDBTrace.Query(ctx, stmt, roleNames, len(roleNames))
	if err != nil {
		return ctx, err
	}
	defer rows.Close()

	var roleIDs []string
	for rows.Next() {
		roleID := ""
		if err := rows.Scan(&roleID); err != nil {
			return ctx, fmt.Errorf("rows.Err: %w", err)
		}
		roleIDs = append(roleIDs, roleID)
	}
	if err := rows.Err(); err != nil {
		return ctx, fmt.Errorf("rows.Err: %w", err)
	}

	for _, userGroupID := range s.userGroupIDs {
		req := &upb.UpdateUserGroupRequest{
			UserGroupName: fmt.Sprintf("user-group_%s", idutil.ULIDNow()),
			UserGroupId:   userGroupID,
		}

		for _, roleID := range roleIDs {
			req.RoleWithLocations = append(
				req.RoleWithLocations,
				&upb.RoleWithLocations{
					RoleId:      roleID,
					LocationIds: grantedLocations,
				},
			)
		}

		_, err = upb.NewUserGroupMgmtServiceClient(s.CommonSuite.UserMgmtConn).UpdateUserGroup(contextWithToken(ctx2, s.schoolAdminToken), req)
		if err != nil {
			return ctx, fmt.Errorf("CreateUserGroup: %w", err)
		}
	}

	return ctx, nil
}
