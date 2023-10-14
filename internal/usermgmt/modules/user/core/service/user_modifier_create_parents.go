package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserModifierService) CreateParentsAndAssignToStudent(ctx context.Context, req *pb.CreateParentsAndAssignToStudentRequest) (*pb.CreateParentsAndAssignToStudentResponse, error) {
	parentProfilePBs := make([]*pb.CreateParentsAndAssignToStudentResponse_ParentProfile, 0, len(req.ParentProfiles))

	zapLogger := ctxzap.Extract(ctx).Sugar()
	parentIDs := make([]string, 0, len(req.ParentProfiles))

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	isEnableUsernameToggle := unleash.IsFeatureUserNameStudentParentEnabled(s.UnleashClient, s.Env, organization)
	authUsernameConfig, err := s.InternalConfigurationRepo.GetByKey(ctx, s.DB, constant.KeyAuthUsernameConfig)
	isAuthUsernameConfigOn := false
	if err != nil {
		if !strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return nil, status.Error(codes.Internal, err.Error())
		}
	} else {
		isAuthUsernameConfigOn = authUsernameConfig.ConfigValue().String() == constant.ConfigValueOn
	}
	isEnableUsername := isEnableUsernameToggle && isAuthUsernameConfigOn

	if err := validCreateParentRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	parentDomains := make(entity.Users, 0, len(req.ParentProfiles))
	for _, parentProfile := range req.ParentProfiles {
		// if feature is not enabled, username will be the same as email
		if !isEnableUsername {
			parentProfile.Username = parentProfile.Email
		}

		parentProfile.Username = strings.TrimSpace(strings.ToLower(parentProfile.Username))

		parent := createParentProfileToParentDomain(parentProfile)
		if err := entity.ValidParent(parent, isEnableUsername); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		parentDomains = append(parentDomains, parent)
	}

	if err := ValidateUserNamesExistedInSystem(ctx, s.DomainUserRepo, s.DB, parentDomains); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// the request has list parent, each parent has list tag ids
	// merge tag ids into 1 slice and query record data
	tagIDs := getTagIDsFromParentProfiles(req)
	tagInDBs, err := s.DomainTagRepo.GetByIDs(ctx, s.DB, tagIDs)
	if err != nil {
		return nil, status.Error(codes.Internal, errors.Wrap(err, "DomainTagRepo.GetByIDs").Error())
	}

	if err := validUserTags(constant.RoleParent, tagIDs, tagInDBs); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	// map id -> self records data
	mapExistedTags := map[string]entity.DomainTag{}
	for _, tag := range tagInDBs {
		mapExistedTags[tag.TagID().String()] = tag
	}

	resourcePath, err := strconv.ParseInt(golibs.ResourcePathFromCtx(ctx), 10, 32)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "resource path is invalid")
	}

	toCreateParents, userPhoneNumbers, err := toCreateParents(int32(resourcePath), req)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Guarantee each user email has only corresponding one uid
	for _, toCreateParent := range toCreateParents {
		if isEnableUsername {
			_ = toCreateParent.LegacyUser.LoginEmail.Set(toCreateParent.ID.String + constant.LoginEmailPostfix)
		}
		createdUsrEmail, err := s.UsrEmailRepo.Create(ctx, s.DB, toCreateParent.ID, toCreateParent.LoginEmail)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		toCreateParent.ID = createdUsrEmail.UsrID
		toCreateParent.LegacyUser.ID = createdUsrEmail.UsrID
		parentIDs = append(parentIDs, toCreateParent.GetUID())
	}

	parentLegacyUsers := toCreateParents.Users()

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		parentsWithTags := map[entity.User][]entity.DomainTag{}
		// Check if student to assign exist
		findStudentResult, err := s.StudentRepo.FindStudentProfilesByIDs(ctx, tx, database.TextArray([]string{req.StudentId}))
		if err != nil {
			return errorx.ToStatusError(err)
		}
		if len(findStudentResult) == 0 {
			return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot assign parent with un-existing student in system: %s", req.StudentId))
		}
		student := findStudentResult[0]

		// Resolve parent data will be create
		if toCreateParents.Len() > 0 {
			// Valid emails for new parents to create

			emailExistingParents, err := s.UserRepo.GetByEmailInsensitiveCase(ctx, tx, toCreateParents.Emails())
			if err != nil {
				return errorx.ToStatusError(err)
			}
			if len(emailExistingParents) > 0 {
				return status.Error(codes.AlreadyExists, errcode.ErrUserEmailExists.Error())
			}

			// Valid phone numbers for new parents to create
			phoneNumberExistingParents, err := s.UserRepo.GetByPhone(ctx, tx, database.TextArray(toCreateParents.PhoneNumbers()))
			if err != nil {
				return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByPhone: %w", err).Error())
			}
			if len(phoneNumberExistingParents) > 0 {
				return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with phone number existing in system: %s", strings.Join(entity.ToUsers(phoneNumberExistingParents...).PhoneNumbers(), ", ")))
			}

			// Valid external user id for new parents to create
			externalUserIDs := []string{}
			for _, externalUserID := range parentLegacyUsers.ExternalUserIDs() {
				if externalUserID != "" {
					externalUserIDs = append(externalUserIDs, externalUserID)
				}
			}
			if len(externalUserIDs) > 0 {
				externalUserIDExistingParents, err := s.DomainUserRepo.GetByExternalUserIDs(ctx, tx, externalUserIDs)
				if err != nil {
					return status.Error(codes.Internal, fmt.Errorf("s.DomainUserRepo.GetByExternalUserIDs: %w", err).Error())
				}
				if len(externalUserIDExistingParents) > 0 {
					return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot create parent with external user id existing in system: %s", strings.Join(externalUserIDExistingParents.ExternalUserIDs(), ", ")))
				}
			}

			// Create multiple parents
			err = s.createParents(ctx, tx, toCreateParents)
			if err != nil {
				return errorx.ToStatusError(err)
			}

			if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
				return errorx.ToStatusError(err)
			}

			for _, parent := range toCreateParents {
				if err := s.StudentParentRepo.InsertParentAccessPathByStudentID(ctx, tx, parent.ID.String, req.StudentId); err != nil {
					return errorx.ToStatusError(err)
				}
			}

			// Upsert ignore conflict for relationships between student and parent
			err = s.assignParentsToStudent(ctx, tx, req.StudentId, toCreateParents...)
			if err != nil {
				return errorx.ToStatusError(err)
			}

			/*
			 * TODO:
			 *   if user group is stable, return error directly when user group not found
			 */

			// find student user group id if existed then assign this user group for created user
			parentUserGroup, err := s.UserGroupV2Repo.FindUserGroupByRoleName(ctx, tx, constant.RoleParent)
			if err == nil {
				if err := s.UserGroupsMemberRepo.AssignWithUserGroup(ctx, tx, parentLegacyUsers, parentUserGroup.UserGroupID); err != nil {
					return status.Error(codes.Internal, errors.Wrapf(err, "can not assign parent user group to user %s", student.GetUID()).Error())
				}
			} else {
				zapLogger.Warn(errors.Wrap(err, "can not find parent user group"))
			}

			// Protobuf responses
			parentProfilePBs = append(parentProfilePBs, parentsToParentPBsInCreateParentResponse(toCreateParents...)...)
		}

		// we convert data
		// from parent1 {tagid1,        tagid2,        tagid3}
		//   to parent1 {record_tagid1, record_tagid2, record_tagid3}
		for _, parentProfilePB := range parentProfilePBs {
			parentTags := []entity.DomainTag{}
			for _, tagID := range parentProfilePB.GetTagIds() {
				parentTags = append(parentTags, mapExistedTags[tagID])
			}
			userProfile := grpc.NewUserProfile(parentProfilePB.Parent.GetUserProfile())
			parentsWithTags[userProfile] = parentTags
		}

		taggedUsers, err := s.DomainTaggedUserRepo.GetByUserIDs(ctx, tx, parentIDs)
		if err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "DomainTaggedUserRepo.GetByUserIDs").Error())
		}

		if err := s.UpsertTaggedUsers(ctx, tx, parentsWithTags, taggedUsers); err != nil {
			return status.Error(codes.Internal, errors.Wrap(err, "UpsertTaggedUsers").Error())
		}

		// Import to identity platform
		tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, tx, strconv.FormatInt(resourcePath, 10))
		if err != nil {
			zapLogger.Error(
				"cannot get tenant id",
				zap.Error(err),
				zap.Int64("organizationID", resourcePath),
			)
			switch err {
			case pgx.ErrNoRows:
				return status.Error(codes.FailedPrecondition, errcode.TenantDoesNotExistErr{OrganizationID: strconv.FormatInt(resourcePath, 10)}.Error())
			default:
				return status.Error(codes.Internal, errcode.ErrCannotGetTenant.Error())
			}
		}

		err = s.CreateUsersInIdentityPlatform(ctx, tenantID, parentLegacyUsers, resourcePath)
		if err != nil {
			zapLogger.Error(
				"cannot create users on identity platform",
				zap.Error(err),
				zap.Int64("organizationID", resourcePath),
				zap.String("tenantID", tenantID),
				zap.Strings("emails", parentLegacyUsers.Limit(10).Emails()),
			)
			switch err {
			case user.ErrUserNotFound:
				return status.Error(codes.NotFound, errcode.NewUserNotFoundErr(parentLegacyUsers.IDs()...).Error())
			default:
				return status.Error(codes.Internal, err.Error())
			}
		}

		/*err = s.createUserInFirebase(ctx, firebaseAccounts, resourcePath)
		  if err != nil {
		  	return status.Error(codes.Internal, fmt.Errorf("s.CreateUserInFirebase: %w", err).Error())
		  }
		  for _, firebaseAccount := range firebaseAccounts {
		  	err := overrideUserPassword(ctx, s.FirebaseClient, firebaseAccount.ID.String, firebaseAccount.UserAdditionalInfo.Password)
		  	if err != nil {
		  		return status.Error(codes.Internal, fmt.Errorf("overrideUserPassword: %w", err).Error())
		  	}
		  }*/

		// Firebase accounts and nat streaming events
		userEvents := make([]*pb.EvtUser, 0, len(toCreateParents))
		userEvents = append(userEvents, newCreateParentEvents(int32(resourcePath), student, toCreateParents...)...)

		_ = s.publishAsyncUserEvent(ctx, userEvents...)

		return nil
	})
	if err != nil {
		return nil, err
	}
	response := &pb.CreateParentsAndAssignToStudentResponse{
		StudentId:      req.StudentId,
		ParentProfiles: parentProfilePBs,
	}
	return response, nil
}

