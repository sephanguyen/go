package user_group

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func (s *UserGroupService) CreateUserGroup(ctx context.Context, req *pb.CreateUserGroupRequest) (*pb.CreateUserGroupResponse, error) {
	// check resourcePath
	resourcePath, err := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	if err := validationCreateUserGroup(req); err != nil {
		return nil, err
	}

	if err := s.validRoleWithLocations(ctx, req.RoleWithLocations); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "validRoleWithLocations").Error())
	}

	userGroup, err := s.HandleCreateUserGroup(ctx, req, resourcePath)
	if err != nil {
		return nil, fmt.Errorf("s.HandleCreateUserGroup: %w", err)
	}

	upsertUserGroupEvt := s.toEventUpsertUserGroup(userGroup.UserGroupID.String)
	if err := s.publishUpsertUserGroupEvent(ctx, constants.SubjectUpsertUserGroup, upsertUserGroupEvt); err != nil {
		return nil, fmt.Errorf("s.publishUpsertUserGroupEvent: %w", err)
	}

	return &pb.CreateUserGroupResponse{UserGroupId: userGroup.UserGroupID.String}, nil
}

func (s *UserGroupService) HandleCreateUserGroup(ctx context.Context, req *pb.CreateUserGroupRequest, resourcePath int) (*entity.UserGroupV2, error) {
	var userGroup *entity.UserGroupV2
	var err error

	if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		orgLocation, err := s.LocationRepo.GetLocationOrg(ctx, tx, fmt.Sprint(resourcePath))
		if err != nil {
			return fmt.Errorf("s.LocationRepo.GetLocationOrg: %w", err)
		}

		// convert payload to entity
		if userGroup, err = userGroupPayloadToUserGroupEnt(req, fmt.Sprint(resourcePath), orgLocation); err != nil {
			return fmt.Errorf("s.UserGroupPayloadToUserGroupEnts: %w", err)
		}

		// create user group first
		if err = s.UserGroupV2Repo.Create(ctx, tx, userGroup); err != nil {
			return fmt.Errorf("s.UserGroupV2Repo.Create: %w", err)
		}

		var grantedRole *entity.GrantedRole
		for _, roleWithLocations := range req.RoleWithLocations {
			// convert payload to entity
			if grantedRole, err = roleWithLocationsPayloadToGrantedRole(roleWithLocations, userGroup.UserGroupID.String, fmt.Sprint(resourcePath)); err != nil {
				return fmt.Errorf("s.RoleWithLocationsPayloadToGrantedRole: %w", err)
			}
			// create granted_role
			if err = s.GrantedRoleRepo.Create(ctx, tx, grantedRole); err != nil {
				return fmt.Errorf("s.UserGroupV2Repo.Create: %w", err)
			}

			// link granted_role to access path(by location ids)
			if err = s.GrantedRoleRepo.LinkGrantedRoleToAccessPath(ctx, tx, grantedRole, roleWithLocations.LocationIds); err != nil {
				return fmt.Errorf("s.GrantedRoleRepo.LinkGrantedRoleToAccessPath: %w", err)
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("database.ExecInTx: %w", err)
	}

	return userGroup, nil
}

func validationCreateUserGroup(req *pb.CreateUserGroupRequest) error {
	// check user group name is exist in params
	if req.UserGroupName == "" {
		return status.Error(codes.InvalidArgument, "user group name is empty")
	}

	// check role id & location id is exist in params
	if err := validateRoleWithLocationsParams(req.RoleWithLocations); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	return nil
}

func sumUpRoleIDAndLocationIDsTogether(roleWithLocations []*pb.RoleWithLocations) ([]string, []string) {
	// merge all factors and check in one time and using map for removing duplicate ids fastly
	mapRoleIDs := make(map[string]struct{})
	roleIDs := make([]string, 0)
	mapLocationIDs := make(map[string]struct{})
	locationIDs := make([]string, 0)

	for _, roleWithLocation := range roleWithLocations {
		if _, ok := mapRoleIDs[roleWithLocation.RoleId]; !ok {
			mapRoleIDs[roleWithLocation.RoleId] = struct{}{}
			roleIDs = append(roleIDs, roleWithLocation.RoleId)
		}

		for _, locationID := range roleWithLocation.LocationIds {
			if _, ok := mapLocationIDs[locationID]; !ok {
				mapLocationIDs[locationID] = struct{}{}
				locationIDs = append(locationIDs, locationID)
			}
		}
	}

	return roleIDs, locationIDs
}

func (s *UserGroupService) validRoleWithLocations(ctx context.Context, roleWithLocations []*pb.RoleWithLocations) error {
	if len(roleWithLocations) == 0 {
		return nil
	}
	roleIDs, locationIDs := sumUpRoleIDAndLocationIDsTogether(roleWithLocations)
	// check role id is exist in DB
	var ids pgtype.TextArray
	if err := ids.Set(roleIDs); err != nil {
		return fmt.Errorf("GetRoles combine locationId fail")
	}
	roles, err := s.RoleRepo.GetRolesByRoleIDs(ctx, s.DB, ids)
	if err != nil {
		return fmt.Errorf("error when get roles by role ids")
	}
	if len(roles) != len(roleIDs) {
		return fmt.Errorf("role ids are invalid")
	}

	isAllowCombinationMultipleRoles, err := s.UnleashClient.IsFeatureEnabled(constant.FeatureToggleAllowCombinationMultipleRoles, s.Env)
	if err != nil {
		isAllowCombinationMultipleRoles = false
	}

	if !isAllowCombinationMultipleRoles {
		if _, err = combineRolesToLegacyUserGroup(roles); err != nil {
			return err
		}
	}

	// check all location ids is exist in DB
	if _, err := s.UserModifierService.GetLocations(ctx, locationIDs); err != nil {
		return fmt.Errorf("location ids are invalid")
	}

	return nil
}

func (s *UserGroupService) toEventUpsertUserGroup(userGroupID string) *pb.EvtUpsertUserGroup {
	return &pb.EvtUpsertUserGroup{
		UserGroupId: userGroupID,
	}
}

func (s *UserGroupService) publishUpsertUserGroupEvent(ctx context.Context, subject string, event *pb.EvtUpsertUserGroup) error {
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event %s error, %w", subject, err)
	}
	_, err = s.JSM.TracedPublish(ctx, "publishUpsertUserGroupEvent", subject, data)
	if err != nil {
		return fmt.Errorf("publishUpsertUserGroupEvent with %s: s.JSM.Publish failed: %w", subject, err)
	}

	return nil
}
