package staff

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	usvc "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *StaffService) UpdateStaff(ctx context.Context, req *pb.UpdateStaffRequest) (*pb.UpdateStaffResponse, error) {
	profile := req.Staff
	organization, err := interceptors.OrganizationFromContext(ctx)
	orgID := organization.OrganizationID().String()

	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	isFeatureStaffUsernameEnabled := unleash.IsFeatureStaffUsernameEnabled(s.UnleashClient, s.Env, organization)

	// if feature is not enabled, username will be the same as email
	if !isFeatureStaffUsernameEnabled {
		req.Staff.Username = req.Staff.Email
	}
	req.Staff.Username = strings.TrimSpace(strings.ToLower(req.Staff.Username))

	if err := s.validationsUpdateStaff(ctx, req.Staff, isFeatureStaffUsernameEnabled); err != nil {
		return nil, err
	}

	trimmedExternalUserID := strings.TrimSpace(req.Staff.ExternalUserId)
	if trimmedExternalUserID != "" {
		user := grpc.NewUserProfile(
			&pb.UserProfile{
				UserId:         req.Staff.StaffId,
				ExternalUserId: trimmedExternalUserID,
			},
		)
		if err := s.DomainUser.ValidateExternalUserIDIsExists(ctx, entity.Users{user}); err != nil {
			return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "s.DomainUser.ValidateExternalUserIDIsExists").Error())
		}
	}

	if err := s.checkPermissionToAssignUserGroup(ctx, s.DB, req.Staff.UserGroupIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "checkPermissionToAssignUserGroup").Error())
	}

	legacyUserGroup, err := s.getLegacyUserGroup(ctx, s.DB, req.Staff.UserGroupIds)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "getLegacyUserGroup").Error())
	}

	toUpdateStaff, err := updateStaffPbToStaffEnt(req, legacyUserGroup, orgID)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "UpdateStaffPbToUserEn").Error())
	}

	existingStaff, err := s.findStaffUserByID(ctx, s.DB, toUpdateStaff.ID)
	if err != nil {
		return nil, err
	}

	if err = s.checkPermissionUpdateStaff(existingStaff.GetUID(), interceptors.UserIDFromContext(ctx)); err != nil {
		return nil, err
	}

	userGroupMembers, err := createUserGroupMemberEnt(profile.UserGroupIds, existingStaff.GetUID(), orgID)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "createUserGroupMemberEnt").Error())
	}

	existedUserGroupMembers, err := s.UserGroupV2Service.UserGroupsMemberRepo.GetByUserID(ctx, s.DB, database.Text(profile.GetStaffId()))
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "UserRepo.GetUserGroupMembers").Error())
	}

	tags, err := s.UserModifierService.DomainTagRepo.GetByIDs(ctx, s.DB, profile.TagIds)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "DomainTagRepo.GetByIDs").Error())
	}

	locations, err := s.UserModifierService.GetLocations(ctx, req.GetStaff().GetLocationIds())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// user group member must be revoked from existed with the request
	revokeGroupMembers := userGroupMembersToRevoke(userGroupMembers, existedUserGroupMembers)
	// new user group member must be inserted from existed with the request
	createNewGroupMembers := userGroupMembersToCreateNew(userGroupMembers, existedUserGroupMembers)

	if err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		// update userStaff email if the email is not equal with old email
		if existingStaff.LegacyUser.Email != toUpdateStaff.LegacyUser.Email {
			oldEmail := existingStaff.LegacyUser.Email.String
			existingStaff.LegacyUser.Email = toUpdateStaff.LegacyUser.Email
			existingStaff.LegacyUser.LoginEmail = toUpdateStaff.LegacyUser.Email
			if err := s.updateStaffEmail(ctx, tx, &existingStaff.LegacyUser, oldEmail); err != nil {
				return err
			}
		}

		if len(createNewGroupMembers) > 0 {
			if err := s.UserGroupV2Service.UserGroupsMemberRepo.UpsertBatch(ctx, tx, createNewGroupMembers); err != nil {
				return status.Error(codes.Internal, errors.Wrap(err, "s.UserGroupService.UserGroupsMemberRepo.UpsertBatch").Error())
			}
		}

		if len(revokeGroupMembers) > 0 {
			if err := s.UserGroupV2Service.UserGroupsMemberRepo.SoftDelete(ctx, tx, revokeGroupMembers); err != nil {
				return status.Error(codes.Internal, errors.Wrap(err, "s.UserGroupService.UserGroupsMemberRepo.SoftDelete").Error())
			}
		}

		if legacyUserGroup != existingStaff.Group.String {
			if err := s.revokeGroupOfUser(ctx, tx, existingStaff.GetUID(), existingStaff.Group.String); err != nil {
				return status.Error(codes.Internal, errors.Wrap(err, "s.revokeGroupOfUser").Error())
			}
			if err := s.grantedGroupForUser(ctx, tx, &existingStaff.LegacyUser, legacyUserGroup); err != nil {
				return status.Error(codes.Internal, errors.Wrap(err, "s.grantedGroupForUser").Error())
			}
		}

		if err := usvc.UpsertUserAccessPath(ctx, s.UserAccessPathRepo, tx, locations, existingStaff.GetUID()); err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "usvc.UpsertUserAccessPath").Error())
		}

		assignDataToUpdateToExistingStaff(existingStaff, toUpdateStaff)
		if _, err := s.StaffRepo.Update(ctx, tx, existingStaff); err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "s.StaffRepo.Update").Error())
		}

		userPhoneNumbers, err := pbStaffPhoneNumberToUserPhoneNumber(req.Staff.StaffPhoneNumber, existingStaff.GetUID(), orgID)
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
			return errorx.ToStatusError(err)
		}

		taggedUsers, err := s.UserModifierService.DomainTaggedUserRepo.GetByUserIDs(ctx, tx, []string{profile.GetStaffId()})
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("DomainTaggedUserRepo.GetByUserIDs: %v", err).Error())
		}

		userProfile := grpc.NewUserProfileWithID(existingStaff.GetUID())
		userWithTags := map[entity.User][]entity.DomainTag{userProfile: tags}
		if err := s.UserModifierService.UpsertTaggedUsers(ctx, tx, userWithTags, taggedUsers); err != nil {
			return errors.Wrap(err, "UpsertTaggedUsers")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	eventUpsertStaff := toEventUpsertStaff(profile.StaffId, profile.UserGroupIds, req.Staff.LocationIds, pb.EvtUpsertStaff_UPSERT_STAFF_TYPE_UPDATE)
	err = s.publishUpsertStaffEvent(ctx, constants.SubjectUpsertStaff, eventUpsertStaff)
	if err != nil {
		return nil, errorx.ToStatusError(err)
	}

	return &pb.UpdateStaffResponse{Successful: true}, nil
}