// createParentProfileToParentDomain will convert parent profile from parent domain, support for the function entity.ValidParent
//
// entity.ValidParent will check if parent profile is valid or not with the following rules:
// 1. Username is required and valid format
// 2. Email is required and valid format
// 3. First name is required
// 4. Last name is required
func createParentProfileToParentDomain(profile *pb.CreateParentsAndAssignToStudentRequest_ParentProfile) entity.DomainParent {
	repoUser := &repository.User{
		ID: field.NewNullString(),
		// with trim space will auto set to null if empty
		UserNameAttr:  field.NewString(profile.Username).TrimSpace(),
		EmailAttr:     field.NewString(profile.Email).TrimSpace(),
		FirstNameAttr: field.NewNullString(),
		LastNameAttr:  field.NewNullString(),
		UserRoleAttr:  field.NewString(string(constant.UserRoleParent)),
	}
	if profile.UserNameFields != nil {
		repoUser.FirstNameAttr = field.NewString(profile.UserNameFields.FirstName)
		repoUser.LastNameAttr = field.NewString(profile.UserNameFields.LastName)
	} else {
		firstName, lastName := SplitNameToFirstNameAndLastName(profile.Name)
		repoUser.FirstNameAttr = field.NewString(firstName)
		repoUser.LastNameAttr = field.NewString(lastName)
	}
	return entity.ParentWillBeDelegated{
		DomainParentProfile: repoUser,
		HasUserID:           repoUser,
	}
}

