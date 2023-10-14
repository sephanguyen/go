package user_group

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserGroupService) UpdateUserGroup(ctx context.Context, req *pb.UpdateUserGroupRequest) (*pb.UpdateUserGroupResponse, error) {
	// check resourcePath
	resourcePath, err := strconv.Atoi(golibs.ResourcePathFromCtx(ctx))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	if err := validateUpdateUserGroupParams(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "validateUpdateUserGroupParams").Error())
	}

	if err := s.validRoleWithLocations(ctx, req.RoleWithLocations); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "validRoleWithLocations").Error())
	}

	existedUserGroup, err := s.UserGroupV2Repo.Find(ctx, s.DB, database.Text(req.UserGroupId))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if err := assignParameterToUpdate(req, existedUserGroup); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	existedGrantedRoles, err := s.GrantedRoleRepo.GetByUserGroup(ctx, s.DB, database.Text(existedUserGroup.UserGroupID.String))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	grantedRoleEntities, err := toGrantedRoleEntities(existedGrantedRoles, req.RoleWithLocations, existedUserGroup.UserGroupID.String, fmt.Sprint(resourcePath))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	grantedRoleAccessPathEntities, err := toGrantedRoleAccessPaths(existedGrantedRoles, req.RoleWithLocations, grantedRoleEntities, fmt.Sprint(resourcePath))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	grantedRoleIDsToRevoke := grantedRoleIDsToRevoke(existedGrantedRoles, req.RoleWithLocations)

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if err := s.UserGroupV2Repo.Update(ctx, tx, existedUserGroup); err != nil {
			return status.Error(codes.Internal, fmt.Errorf("UserGroupV2Repo.Update: %w", err).Error())
		}
		if len(grantedRoleIDsToRevoke) > 0 {
			if err := s.GrantedRoleRepo.SoftDelete(ctx, tx, database.TextArray(grantedRoleIDsToRevoke)); err != nil {
				return status.Error(codes.Internal, fmt.Errorf("UserGroupV2Repo.Update: %w", err).Error())
			}
		}
		if len(grantedRoleEntities) > 0 {
			if err := s.GrantedRoleRepo.Upsert(ctx, tx, grantedRoleEntities); err != nil {
				return status.Error(codes.Internal, fmt.Errorf("GrantedRoleRepo.Upsert: %w", err).Error())
			}
		}
		if len(grantedRoleAccessPathEntities) > 0 {
			if err := s.GrantedRoleAccessPathRepo.Upsert(ctx, tx, grantedRoleAccessPathEntities); err != nil {
				return status.Error(codes.Internal, fmt.Errorf("GrantedRoleAccessPathRepo.Upsert: %w", err).Error())
			}
		}
		roleIDs := make([]string, 0)
		for _, roleWithLocation := range req.GetRoleWithLocations() {
			roleIDs = append(roleIDs, roleWithLocation.GetRoleId())
		}

		roles, err := s.RoleRepo.GetRolesByRoleIDs(ctx, tx, database.TextArray(roleIDs))
		if err != nil {
			return errors.Wrap(err, "s.RoleRepo.GetRolesByRoleIDs")
		}

		if err := s.updateLegacyUserGroupOfUsers(ctx, tx, roles, req.UserGroupId); err != nil {
			return errors.Wrap(err, "s.updateLegacyUserGroupOfUsers")
		}

		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	upsertUserGroupEvt := s.toEventUpsertUserGroup(req.UserGroupId)
	if err := s.publishUpsertUserGroupEvent(ctx, constants.SubjectUpsertUserGroup, upsertUserGroupEvt); err != nil {
		return nil, fmt.Errorf("s.publishUpsertUserGroupEvent: %w", err)
	}

	return &pb.UpdateUserGroupResponse{Successful: true}, nil
}

func validateUpdateUserGroupParams(req *pb.UpdateUserGroupRequest) error {
	if req.UserGroupId == "" {
		return fmt.Errorf("userGroupID empty")
	}
	if req.UserGroupName == "" {
		return fmt.Errorf("userGroupName empty")
	}
	if err := validateRoleWithLocationsParams(req.RoleWithLocations); err != nil {
		return err
	}

	return nil
}

func validateRoleWithLocationsParams(roleWithLocations []*pb.RoleWithLocations) error {
	for _, roleWithLocation := range roleWithLocations {
		if roleWithLocation.RoleId == "" {
			return fmt.Errorf("roleID empty")
		}
		if len(roleWithLocation.LocationIds) == 0 {
			return fmt.Errorf("granted role missing location")
		}
		for _, id := range roleWithLocation.LocationIds {
			if id == "" {
				return fmt.Errorf("locationID empty")
			}
		}
	}

	return nil
}

func assignParameterToUpdate(req *pb.UpdateUserGroupRequest, userGroup *entity.UserGroupV2) error {
	if err := multierr.Combine(
		userGroup.UserGroupID.Set(req.UserGroupId),
		userGroup.UserGroupName.Set(req.UserGroupName),
	); err != nil {
		return errors.Wrap(err, "assignParameterToUpdate")
	}

	return nil
}

