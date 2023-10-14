package staff

import (
	"context"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *StaffService) validateStaffUserGroup(ctx context.Context, db database.QueryExecer, userGroupIDs []string) error {
	if len(userGroupIDs) == 0 {
		return nil
	}

	userGroups, err := s.UserGroupV2Service.UserGroupV2Repo.FindByIDs(ctx, db, userGroupIDs)
	if err != nil {
		return status.Error(codes.Internal, errors.Wrap(err, "UserGroupV2Repo.FindByIDs").Error())
	}

	if len(userGroups) != len(userGroupIDs) {
		return status.Error(codes.InvalidArgument, errcode.ErrUserUserGroupDoesNotExist.Error())
	}
	return nil
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

func (s *StaffService) grantedGroupForUser(ctx context.Context, db database.QueryExecer, user *entity.LegacyUser, group string) error {
	resourcePath, err := strconv.ParseInt(user.ResourcePath.String, 10, 32)
	if err != nil {
		return err
	}
	switch group {
	case constant.UserGroupSchoolAdmin:
		schoolAdmin := userToSchoolAdmin(user, int32(resourcePath))
		if err := s.SchoolAdminRepo.Upsert(ctx, db, schoolAdmin); err != nil {
			return err
		}
	case constant.UserGroupTeacher:
		teacher := userToTeacher(user, []int32{int32(resourcePath)})
		if err := s.TeacherRepo.Upsert(ctx, db, teacher); err != nil {
			return err
		}
	}

	userGroup := newUserGroupEntity(user.ID.String, group, entity.UserGroupStatusActive, true, user.ResourcePath.String)
	if err := s.setDefaultUserGroup(ctx, db, userGroup); err != nil {
		return errors.Wrap(err, "UserGroupRepo.Upsert")
	}
	return nil
}

func (s *StaffService) revokeGroupOfUser(ctx context.Context, db database.QueryExecer, userID, group string) error {
	switch group {
	case constant.UserGroupSchoolAdmin:
		if err := s.SchoolAdminRepo.SoftDelete(ctx, db, database.Text(userID)); err != nil {
			return err
		}
	case constant.UserGroupTeacher:
		if err := s.TeacherRepo.SoftDelete(ctx, db, database.Text(userID)); err != nil {
			return err
		}
	}
	return nil
}

func (s *StaffService) setDefaultUserGroup(ctx context.Context, db database.QueryExecer, userGroup *entity.UserGroup) error {
	// set inactive for all user group
	if err := s.UserGroupRepo.UpdateOrigin(ctx, db, userGroup.UserID, database.Bool(false)); err != nil {
		return errors.Wrap(err, "s.UserGroupRepo.UpdateOrigin")
	}

	if err := s.UserGroupRepo.UpdateStatus(ctx, db, userGroup.UserID, database.Text(entity.UserGroupStatusInActive)); err != nil {
		return errors.Wrap(err, "s.UserGroupRepo.UpdateStatus")
	}

	// set active for specific user group
	userGroup.IsOrigin = database.Bool(true)
	userGroup.Status = database.Text(entity.UserGroupStatusActive)
	if err := s.UserGroupRepo.Upsert(ctx, db, userGroup); err != nil {
		return errors.Wrap(err, "s.UserGroupRepo.Upsert")
	}
	return nil
}