func (s *StaffService) updateStaffEmail(ctx context.Context, tx database.QueryExecer, userEnt *entity.LegacyUser, oldEmail string) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	orgID := golibs.ResourcePathFromCtx(ctx)

	emailExistingStaff, err := s.UserModifierService.UserRepo.GetByEmailInsensitiveCase(ctx, tx, []string{userEnt.Email.String})
	if err != nil {
		return status.Error(codes.Internal, errors.Wrap(err, "s.UserModifierService.UserRepo.GetByEmailInsensitiveCase").Error())
	}
	if len(emailExistingStaff) > 0 {
		return status.Error(codes.AlreadyExists, errcode.ErrUserEmailExists.Error())
	}

	err = s.UserModifierService.UsrEmailRepo.UpdateEmail(ctx, tx, userEnt.ID, database.Text(orgID), userEnt.Email)
	switch err {
	case nil:
		break
	case repository.ErrNoRowAffected:
		// it's ok to have no row affected
		break
	default:
		return status.Error(codes.Internal, errors.Wrap(err, "s.UserModifierService.UsrEmailRepo.UpdateEmail").Error())
	}

	// find tenantID
	tenantID, err := s.UserModifierService.OrganizationRepo.GetTenantIDByOrgID(ctx, tx, orgID)
	if err != nil {
		zapLogger.Error(
			"cannot get tenant id",
			zap.Error(err),
			zap.String("organizationID", orgID),
		)
		switch err {
		case pgx.ErrNoRows:
			return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: orgID}.Error())
		default:
			return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
		}
	}

	// update user email in identity platform
	err = s.UserModifierService.UpdateUserEmailInIdentityPlatform(ctx, tenantID, userEnt.ID.String, userEnt.Email.String)
	if err != nil {
		zapLogger.Error(
			"cannot update users on identity platform",
			zap.Error(err),
			zap.String("organizationID", orgID),
			zap.String("tenantID", tenantID),
			zap.String("uid", userEnt.GetUID()),
			zap.String("email", oldEmail),
			zap.String("emailToUpdate", userEnt.Email.String),
		)
		switch err {
		case user.ErrUserNotFound:
			return status.Error(codes.NotFound, errcode.NewUserNotFoundErr(userEnt.ID.String).Error())
		default:
			return status.Error(codes.Internal, err.Error())
		}
	}

	return nil
}