func grantedRoleIDsToRevoke(existedGrantedRoles []*entity.GrantedRole, roleWithLocations []*pb.RoleWithLocations) []string {
	// compare existed grantedRole with request
	grantedRoleIDsToRevoke := []string{}
	mapRoleWithLocation := mapRoleWithLocation(roleWithLocations)
	for _, eGrantedRole := range existedGrantedRoles {
		if _, ok := mapRoleWithLocation[eGrantedRole.RoleID.String]; !ok {
			grantedRoleIDsToRevoke = append(grantedRoleIDsToRevoke, eGrantedRole.GrantedRoleID.String)
		}
	}

	return grantedRoleIDsToRevoke
}

func toGrantedRoleEntities(existedGrantedRoles []*entity.GrantedRole, roleWithLocations []*pb.RoleWithLocations, userGroupID string, resourcePath string) ([]*entity.GrantedRole, error) {
	grantedRoles := []*entity.GrantedRole{}

	// compare request with existed grantedRole
	existedGrantedRoleIDs := make(map[string]bool)
	for _, grantedRole := range existedGrantedRoles {
		if grantedRole.DeletedAt.Status != pgtype.Present {
			existedGrantedRoleIDs[grantedRole.RoleID.String] = true
		}
	}

	for _, roleWithLocation := range roleWithLocations {
		// continue when grantedRole already existed
		if _, ok := existedGrantedRoleIDs[roleWithLocation.RoleId]; ok {
			continue
		}

		grantedRole := &entity.GrantedRole{}
		database.AllNullEntity(grantedRole)
		if err := multierr.Combine(
			grantedRole.GrantedRoleID.Set(idutil.ULIDNow()),
			grantedRole.UserGroupID.Set(userGroupID),
			grantedRole.RoleID.Set(roleWithLocation.RoleId),
			grantedRole.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, errors.Wrap(err, "toGrantedRoleEntities")
		}
		grantedRoles = append(grantedRoles, grantedRole)
	}

	return grantedRoles, nil
}

func toGrantedRoleAccessPaths(existedGrantedRoles []*entity.GrantedRole, roleWithLocations []*pb.RoleWithLocations, grantedRoleEntities []*entity.GrantedRole, resourcePath string) ([]*entity.GrantedRoleAccessPath, error) {
	grantedRoleAccessPaths := []*entity.GrantedRoleAccessPath{}
	mapRoleWithLocation := mapRoleWithLocation(roleWithLocations)

	for _, grantedRole := range existedGrantedRoles {
		// re-assign location when roleWithLocations in request already existed
		if locationIDs, ok := mapRoleWithLocation[grantedRole.RoleID.String]; ok {
			grantedRoleAccessPathEntities, err := toGrantedRoleAccessPathEnts(grantedRole.GrantedRoleID.String, resourcePath, locationIDs)
			if err != nil {
				return nil, err
			}

			grantedRoleAccessPaths = append(grantedRoleAccessPaths, grantedRoleAccessPathEntities...)
			continue
		}

		// soft delete grantedRoleAccessPath if grantedRole has been removed
		grantedRoleAccessPath := &entity.GrantedRoleAccessPath{}
		database.AllNullEntity(grantedRoleAccessPath)
		if err := grantedRoleAccessPath.GrantedRoleID.Set(grantedRole.GrantedRoleID); err != nil {
			return nil, err
		}

		grantedRoleAccessPaths = append(grantedRoleAccessPaths, grantedRoleAccessPath)
	}

	for _, grantedRole := range grantedRoleEntities {
		// skip incase grantedRole will be soft deleted
		if grantedRole.RoleID.String == "" {
			continue
		}

		locationIDs := mapRoleWithLocation[grantedRole.RoleID.String]
		grantedRoleAccessPathEntities, err := toGrantedRoleAccessPathEnts(grantedRole.GrantedRoleID.String, resourcePath, locationIDs)
		if err != nil {
			return nil, err
		}

		grantedRoleAccessPaths = append(grantedRoleAccessPaths, grantedRoleAccessPathEntities...)
	}

	return grantedRoleAccessPaths, nil
}

func mapRoleWithLocation(roleWithLocations []*pb.RoleWithLocations) map[string][]string {
	mapRoleWithLocation := make(map[string][]string)
	for _, roleWithLocation := range roleWithLocations {
		mapRoleWithLocation[roleWithLocation.RoleId] = roleWithLocation.LocationIds
	}

	return mapRoleWithLocation
}

func toGrantedRoleAccessPathEnts(grantedRoleID, resourcePath string, locationIDs []string) ([]*entity.GrantedRoleAccessPath, error) {
	grantedRoleAccessPaths := []*entity.GrantedRoleAccessPath{}
	for _, locationID := range locationIDs {
		grantedRoleAccessPath := &entity.GrantedRoleAccessPath{}
		database.AllNullEntity(grantedRoleAccessPath)
		if err := multierr.Combine(
			grantedRoleAccessPath.GrantedRoleID.Set(grantedRoleID),
			grantedRoleAccessPath.LocationID.Set(locationID),
			grantedRoleAccessPath.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, err
		}

		grantedRoleAccessPaths = append(grantedRoleAccessPaths, grantedRoleAccessPath)
	}

	return grantedRoleAccessPaths, nil
}