// updateParentProfileToParentDomain will convert parent profile from parent domain, support for the function entity.ValidParent
//
// entity.ValidParent will check if parent profile is valid or not with the following rules:
// 1. Username is required and valid format
// 2. Email is required and valid format
// 3. First name is required
// 4. Last name is required
func updateParentProfileToParentDomain(profile *pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile) entity.DomainParent {
	repoUser := &repository.User{
		ID: field.NewString(profile.Id),
		// with trim space will auto set to null if empty
		UserNameAttr:  field.NewString(profile.Username).TrimSpace(),
		EmailAttr:     field.NewString(profile.Email).TrimSpace(),
		FirstNameAttr: field.NewNullString(),
		LastNameAttr:  field.NewNullString(),
	}
	if profile.UserNameFields != nil {
		repoUser.FirstNameAttr = field.NewString(profile.UserNameFields.FirstName)
		repoUser.LastNameAttr = field.NewString(profile.UserNameFields.LastName)
	}
	return entity.ParentWillBeDelegated{
		DomainParentProfile: repoUser,
		HasUserID:           repoUser,
	}
}

func (s *UserModifierService) createParents(ctx context.Context, db database.QueryExecer, parents entity.Parents) error {
	// UserGroups need to be inserted with new parent
	userGroups := make([]*entity.UserGroup, 0, len(parents))
	for _, parent := range parents {
		userGroup := &entity.UserGroup{}
		database.AllNullEntity(userGroup)
		err := multierr.Combine(
			userGroup.UserID.Set(parent.ID.String),
			userGroup.GroupID.Set(entity.UserGroupParent),
			userGroup.IsOrigin.Set(true),
			userGroup.Status.Set(entity.UserGroupStatusActive),
			userGroup.ResourcePath.Set(parent.ResourcePath),
		)
		if err != nil {
			return err
		}
		userGroups = append(userGroups, userGroup)
	}
	// Insert new parents
	err := s.UserRepo.CreateMultiple(ctx, db, parents.Users())
	if err != nil {
		return errorx.ToStatusError(err)
	}
	err = s.ParentRepo.CreateMultiple(ctx, db, parents)
	if err != nil {
		return errorx.ToStatusError(err)
	}
	err = s.UserGroupRepo.CreateMultiple(ctx, db, userGroups)
	if err != nil {
		return errorx.ToStatusError(err)
	}
	return nil
}

