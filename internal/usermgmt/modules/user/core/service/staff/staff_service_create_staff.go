package staff

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	internal_auth_user "github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	usvc "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/service"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ErrTagIDsMustBeExisted = errors.New("tag ids must be existed in system")
	ErrTagIsNotForStaff    = errors.New("this tag is not for staff")
	ErrTagIsArchived       = errors.New("this tag is archived")
)

func (s *StaffService) handleCreateStaff(ctx context.Context, staffProfile *pb.CreateStaffRequest_StaffProfile, userGroup string) (*entity.Staff, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	orgID, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	staff, err := createStaffPbToStaffEnt(staffProfile, userGroup, orgID.OrganizationID().String())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "createStaffPbToStaffEnt").Error())
	}

	// Guarantee each user email has only corresponding one uid
	createdUsrEmailForStaff, err := s.UserModifierService.UsrEmailRepo.Create(ctx, s.DB, staff.ID, staff.Email)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "s.UserModifierService.UsrEmailRepo.Create").Error())
	}
	staff.ID = createdUsrEmailForStaff.UsrID

	var locations []*domain.Location
	if len(staffProfile.GetLocationIds()) > 0 {
		locations, err = s.UserModifierService.GetLocations(ctx, staffProfile.GetLocationIds())
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "UserModifierService.GetLocations").Error())
		}
	}

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		userGroupMembers, err := createUserGroupMemberEnt(staffProfile.UserGroupIds, staff.ID.String, orgID.OrganizationID().String())
		if err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "createUserGroupMemberEnt").Error())
		}

		if err := s.StaffRepo.Create(ctx, tx, staff); err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "s.StaffRepo.Create").Error())
		}

		// assign user group for staff
		if len(userGroupMembers) > 0 {
			if err := s.UserGroupV2Service.UserGroupsMemberRepo.UpsertBatch(ctx, tx, userGroupMembers); err != nil {
				return errors.Wrap(err, "s.UserGroupService.UserGroupsMemberRepo.UpsertBatch")
			}
		}

		// add location for staff
		if err := usvc.UpsertUserAccessPath(ctx, s.UserAccessPathRepo, tx, locations, staff.ID.String); err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "usvc.UpsertUserAccessPath").Error())
		}

		// Import to identity platform
		tenantID, err := s.UserModifierService.OrganizationRepo.GetTenantIDByOrgID(ctx, tx, orgID.OrganizationID().String())
		if err != nil {
			zapLogger.Error(
				"cannot get tenant id",
				zap.Error(err),
				zap.String("organizationID", orgID.OrganizationID().String()),
			)
			switch err {
			case pgx.ErrNoRows:
				return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: orgID.OrganizationID().String()}.Error())
			default:
				return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
			}
		}

		err = s.UserModifierService.CreateUsersInIdentityPlatform(ctx, tenantID, []*entity.LegacyUser{&staff.LegacyUser}, int64(orgID.SchoolID().Int32()))
		if err != nil {
			zapLogger.Error(
				"cannot create users on identity platform",
				zap.Error(err),
				zap.String("organizationID", orgID.OrganizationID().String()),
				zap.String("tenantID", tenantID),
				zap.String("email", staff.Email.String),
			)
			switch err {
			case internal_auth_user.ErrUserNotFound:
				return status.Error(codes.NotFound, errcode.NewUserNotFoundErr(staff.ID.String).Error())
			default:
				return status.Error(codes.Internal, errors.Wrap(err, "cannot create user").Error())
			}
		}

		userPhoneNumbers, err := pbStaffPhoneNumberToUserPhoneNumber(staffProfile.StaffPhoneNumber, staff.LegacyUser.ID.String, orgID.OrganizationID().String())
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
			return errorx.ToStatusError(err)
		}

		if len(staffProfile.TagIds) > 0 {
			staffTags, err := s.UserModifierService.DomainTagRepo.GetByIDs(ctx, s.DB, staffProfile.TagIds)
			if err != nil {
				return status.Error(codes.Internal, errors.Wrap(err, "UserModifierService.DomainTagRepo.GetByIDs").Error())
			}
			userProfile := grpc.NewUserProfileWithID(staff.GetUID())
			staffWithTags := map[entity.User][]entity.DomainTag{userProfile: staffTags}

			if err := s.UserModifierService.UpsertTaggedUsers(ctx, tx, staffWithTags, nil); err != nil {
				return errors.Wrap(err, "UserModifierService.UpsertTaggedUsers")
			}
		}

		// backward compatible with old school admin and teacher
		// will be deprecated soon
		switch userGroup {
		case constant.UserGroupSchoolAdmin:
			if err := s.createSchoolAdmin(ctx, tx, &staff.LegacyUser, []int64{int64(orgID.SchoolID().Int32())}); err != nil {
				return status.Error(codes.Internal, errors.Wrapf(err, "s.createSchoolAdmin: %s", userGroup).Error())
			}

		case constant.UserGroupTeacher:
			if err := s.createTeacher(ctx, tx, &staff.LegacyUser, []int64{int64(orgID.SchoolID().Int32())}); err != nil {
				return status.Error(codes.Internal, errors.Wrapf(err, "s.createTeacher: %s", userGroup).Error())
			}

		default:
			return status.Error(codes.Internal, "invalid legacy user_group")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return staff, err
}