func (s *StaffService) checkPermissionUpdateStaff(staffID, currentUserID string) error {
	// check role is school admin and has diff id
	if currentUserID == staffID {
		return status.Error(codes.PermissionDenied, "school admin can only update their staff profile")
	}

	return nil
}

func (s *StaffService) validationsUpdateStaff(ctx context.Context, staffProfile *pb.UpdateStaffRequest_StaffProfile, isFeatureStaffUsernameEnabled bool) error {
	var startDate, endDate = staffProfile.StartDate, staffProfile.EndDate
	if staffProfile.UserNameFields != nil {
		if staffProfile.UserNameFields.LastName == "" {
			return status.Error(codes.InvalidArgument, errcode.ErrUserLastNameIsEmpty.Error())
		}
		if staffProfile.UserNameFields.FirstName == "" {
			return status.Error(codes.InvalidArgument, errcode.ErrUserFirstNameIsEmpty.Error())
		}
	} else if staffProfile.Name == "" {
		return status.Error(codes.InvalidArgument, errcode.ErrUserFullNameIsEmpty.Error())
	}

	if isFeatureStaffUsernameEnabled {
		user := grpc.NewUserProfile(
			&pb.UserProfile{
				UserId:   staffProfile.StaffId,
				Username: staffProfile.Username,
			},
		)
		if err := s.validateStaffUsername(ctx, s.DB, user); err != nil {
			return err
		}
	}

	switch {
	case staffProfile.Email == "":
		return status.Error(codes.InvalidArgument, errcode.ErrUserEmailIsEmpty.Error())
	case startDate != nil && !startDate.IsValid():
		return status.Error(codes.InvalidArgument, errcode.ErrStaffStartDateIsInvalid.Error())
	case endDate != nil && !endDate.IsValid():
		return status.Error(codes.InvalidArgument, errcode.ErrStaffEndDateIsInvalid.Error())
	case startDate != nil && endDate != nil && endDate.AsTime().Before(startDate.AsTime()):
		return status.Error(codes.InvalidArgument, errcode.ErrStaffStartDateIsLessThanEndDate.Error())
	}

	if err := validateStaffPhoneNumber(staffProfile.StaffPhoneNumber); err != nil {
		return err
	}

	trimmedExternalUserID := strings.TrimSpace(staffProfile.ExternalUserId)
	if trimmedExternalUserID != "" {
		user := grpc.NewUserProfile(
			&pb.UserProfile{
				UserId:         staffProfile.StaffId,
				ExternalUserId: trimmedExternalUserID,
			},
		)
		err := s.DomainUser.ValidateExternalUserIDExistedInSystem(ctx, entity.Users{user})
		if err != nil {
			return status.Error(codes.AlreadyExists, errors.Wrap(err, "s.DomainUser.validateExternalUserIDExistedInSystem").Error())
		}
	}

	if len(staffProfile.GetLocationIds()) == 0 {
		return status.Error(codes.InvalidArgument, errcode.ErrUserLocationIsEmpty.Error())
	}

	if err := s.validateStaffUserGroup(ctx, s.DB, staffProfile.UserGroupIds); err != nil {
		return err
	}

	if len(staffProfile.TagIds) > 0 {
		staffTags, err := s.UserModifierService.DomainTagRepo.GetByIDs(ctx, s.DB, staffProfile.TagIds)
		if err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "UserModifierService.DomainTagRepo.GetByIDs").Error())
		}

		if err := validateStaffTags(staffProfile.TagIds, staffTags); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return nil
}