func (s *UserGroupService) updateLegacyUserGroupOfUsers(ctx context.Context, db database.QueryExecer, roles []*entity.Role, userGroupID string) error {
	hasRoleSchoolAdmin := false

	for _, role := range roles {
		if constant.MapRoleWithLegacyUserGroup[role.RoleName.String] == constant.UserGroupSchoolAdmin {
			hasRoleSchoolAdmin = true
			break
		}
	}

	users, err := s.UserRepo.GetUsersByUserGroupID(ctx, db, database.Text(userGroupID))
	if err != nil {
		return errors.Wrap(err, "s.UserRepo.GetUsersByUserGroupID")
	}

	if hasRoleSchoolAdmin {
		if err := s.switchUserGroupOfUser(ctx, db, users, constant.UserGroupSchoolAdmin); err != nil {
			return errors.Wrapf(err, "s.switchUserGroupOfUser: %s", constant.UserGroupSchoolAdmin)
		}
	} else {
		if err := s.switchUserGroupOfUser(ctx, db, users, constant.UserGroupTeacher); err != nil {
			return errors.Wrapf(err, "s.switchUserGroupOfUser: %s", constant.UserGroupTeacher)
		}
	}

	return nil
}

func (s *UserGroupService) switchUserGroupOfUser(ctx context.Context, db database.QueryExecer, users []*entity.LegacyUser, grantedUserGroup string) error {
	revokeUserGroup := ""
	listLegacyUserGroups := make([]*entity.UserGroup, 0)
	userIDs := make([]string, 0)
	for _, user := range users {
		userIDs = append(userIDs, user.ID.String)
	}

	switch grantedUserGroup {
	case constant.UserGroupTeacher:
		teachers := make([]*entity.Teacher, 0)
		for _, user := range users {
			listLegacyUserGroups = append(
				listLegacyUserGroups,
				newUserGroupEntity(
					user.ID.String,
					constant.UserGroupTeacher,
					entity.UserGroupStatusActive,
					true,
					user.ResourcePath.String,
				),
			)

			teacher, err := userToTeacher(user)
			if err != nil {
				return errors.Wrap(err, "userToTeacher")
			}
			teachers = append(teachers, teacher)
		}

		if err := s.UserModifierService.TeacherRepo.UpsertMultiple(ctx, db, teachers); err != nil {
			return errors.Wrap(err, "s.UserModifierService.TeacherRepo.UpsertMultiple")
		}

		if err := s.UserModifierService.SchoolAdminRepo.SoftDeleteMultiple(ctx, db, database.TextArray(userIDs)); err != nil {
			return errors.Wrap(err, "s.UserModifierService.SchoolAdminRepo.SoftDeleteMultiple")
		}

		revokeUserGroup = constant.UserGroupSchoolAdmin

	case constant.UserGroupSchoolAdmin:
		schoolAdmins := make([]*entity.SchoolAdmin, 0)
		for _, user := range users {
			listLegacyUserGroups = append(
				listLegacyUserGroups,
				newUserGroupEntity(
					user.ID.String,
					constant.UserGroupSchoolAdmin,
					entity.UserGroupStatusActive,
					true,
					user.ResourcePath.String,
				),
			)

			schoolAdmin, err := userToSchoolAdmin(user)
			if err != nil {
				return errors.Wrap(err, "userToSchoolAdmin")
			}
			schoolAdmins = append(schoolAdmins, schoolAdmin)
		}

		if err := s.UserModifierService.SchoolAdminRepo.UpsertMultiple(ctx, db, schoolAdmins); err != nil {
			return errors.Wrap(err, "s.UserModifierService.SchoolAdminRepo.UpsertMultiple")
		}

		if err := s.UserModifierService.TeacherRepo.SoftDeleteMultiple(ctx, db, database.TextArray(userIDs)); err != nil {
			return errors.Wrap(err, "s.UserModifierService.TeacherRepo.SoftDeleteMultiple")
		}

		revokeUserGroup = constant.UserGroupTeacher
	}

	if err := s.UserGroupRepo.UpsertMultiple(ctx, db, listLegacyUserGroups); err != nil {
		return errors.Wrap(err, "s.UserGroupRepo.UpsertMultiple")
	}

	if err := s.UserRepo.UpdateManyUserGroup(ctx, db, database.TextArray(userIDs), database.Text(grantedUserGroup)); err != nil {
		return errors.Wrap(err, "s.UserRepo.UpdateManyUserGroup")
	}

	if err := s.UserGroupRepo.DeactivateMultiple(ctx, db, database.TextArray(userIDs), database.Text(revokeUserGroup)); err != nil {
		return errors.Wrap(err, "s.UserGroupRepo.DeactivateMultiple")
	}

	return nil
}