func (s *StaffService) CreateStaff(ctx context.Context, req *pb.CreateStaffRequest) (*pb.CreateStaffResponse, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	isFeatureStaffUsernameEnabled := unleash.IsFeatureStaffUsernameEnabled(s.UnleashClient, s.Env, organization)

	// if feature is not enabled, username will be the same as email
	if !isFeatureStaffUsernameEnabled {
		req.Staff.Username = req.Staff.Email
	}
	req.Staff.Username = strings.TrimSpace(strings.ToLower(req.Staff.Username))

	if err := s.validationsCreateStaff(ctx, req.Staff, isFeatureStaffUsernameEnabled); err != nil {
		return nil, err
	}

	if err := s.checkPermissionToAssignUserGroup(ctx, s.DB, req.Staff.UserGroupIds); err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "checkPermissionToAssignUserGroup").Error())
	}

	legacyUserGroup, err := s.getLegacyUserGroup(ctx, s.DB, req.Staff.UserGroupIds)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, errors.Wrap(err, "getLegacyUserGroup").Error())
	}

	staff, err := s.handleCreateStaff(ctx, req.Staff, legacyUserGroup)
	if err != nil {
		return nil, err
	}

	eventStaffConfig := toEventStaffUpsertTimesheetSetting(staff.ID.String, false, staff.UpdatedAt.Time)
	err = s.publishStaffSettingEvent(ctx, constants.SubjectStaffUpsertTimesheetConfig, eventStaffConfig)
	if err != nil {
		return nil, errorx.ToStatusError(err)
	}

	eventUpsertStaff := toEventUpsertStaff(staff.ID.String, req.Staff.UserGroupIds, req.Staff.LocationIds, pb.EvtUpsertStaff_UPSERT_STAFF_TYPE_CREATE)
	err = s.publishUpsertStaffEvent(ctx, constants.SubjectUpsertStaff, eventUpsertStaff)
	if err != nil {
		return nil, errorx.ToStatusError(err)
	}

	return &pb.CreateStaffResponse{
		Staff: &pb.CreateStaffResponse_StaffProfile{
			StaffId:          staff.ID.String,
			Name:             staff.FullName.String,
			Email:            staff.Email.String,
			Avatar:           staff.Avatar.String,
			PhoneNumber:      staff.PhoneNumber.String,
			Country:          cpb.Country(cpb.Country_value[staff.Country.String]),
			UserGroup:        req.Staff.UserGroup,
			OrganizationId:   req.Staff.OrganizationId,
			LocationIds:      req.Staff.LocationIds,
			StaffPhoneNumber: req.Staff.StaffPhoneNumber,
			Birthday:         timestamppb.New(staff.Birthday.Time),
			Gender:           req.Staff.Gender,
			TagIds:           req.Staff.TagIds,
			ExternalUserId:   req.Staff.ExternalUserId,
			Username:         req.Staff.Username,
		},
	}, nil
}