func (s *StaffService) findStaffUserByID(ctx context.Context, db database.QueryExecer, id pgtype.Text) (*entity.Staff, error) {
	staff, err := s.StaffRepo.FindByID(ctx, db, id)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "s.StaffRepo.FindByID").Error())
	}
	if staff == nil {
		return nil, status.Error(codes.InvalidArgument, errors.Errorf("staff with id %s not found", id.String).Error())
	}

	return staff, nil
}

func updateStaffPbToStaffEnt(req *pb.UpdateStaffRequest, userGroup string, resourcePath string) (*entity.Staff, error) {
	staff := new(entity.Staff)
	database.AllNullEntity(staff)
	database.AllNullEntity(&staff.LegacyUser)
	genderForCheckNil := &req.Staff.Gender

	if err := req.Staff.Birthday.CheckValid(); err == nil {
		err = staff.Birthday.Set(req.Staff.Birthday.AsTime())
		if err != nil {
			return nil, fmt.Errorf("updateStaffPbToStaffEnt staff.Birthday.Set: %v", err)
		}
	}

	if genderForCheckNil != nil && req.Staff.Gender != pb.Gender_NONE {
		err := staff.Gender.Set(req.Staff.Gender.String())
		if err != nil {
			return nil, fmt.Errorf("updateStaffPbToStaffEnt staff.Gender.Set: %v", err)
		}
	}

	if strings.TrimSpace(req.Staff.ExternalUserId) != "" {
		staff.ExternalUserID = database.Text(strings.TrimSpace(req.Staff.ExternalUserId))
	}
	// Temporarily set loginEmail equal Email
	_ = staff.LoginEmail.Set(req.Staff.Email)
	firstName, lastName := usvc.SplitNameToFirstNameAndLastName(req.Staff.Name)
	err := multierr.Combine(
		staff.ID.Set(req.Staff.StaffId),
		staff.FullName.Set(req.Staff.Name),
		staff.LastName.Set(lastName),
		staff.FirstName.Set(firstName),
		staff.Email.Set(req.Staff.Email),
		staff.Group.Set(userGroup),
		staff.UpdatedAt.Set(time.Now()),
		staff.Remarks.Set(req.Staff.Remarks),
		staff.LegacyUser.ResourcePath.Set(resourcePath),

		staff.WorkingStatus.Set(req.Staff.WorkingStatus),
		staff.StartDate.Set(database.DateFromPb(req.Staff.StartDate)),
		staff.EndDate.Set(database.DateFromPb(req.Staff.EndDate)),
		staff.UserName.Set(req.Staff.Username),
		staff.ResourcePath.Set(resourcePath),
	)
	if req.Staff.UserNameFields != nil {
		if err := multierr.Combine(
			staff.FullName.Set(usvc.CombineFirstNameAndLastNameToFullName(req.Staff.UserNameFields.FirstName, req.Staff.UserNameFields.LastName)),
			staff.LastName.Set(req.Staff.UserNameFields.LastName),
			staff.FirstName.Set(req.Staff.UserNameFields.FirstName),
			staff.FullNamePhonetic.Set(usvc.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(req.Staff.UserNameFields.FirstNamePhonetic, req.Staff.UserNameFields.LastNamePhonetic)),
			staff.LastNamePhonetic.Set(req.Staff.UserNameFields.LastNamePhonetic),
			staff.FirstNamePhonetic.Set(req.Staff.UserNameFields.FirstNamePhonetic),
		); err != nil {
			return nil, err
		}
	}

	return staff, err
}

