package user_group

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserGroupService) ValidateUserLogin(ctx context.Context, req *pb.ValidateUserLoginRequest) (*pb.ValidateUserLoginResponse, error) {
	isAllowRolesToLoginTeacherWeb, err := s.UserModifierService.UnleashClient.IsFeatureEnabled(isAllowRolesToLoginTeacherWeb, s.UserModifierService.Env)
	if err != nil {
		isAllowRolesToLoginTeacherWeb = false
	}
	// current platform permissions
	currentPlatformPermissions := PlatformPermissionMapper[req.Platform]

	if isAllowRolesToLoginTeacherWeb {
		currentPlatformPermissions = PlatformPermissionMapperV2[req.Platform]
	}

	// validate user id
	currentUserID := interceptors.UserIDFromContext(ctx)
	_, err = s.UserRepo.Get(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, errors.Wrap(err, "get user failed").Error())
	}

	// get list user groups with role by user id
	mapUserGroupRoles, err := s.UserGroupV2Repo.FindUserGroupAndRoleByUserID(ctx, s.DB, database.Text(currentUserID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, errors.Wrap(err, "find user group and roles failed").Error())
	}

	if len(mapUserGroupRoles) == 0 {
		return &pb.ValidateUserLoginResponse{Allowable: false}, nil
	}

	// check existed at least one of user groups have role to access this platform with new flow
	allowable := false
	for _, roles := range mapUserGroupRoles {
		for _, role := range roles {
			if _, ok := currentPlatformPermissions[role.RoleName.String]; ok {
				allowable = true
				break
			}
		}
	}

	return &pb.ValidateUserLoginResponse{Allowable: allowable}, nil
}