func validateStaffPhoneNumber(staffsPhoneNumber []*pb.StaffPhoneNumber) error {
	isHadPrimaryPhone := false
	primaryPhoneType := pb.StaffPhoneNumberType_STAFF_PRIMARY_PHONE_NUMBER
	secondaryPhoneType := pb.StaffPhoneNumberType_STAFF_SECONDARY_PHONE_NUMBER

	if len(staffsPhoneNumber) == 0 {
		return nil
	}

	// Sort before validate because need to compare 2 near value to check duplicate
	sort.Slice(staffsPhoneNumber, func(i, j int) bool {
		return staffsPhoneNumber[i].PhoneNumber < staffsPhoneNumber[j].PhoneNumber
	})

	for index, staffPhoneNumber := range staffsPhoneNumber {
		staffPhoneNumber.PhoneNumber = strings.TrimSpace(staffPhoneNumber.PhoneNumber)

		err := usvc.MatchingRegex(usvc.PhoneNumberPattern, staffPhoneNumber.PhoneNumber)

		switch {
		case staffPhoneNumber.PhoneNumber == "":
			return nil

		case staffPhoneNumber.PhoneNumberType != primaryPhoneType &&
			staffPhoneNumber.PhoneNumberType != secondaryPhoneType:
			return status.Error(codes.InvalidArgument, errcode.ErrUserPhoneNumberIsWrongType.Error())

		case err != nil:
			return status.Error(codes.InvalidArgument, err.Error())
		case index > 0 && staffPhoneNumber.PhoneNumber == staffsPhoneNumber[index-1].PhoneNumber:
			return status.Error(codes.InvalidArgument, errcode.ErrUserPhoneNumberIsDuplicate.Error())

		case isHadPrimaryPhone && staffPhoneNumber.PhoneNumberType == primaryPhoneType:
			return status.Error(codes.InvalidArgument, errcode.ErrUserPrimaryPhoneNumberIsRedundant.Error())

		case !isHadPrimaryPhone && staffPhoneNumber.PhoneNumberType == primaryPhoneType:
			isHadPrimaryPhone = true
		}
	}

	return nil
}

func pbStaffPhoneNumberToUserPhoneNumber(staffPhoneNumbers []*pb.StaffPhoneNumber, userID string, resourcePath string) ([]*entity.UserPhoneNumber, error) {
	userPhoneNumbers := make([]*entity.UserPhoneNumber, 0)

	for i, userPhoneNumber := range staffPhoneNumbers {
		newStaffPhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(newStaffPhoneNumber)
		userPhoneNumberID := idutil.ULIDNow()

		if staffPhoneNumbers[i].GetPhoneNumberId() != "" {
			userPhoneNumberID = staffPhoneNumbers[i].GetPhoneNumberId()
		}

		if err := multierr.Combine(
			newStaffPhoneNumber.ID.Set(userPhoneNumberID),
			newStaffPhoneNumber.UserID.Set(userID),
			newStaffPhoneNumber.PhoneNumber.Set(userPhoneNumber.PhoneNumber),
			newStaffPhoneNumber.PhoneNumberType.Set(userPhoneNumber.PhoneNumberType),
			newStaffPhoneNumber.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, fmt.Errorf("pbStaffPhoneNumberToUserPhoneNumber multierr.Combine: %v", err)
		}

		userPhoneNumbers = append(userPhoneNumbers, newStaffPhoneNumber)
	}
	return userPhoneNumbers, nil
}

func validateStaffTags(tagIDs []string, existingTags entity.DomainTags) error {
	if len(tagIDs) == 0 {
		return nil
	}

	if ok := existingTags.ContainIDs(tagIDs...); !ok {
		return ErrTagIDsMustBeExisted
	}

	for _, tag := range existingTags {
		if tag.IsArchived().Boolean() {
			return ErrTagIsArchived
		}

		if !entity.IsStaffTag(tag) {
			return ErrTagIsNotForStaff
		}
	}

	return nil
}