func assignDataToUpdateToExistingStaff(existingStaff *entity.Staff, toUpdateStaff *entity.Staff) {
	existingStaff.LegacyUser.FullName = toUpdateStaff.LegacyUser.FullName
	existingStaff.LegacyUser.LastName = toUpdateStaff.LegacyUser.LastName
	existingStaff.LegacyUser.FirstName = toUpdateStaff.LegacyUser.FirstName
	existingStaff.LegacyUser.FullNamePhonetic = toUpdateStaff.LegacyUser.FullNamePhonetic
	existingStaff.LegacyUser.LastNamePhonetic = toUpdateStaff.LegacyUser.LastNamePhonetic
	existingStaff.LegacyUser.FirstNamePhonetic = toUpdateStaff.LegacyUser.FirstNamePhonetic

	existingStaff.LegacyUser.Group = toUpdateStaff.LegacyUser.Group
	existingStaff.LegacyUser.Remarks = toUpdateStaff.LegacyUser.Remarks
	existingStaff.LegacyUser.Gender = toUpdateStaff.LegacyUser.Gender
	existingStaff.LegacyUser.Birthday = toUpdateStaff.LegacyUser.Birthday
	existingStaff.LegacyUser.ExternalUserID = toUpdateStaff.LegacyUser.ExternalUserID
	existingStaff.LegacyUser.UserName = toUpdateStaff.LegacyUser.UserName

	existingStaff.WorkingStatus = toUpdateStaff.WorkingStatus
	existingStaff.StartDate = toUpdateStaff.StartDate
	existingStaff.EndDate = toUpdateStaff.EndDate
}

func userGroupMembersToCreateNew(updateUserGroupMembers []*entity.UserGroupMember, existedUserGroupMembers []*entity.UserGroupMember) []*entity.UserGroupMember {
	createNewUserGroupMembers := []*entity.UserGroupMember{}
	mapUserGroupMembers := mapUserGroupMember(existedUserGroupMembers)
	for _, updateUserGroupMember := range updateUserGroupMembers {
		if _, ok := mapUserGroupMembers[updateUserGroupMember.UserGroupID.String]; !ok {
			createNewUserGroupMembers = append(createNewUserGroupMembers, updateUserGroupMember)
		}
	}
	return createNewUserGroupMembers
}

func userGroupMembersToRevoke(updateUserGroupMembers []*entity.UserGroupMember, existedUserGroupMembers []*entity.UserGroupMember) []*entity.UserGroupMember {
	revokeUserGroupMembers := []*entity.UserGroupMember{}
	mapUserGroupMembers := mapUserGroupMember(updateUserGroupMembers)
	for _, existedUserGroupMember := range existedUserGroupMembers {
		if _, ok := mapUserGroupMembers[existedUserGroupMember.UserGroupID.String]; !ok {
			revokeUserGroupMembers = append(revokeUserGroupMembers, existedUserGroupMember)
		}
	}

	return revokeUserGroupMembers
}

func mapUserGroupMember(userGroupMembers []*entity.UserGroupMember) map[string]struct{} {
	mapper := map[string]struct{}{}
	for _, userGroupMember := range userGroupMembers {
		mapper[userGroupMember.UserGroupID.String] = struct{}{}
	}
	return mapper
}