func toCreateParents(schoolID int32, req *pb.CreateParentsAndAssignToStudentRequest) (entity.Parents, entity.UserPhoneNumbers, error) {
	toCreateParents := entity.Parents{}
	userPhoneNumbers := entity.UserPhoneNumbers{}

	for _, parentProfile := range req.ParentProfiles {
		newParentID := idutil.ULIDNow()
		parentEnt, err := parentPbToParentEntity(schoolID, parentProfile, newParentID)
		if err != nil {
			return nil, nil, err
		}
		userPhoneNumberEnt, err := parentPbPhoneNumberToUserPhoneNumber(schoolID, parentProfile, newParentID)
		if err != nil {
			return nil, nil, err
		}
		userPhoneNumbers = append(userPhoneNumbers, userPhoneNumberEnt...)
		toCreateParents = append(toCreateParents, parentEnt)
	}
	return toCreateParents, userPhoneNumbers, nil
}

func parentPbToParentEntity(schoolID int32, parentPb *pb.CreateParentsAndAssignToStudentRequest_ParentProfile, parentID string) (*entity.Parent, error) {
	parentEnt := &entity.Parent{ParentAdditionalInfo: &entity.ParentAdditionalInfo{}}
	database.AllNullEntity(parentEnt)
	database.AllNullEntity(&parentEnt.LegacyUser)
	fistName, lastName := SplitNameToFirstNameAndLastName(parentPb.Name)
	err := multierr.Combine(
		parentEnt.LegacyUser.ID.Set(parentID),
		parentEnt.LegacyUser.Email.Set(parentPb.Email),
		parentEnt.LegacyUser.UserName.Set(parentPb.Username),
		parentEnt.LegacyUser.FullName.Set(parentPb.Name),
		parentEnt.LegacyUser.FirstName.Set(fistName),
		parentEnt.LegacyUser.LastName.Set(lastName),
		parentEnt.LegacyUser.Group.Set(entity.UserGroupParent),
		parentEnt.LegacyUser.Country.Set(parentPb.CountryCode.String()),
		// nolint:staticcheck //lint:ignore SA1019 Ignore the deprecation warnings until we completely remove this field
		parentEnt.LegacyUser.PhoneNumber.Set(parentPb.PhoneNumber),
		parentEnt.LegacyUser.Remarks.Set(parentPb.Remarks),
		parentEnt.LegacyUser.UserRole.Set(constant.UserRoleParent),
		parentEnt.ID.Set(parentID),
		parentEnt.SchoolID.Set(schoolID),
		parentEnt.ResourcePath.Set(fmt.Sprint(schoolID)),
		parentEnt.LegacyUser.LoginEmail.Set(parentPb.Email),
	)
	if parentPb.PhoneNumber == "" {
		if err := parentEnt.PhoneNumber.Set(nil); err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	if parentPb.UserNameFields != nil {
		if parentPb.UserNameFields.FirstName != "" && parentPb.UserNameFields.LastName != "" {
			if err := multierr.Combine(
				parentEnt.LegacyUser.FullName.Set(CombineFirstNameAndLastNameToFullName(parentPb.UserNameFields.FirstName, parentPb.UserNameFields.LastName)),
				parentEnt.LegacyUser.FirstName.Set(parentPb.UserNameFields.FirstName),
				parentEnt.LegacyUser.LastName.Set(parentPb.UserNameFields.LastName),
			); err != nil {
				return nil, err
			}
		}
		if err := multierr.Combine(
			parentEnt.LegacyUser.FullNamePhonetic.Set(CombineFirstNamePhoneticAndLastNamePhoneticToFullName(parentPb.UserNameFields.FirstNamePhonetic, parentPb.UserNameFields.LastNamePhonetic)),
			parentEnt.LegacyUser.FirstNamePhonetic.Set(parentPb.UserNameFields.FirstNamePhonetic),
			parentEnt.LegacyUser.LastNamePhonetic.Set(parentPb.UserNameFields.LastNamePhonetic),
		); err != nil {
			return nil, err
		}
	}

	if parentPb.ExternalUserId != "" {
		err := parentEnt.LegacyUser.ExternalUserID.Set(parentPb.ExternalUserId)
		if err != nil {
			return nil, err
		}
	}

	parentEnt.LegacyUser.UserAdditionalInfo.Password = parentPb.Password
	parentEnt.LegacyUser.UserAdditionalInfo.TagIDs = parentPb.TagIds
	parentEnt.ParentAdditionalInfo.Relationship = parentPb.Relationship.String()
	return parentEnt, nil
}

func parentPbPhoneNumberToUserPhoneNumber(schoolID int32, parentPb *pb.CreateParentsAndAssignToStudentRequest_ParentProfile, parentID string) (entity.UserPhoneNumbers, error) {
	userPhoneNumbers := entity.UserPhoneNumbers{}

	for i, userPhoneNumber := range parentPb.ParentPhoneNumbers {
		newParentPhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(newParentPhoneNumber)
		userPhoneNumberID := idutil.ULIDNow()
		if parentPb.ParentPhoneNumbers[i].GetPhoneNumberId() != "" {
			userPhoneNumberID = parentPb.ParentPhoneNumbers[i].GetPhoneNumberId()
		}
		if err := multierr.Combine(
			newParentPhoneNumber.ID.Set(userPhoneNumberID),
			newParentPhoneNumber.UserID.Set(parentID),
			newParentPhoneNumber.PhoneNumber.Set(userPhoneNumber.PhoneNumber),
			newParentPhoneNumber.PhoneNumberType.Set(userPhoneNumber.PhoneNumberType),
			newParentPhoneNumber.ResourcePath.Set(fmt.Sprint(schoolID)),
		); err != nil {
			return nil, fmt.Errorf("parentPbPhoneNumberToUserPhoneNumber multierr.Combine: %v", err)
		}
		userPhoneNumbers = append(userPhoneNumbers, newParentPhoneNumber)
	}
	return userPhoneNumbers, nil
}

func validUserNameFieldsCreateParentRequest(parentProfile *pb.CreateParentsAndAssignToStudentRequest_ParentProfile) error {
	if parentProfile.UserNameFields != nil {
		switch {
		case parentProfile.UserNameFields.FirstName == "":
			return errors.New("parent first_name cannot be empty")
		case parentProfile.UserNameFields.LastName == "":
			return errors.New("parent last_name cannot be empty")
		}
	} else if parentProfile.Name == "" {
		return errors.New("parent name cannot be empty")
	}
	return nil
}

func validateParentPhoneNumber(parentsPhoneNumber []*pb.ParentPhoneNumber) error {
	isHadPrimaryPhone := false
	primaryPhoneType := pb.ParentPhoneNumber_PARENT_PRIMARY_PHONE_NUMBER
	secondaryPhoneType := pb.ParentPhoneNumber_PARENT_SECONDARY_PHONE_NUMBER

	if len(parentsPhoneNumber) == 0 {
		return nil
	}

	// Sort before validate because need to compare 2 near value to check duplicate
	sort.Slice(parentsPhoneNumber, func(i, j int) bool {
		return parentsPhoneNumber[i].PhoneNumber < parentsPhoneNumber[j].PhoneNumber
	})

	for index, parentPhoneNumber := range parentsPhoneNumber {
		parentPhoneNumber.PhoneNumber = strings.TrimSpace(parentPhoneNumber.PhoneNumber)
		err := MatchingRegex(constant.PhoneNumberPattern, parentPhoneNumber.PhoneNumber)
		switch {
		case parentPhoneNumber.PhoneNumber == "":
			return nil
		case parentPhoneNumber.PhoneNumberType != primaryPhoneType &&
			parentPhoneNumber.PhoneNumberType != secondaryPhoneType:
			return errors.New("parent's phone number is wrong type")
		case err != nil:
			return err
		case index > 0 && parentPhoneNumber.PhoneNumber == parentsPhoneNumber[index-1].PhoneNumber:
			return errors.New("parent's phone number can not be duplicate")
		case isHadPrimaryPhone && parentPhoneNumber.PhoneNumberType == primaryPhoneType:
			return errors.New("parent only need one primary phone number")
		case !isHadPrimaryPhone && parentPhoneNumber.PhoneNumberType == primaryPhoneType:
			isHadPrimaryPhone = true
		}
	}
	return nil
}

func parentsToParentPBsInCreateParentResponse(parents ...*entity.Parent) []*pb.CreateParentsAndAssignToStudentResponse_ParentProfile {
	// Parent profile pb to response
	parentProfilePBs := make([]*pb.CreateParentsAndAssignToStudentResponse_ParentProfile, 0, len(parents))
	for _, parent := range parents {
		parentProfilePB := &pb.CreateParentsAndAssignToStudentResponse_ParentProfile{
			Parent: &pb.Parent{
				UserProfile: &pb.UserProfile{
					UserId:            parent.LegacyUser.ID.String,
					Email:             parent.LegacyUser.Email.String,
					Name:              parent.LegacyUser.GetName(),
					Avatar:            parent.LegacyUser.Avatar.String,
					Group:             pb.UserGroup(pb.UserGroup_value[parent.LegacyUser.Group.String]),
					PhoneNumber:       parent.LegacyUser.PhoneNumber.String,
					FacebookId:        parent.LegacyUser.FacebookID.String,
					GivenName:         parent.LegacyUser.GivenName.String,
					CountryCode:       cpb.Country(cpb.Country_value[parent.LegacyUser.Country.String]),
					FirstName:         parent.LegacyUser.FirstName.String,
					LastName:          parent.LegacyUser.LastName.String,
					FirstNamePhonetic: parent.LegacyUser.FirstNamePhonetic.String,
					LastNamePhonetic:  parent.LegacyUser.LastNamePhonetic.String,
					FullNamePhonetic:  parent.LegacyUser.FirstNamePhonetic.String,
				},
				SchoolId: parent.SchoolID.Int,
			},
			ParentPassword: parent.LegacyUser.UserAdditionalInfo.Password,
			Relationship:   pb.FamilyRelationship(pb.FamilyRelationship_value[parent.ParentAdditionalInfo.Relationship]),
			TagIds:         parent.TagIDs,
			UserNameFields: &pb.UserNameFields{
				FirstName:         parent.LegacyUser.FirstName.String,
				LastName:          parent.LegacyUser.LastName.String,
				FirstNamePhonetic: parent.LegacyUser.FirstNamePhonetic.String,
				LastNamePhonetic:  parent.LegacyUser.LastNamePhonetic.String,
			},
		}
		parentProfilePBs = append(parentProfilePBs, parentProfilePB)
	}
	return parentProfilePBs
}

func newCreateParentEvents(schoolID int32, student *entity.LegacyStudent, parents ...*entity.Parent) []*pb.EvtUser {
	createParentEvents := make([]*pb.EvtUser, 0, len(parents))
	for _, parent := range parents {
		createParentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_CreateParent_{
				CreateParent: &pb.EvtUser_CreateParent{
					StudentId:   student.ID.String,
					ParentId:    parent.ID.String,
					StudentName: student.GetName(),
					SchoolId:    strconv.FormatInt(int64(schoolID), 10),
				},
			},
		}
		createParentEvents = append(createParentEvents, createParentEvent)
	}
	return createParentEvents
}

