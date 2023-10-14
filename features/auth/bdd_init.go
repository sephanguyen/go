package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/usermgmt"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/try"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	retryTimes = 5
)

var mapOrgAndAdminID = map[int]string{
	constants.ManabieSchool: "bdd_admin-manabie",
}

func InitRootAccount(ctx context.Context, shamirConn *grpc.ClientConn, firebaseAddr, jwtApplicant string) (map[int]common.AuthInfo, error) {
	err := try.Do(func(attempt int) (bool, error) {
		if shamirConn.GetState() == connectivity.Ready {
			return false, nil
		}

		if attempt < retryTimes {
			time.Sleep(time.Second)
			return true, fmt.Errorf("the shamir service is not READY")
		}

		return false, fmt.Errorf("the shamir service is not READY")
	})
	if err != nil {
		return nil, err
	}
	mapOrgDefaultAdmin := make(map[int]common.AuthInfo)
	for orgID, userID := range mapOrgAndAdminID {
		authInfo, err := usermgmt.GenerateFakeAuthInfo(ctx, shamirConn, firebaseAddr, jwtApplicant, userID, constant.UserGroupSchoolAdmin)
		if err != nil {
			return nil, err
		}
		mapOrgDefaultAdmin[orgID] = authInfo
	}

	return mapOrgDefaultAdmin, nil
}

func InitUser(ctx context.Context, db *pgxpool.Pool, jwtApplicant string) (map[int]common.MapRoleAndAuthInfo, error) {
	mapOrgUser, err := InitStaff(ctx, db, jwtApplicant)
	if err != nil {
		return nil, err
	}

	return mapOrgUser, nil
}

func InitStaff(ctx context.Context, db *pgxpool.Pool, jwtApplicant string) (map[int]common.MapRoleAndAuthInfo, error) {
	mapOrgStaff := make(map[int]common.MapRoleAndAuthInfo)
	for orgID, grantedPermissions := range orgAndGrantedPermission() {
		validContext := common.ValidContext(ctx, orgID, rootAccount[orgID].UserID, rootAccount[orgID].Token)
		mapRoleAndAuthInfo := make(common.MapRoleAndAuthInfo)

		for _, roleWithLocation := range grantedPermissions {
			resp, err := usermgmt.CreateStaff(validContext, db, connections.UserMgmtConn, nil, []usermgmt.RoleWithLocation{roleWithLocation}, roleWithLocation.LocationIDs)
			if err != nil {
				return nil, err
			}

			authInfo, err := usermgmt.GenerateFakeAuthInfo(ctx, connections.ShamirConn, firebaseAddr, jwtApplicant, resp.Staff.StaffId, constant.MapRoleWithLegacyUserGroup[roleWithLocation.RoleName])
			if err != nil {
				return nil, err
			}
			mapRoleAndAuthInfo[roleWithLocation.RoleName] = authInfo
		}
		mapOrgStaff[orgID] = mapRoleAndAuthInfo
	}

	return mapOrgStaff, nil
}

func orgAndGrantedPermission() map[int][]usermgmt.RoleWithLocation {
	mapOrgAndGrantedPermission := make(map[int][]usermgmt.RoleWithLocation)
	for orgID := range mapOrgAndAdminID {
		roleWithLocations := []usermgmt.RoleWithLocation{}
		for _, role := range constant.AllowListRoles {
			var locationIDs []string
			switch role {
			case constant.RoleSchoolAdmin, constant.RoleHQStaff:
				locationIDs = []string{usermgmt.GetOrgLocation(orgID)}
			// case constant.RoleCentreManager, constant.RoleCentreStaff, constant.RoleTeacher:
			// 	locationIDs = getChildrenLocation(orgID)
			default:
				continue
			}

			// jprep only have 2 roles: school admin and teacher so far
			// if orgID == constants.JPREPSchool {
			// 	if !golibs.InArrayString(role, []string{constant.RoleSchoolAdmin, constant.RoleTeacher}) {
			// 		continue
			// 	}
			// }
			// if orgID == constants.ManagaraBase || orgID == constants.ManagaraHighSchool {
			// 	if !golibs.InArrayString(role, []string{constant.RoleUsermgmtScheduleJob}) {
			// 		continue
			// 	}
			// }
			// if orgID == constants.KECDemo {
			// 	if !golibs.InArrayString(role, []string{constant.RoleSchoolAdmin}) {
			// 		continue
			// 	}
			// }
			roleWithLocations = append(roleWithLocations, usermgmt.RoleWithLocation{
				RoleName:    role,
				LocationIDs: locationIDs,
			})
		}

		mapOrgAndGrantedPermission[orgID] = roleWithLocations
	}

	return mapOrgAndGrantedPermission
}
