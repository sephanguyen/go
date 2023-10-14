package user_group

import (
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

const (
	isAllowRolesToLoginTeacherWeb = "User_Auth_AllowAllRolesToLoginTeacherWeb"
)

var (
	errNotAllowedCombinationRole = errors.New("the selected role combination is not allowed")
)

func userGroupPayloadToUserGroupEnt(payload *pb.CreateUserGroupRequest, resourcePath string, orgLocation *domain.Location) (*entity.UserGroupV2, error) {
	userGroup := &entity.UserGroupV2{}
	database.AllNullEntity(userGroup)
	if err := multierr.Combine(
		userGroup.UserGroupID.Set(idutil.ULIDNow()),
		userGroup.UserGroupName.Set(payload.UserGroupName),
		userGroup.ResourcePath.Set(resourcePath),
		userGroup.OrgLocationID.Set(orgLocation.LocationID),
		userGroup.IsSystem.Set(false),
	); err != nil {
		return nil, fmt.Errorf("set user group fail: %w", err)
	}

	return userGroup, nil
}

func roleWithLocationsPayloadToGrantedRole(payload *pb.RoleWithLocations, userGroupID string, resourcePath string) (*entity.GrantedRole, error) {
	grantedRole := &entity.GrantedRole{}
	database.AllNullEntity(grantedRole)
	if err := multierr.Combine(
		grantedRole.GrantedRoleID.Set(idutil.ULIDNow()),
		grantedRole.UserGroupID.Set(userGroupID),
		grantedRole.RoleID.Set(payload.RoleId),
		grantedRole.ResourcePath.Set(resourcePath),
	); err != nil {
		return nil, fmt.Errorf("set granted role fail: %w", err)
	}

	return grantedRole, nil
}

func userToTeacher(user *entity.LegacyUser) (*entity.Teacher, error) {
	schoolID, err := strconv.ParseInt(user.ResourcePath.String, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.ParseInt")
	}

	return &entity.Teacher{
		ID:           user.ID,
		ResourcePath: user.ResourcePath,
		LegacyUser:   *user,
		SchoolIDs:    database.Int4Array([]int32{int32(schoolID)}),
		DeletedAt:    pgtype.Timestamptz{Status: pgtype.Null},
	}, nil
}

func userToSchoolAdmin(user *entity.LegacyUser) (*entity.SchoolAdmin, error) {
	schoolID, err := strconv.ParseInt(user.ResourcePath.String, 10, 32)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.ParseInt")
	}

	return &entity.SchoolAdmin{
		SchoolAdminID: user.ID,
		ResourcePath:  user.ResourcePath,
		LegacyUser:    *user,
		SchoolID:      database.Int4(int32(schoolID)),
	}, nil
}

func newUserGroupEntity(userID, groupID, status string, isOrigin bool, resourcePath string) *entity.UserGroup {
	return &entity.UserGroup{
		UserID:       database.Text(userID),
		GroupID:      database.Text(groupID),
		IsOrigin:     database.Bool(isOrigin),
		Status:       database.Text(status),
		ResourcePath: database.Text(resourcePath),
	}
}

func combineRolesToLegacyUserGroup(roles []*entity.Role) (string, error) {
	if len(roles) == 0 {
		return "", nil
	}

	if len(roles) == 1 {
		return constant.MapRoleWithLegacyUserGroup[roles[0].RoleName.String], nil
	}

	currentRole := roles[0].RoleName.String
	for _, role := range roles[1:] {
		if !golibs.InArrayString(role.RoleName.String, constant.MapCombinationRole[currentRole]) {
			return "", errNotAllowedCombinationRole
		}
	}

	return constant.MapRoleWithLegacyUserGroup[currentRole], nil
}