func validCreateParentRequest(req *pb.CreateParentsAndAssignToStudentRequest) error {
	if req.StudentId == "" {
		return errors.New("student ID cannot be empty")
	}
	for _, parentProfile := range req.ParentProfiles {
		parentProfile.PhoneNumber = strings.TrimSpace(parentProfile.PhoneNumber)
		parentProfile.Email = strings.TrimSpace(parentProfile.Email)
		if pb.FamilyRelationship_name[int32(parentProfile.Relationship.Enum().Number())] == "" {
			return errors.New("parent relationship is not valid")
		}
		switch {
		case parentProfile.Email == "":
			return errors.New("parent email cannot be empty")
		case cpb.Country_name[int32(parentProfile.CountryCode.Enum().Number())] == "":
			return errors.New("parent country code is not valid")
		case parentProfile.Password == "":
			return errors.New("parent password cannot be empty")
		case len(parentProfile.Password) < firebase.MinimumPasswordLength:
			return errors.New("parent password length should be at least 6")
		}
		if err := validUserNameFieldsCreateParentRequest(parentProfile); err != nil {
			return err
		}
		if err := validateParentPhoneNumber(parentProfile.ParentPhoneNumbers); err != nil {
			return err
		}
	}
	return nil
}

func (s *UserModifierService) assignMultiParentsToMultiStudent(ctx context.Context, db database.QueryExecer, parentCSV *ParentCSV) error {
	studentParentEntities := make([]*entity.StudentParent, 0)

	for idx, studentID := range parentCSV.StudentIDs {
		studentParent := &entity.StudentParent{}
		database.AllNullEntity(studentParent)
		err := multierr.Combine(
			studentParent.StudentID.Set(studentID),
			studentParent.ParentID.Set(parentCSV.Parent.ID),
			studentParent.Relationship.Set(parentCSV.Relationship[idx]),
		)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
		}
		studentParentEntities = append(studentParentEntities, studentParent)
	}

	err := s.StudentParentRepo.Upsert(ctx, db, studentParentEntities)
	if err != nil {
		return errorx.ToStatusError(err)
	}
	if err := s.StudentParentRepo.UpsertParentAccessPathByStudentIDs(ctx, db, parentCSV.StudentIDs); err != nil {
		return err
	}
	return nil
}

func (s *UserModifierService) assignParentsToStudent(ctx context.Context, db database.QueryExecer, studentID string, parents ...*entity.Parent) error {
	studentParentEntities := make([]*entity.StudentParent, 0, len(parents))
	for _, parent := range parents {
		studentParent := &entity.StudentParent{}
		database.AllNullEntity(studentParent)
		err := multierr.Combine(
			studentParent.StudentID.Set(studentID),
			studentParent.ParentID.Set(parent.ID),
			studentParent.Relationship.Set(parent.ParentAdditionalInfo.Relationship),
		)
		if err != nil {
			return status.Error(codes.Internal, fmt.Errorf("multierr.Combine: %w", err).Error())
		}
		studentParentEntities = append(studentParentEntities, studentParent)
	}
	// Insert relationship between student and parent
	err := s.StudentParentRepo.Upsert(ctx, db, studentParentEntities)
	if err != nil {
		return errorx.ToStatusError(err)
	}
	return nil
}