func (s *StaffService) validateStaffUsername(ctx context.Context, db database.QueryExecer, user entity.User) error {
	if err := entity.ValidateUserName(user); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	if err := usvc.ValidateUserNamesExistedInSystem(ctx, s.UserRepo, db, entity.Users{user}); err != nil {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	return nil
}

func (s *StaffService) validationsCreateStaff(ctx context.Context, staffProfile *pb.CreateStaffRequest_StaffProfile, isFeatureStaffUsernameEnabled bool) error {
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
				Username: staffProfile.Username,
			},
		)
		if err := s.validateStaffUsername(ctx, s.DB, user); err != nil {
			return err
		}
	}

	switch {
	case staffProfile.Country == cpb.Country_COUNTRY_NONE:
		return status.Error(codes.InvalidArgument, errcode.ErrUserCountryIsEmpty.Error())
	case len(staffProfile.GetLocationIds()) == 0:
		return status.Error(codes.InvalidArgument, errcode.ErrUserLocationIsEmpty.Error())
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
		externalUserIDs := []string{trimmedExternalUserID}
		existingUsers, err := s.UserRepo.GetByExternalUserIDs(ctx, s.DB, externalUserIDs)
		if err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "s.UserRepo.GetUsersByExternalIDs").Error())
		}

		if len(existingUsers) != 0 {
			return status.Error(codes.AlreadyExists, errcode.ErrUserExternalUserIDExists.Error())
		}
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

	staffInDB, err := s.UserModifierService.UserRepo.GetByEmailInsensitiveCase(ctx, s.DB, []string{staffProfile.Email})
	if err != nil {
		return status.Error(codes.Internal, errors.Wrap(err, "s.UserRepo.GetByEmailInsensitiveCase").Error())
	}

	if len(staffInDB) != 0 {
		return status.Error(codes.AlreadyExists, errcode.ErrUserEmailExists.Error())
	}

	if err := s.validateStaffUserGroup(ctx, s.DB, staffProfile.UserGroupIds); err != nil {
		return err
	}

	return nil
}

func createStaffPbToStaffEnt(userProfile *pb.CreateStaffRequest_StaffProfile, userGroup, organization string) (*entity.Staff, error) {
	user := &entity.LegacyUser{}
	database.AllNullEntity(user)
	if err := userProfile.Birthday.CheckValid(); err == nil {
		_ = user.Birthday.Set(userProfile.Birthday.AsTime())
	}
	if userProfile.Gender != pb.Gender_NONE {
		_ = user.Gender.Set(userProfile.Gender.String())
	}

	if strings.TrimSpace(userProfile.ExternalUserId) != "" {
		user.ExternalUserID = database.Text(strings.TrimSpace(userProfile.ExternalUserId))
	}
	// Temporarily set loginEmail equal Email
	_ = user.LoginEmail.Set(userProfile.Email)
	firstName, lastName := usvc.SplitNameToFirstNameAndLastName(userProfile.Name)
	if err := multierr.Combine(
		user.ID.Set(idutil.ULIDNow()),
		user.FullName.Set(userProfile.Name),
		user.FirstName.Set(firstName),
		user.LastName.Set(lastName),
		user.Country.Set(userProfile.Country.String()),
		user.Group.Set(userGroup),
		user.Email.Set(userProfile.Email),
		user.ResourcePath.Set(organization),
		user.UserRole.Set(constant.UserRoleStaff),
		user.UserName.Set(userProfile.Username),
	); err != nil {
		return nil, fmt.Errorf("err set user: %w", err)
	}
	if userProfile.UserNameFields != nil {
		if err := multierr.Combine(
			user.FullName.Set(usvc.CombineFirstNameAndLastNameToFullName(userProfile.UserNameFields.FirstName, userProfile.UserNameFields.LastName)),
			user.LastName.Set(userProfile.UserNameFields.LastName),
			user.FirstName.Set(userProfile.UserNameFields.FirstName),
			user.FullNamePhonetic.Set(usvc.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(userProfile.UserNameFields.FirstNamePhonetic, userProfile.UserNameFields.LastNamePhonetic)),
			user.LastNamePhonetic.Set(userProfile.UserNameFields.LastNamePhonetic),
			user.FirstNamePhonetic.Set(userProfile.UserNameFields.FirstNamePhonetic),
		); err != nil {
			return nil, fmt.Errorf("err set staff name: %w", err)
		}
	}

	// set remarks
	if userProfile.Remarks != "" {
		if err := user.Remarks.Set(userProfile.Remarks); err != nil {
			return nil, fmt.Errorf("err set user remarks: %w", err)
		}
	}

	staff := &entity.Staff{}
	database.AllNullEntity(staff)
	if err := multierr.Combine(
		staff.ID.Set(user.ID),
		staff.ResourcePath.Set(user.ResourcePath),
		staff.DeletedAt.Set(pgtype.Timestamptz{Status: pgtype.Null}),
		staff.AutoCreateTimesheet.Set(database.Bool(false)),
		staff.WorkingStatus.Set(database.Text(userProfile.WorkingStatus.String())),
		staff.StartDate.Set(database.DateFromPb(userProfile.StartDate)),
		staff.EndDate.Set(database.DateFromPb(userProfile.EndDate)),
	); err != nil {
		return nil, fmt.Errorf("err set staff: %w", err)
	}

	staff.LegacyUser = *user
	return staff, nil
}

func createUserGroupMemberEnt(userGroupIDs []string, userID, resourcePath string) ([]*entity.UserGroupMember, error) {
	userGroupMems := make([]*entity.UserGroupMember, 0)
	for _, userGroupID := range userGroupIDs {
		userGroupMem := &entity.UserGroupMember{}
		database.AllNullEntity(userGroupMem)
		if err := multierr.Combine(
			userGroupMem.UserID.Set(userID),
			userGroupMem.UserGroupID.Set(userGroupID),
			userGroupMem.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, err
		}
		userGroupMems = append(userGroupMems, userGroupMem)
	}
	return userGroupMems, nil
}

func (s *StaffService) getLegacyUserGroup(ctx context.Context, db database.QueryExecer, userGroupIDs []string) (string, error) {
	if len(userGroupIDs) == 0 {
		return constant.UserGroupTeacher, nil
	}

	roles, err := s.RoleRepo.GetByUserGroupIDs(ctx, db, userGroupIDs)
	if err != nil {
		return "", fmt.Errorf("GetByUserGroupID failed: %w", err)
	}

	if len(roles) == 0 {
		return constant.UserGroupTeacher, nil
	}
	roleNames := roles.RoleNames()

	// legacy user_group: in case mapping user_group with legacy user_group granted for staff appear legacy UserGroupSchoolAdmin
	// -> return UserGroupSchoolAdmin
	for _, role := range roleNames {
		if constant.MapRoleWithLegacyUserGroup[role] == constant.UserGroupSchoolAdmin {
			return constant.UserGroupSchoolAdmin, nil
		}
	}

	return constant.UserGroupTeacher, nil
}

func (s *StaffService) checkPermissionToAssignUserGroup(ctx context.Context, db database.QueryExecer, userGroupIDs []string) error {
	roles, err := s.RoleRepo.GetByUserGroupIDs(ctx, db, userGroupIDs)
	if err != nil {
		return err
	}

	if !golibs.InArrayString(constant.RoleSchoolAdmin, roles.RoleNames()) {
		return nil
	}

	currentUserID := interceptors.UserIDFromContext(ctx)
	currentUserRoles, err := s.UserRepo.GetUserRoles(ctx, db, currentUserID)
	if err != nil {
		return err
	}
	if len(currentUserRoles) == 0 {
		return fmt.Errorf("current user don't have permission to assign user_group")
	}

	if !golibs.InArrayString(constant.RoleSchoolAdmin, currentUserRoles.RoleNames()) {
		return fmt.Errorf("%s can't assign this user group", currentUserRoles.RoleNames())
	}

	return nil
}
