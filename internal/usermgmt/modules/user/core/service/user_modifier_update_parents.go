package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/auth/user"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/errorx"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/adapter/postgres/repository"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
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

func validateUpdateParentsAndFamilyRelationshipRequest(req *pb.UpdateParentsAndFamilyRelationshipRequest) error {
	if req.StudentId == "" {
		return errors.New("student ID cannot be empty")
	}

	for _, parentProfile := range req.ParentProfiles {
		parentProfile.Email = strings.TrimSpace(parentProfile.Email)

		if parentProfile.Id == "" {
			return errors.New("parent id cannot be empty")
		}
		if parentProfile.Email == "" {
			return errors.New("parent email cannot be empty")
		}

		if parentProfile.UserNameFields != nil {
			switch {
			case parentProfile.UserNameFields.FirstName == "":
				return errcode.ErrUserFirstNameIsEmpty
			case parentProfile.UserNameFields.LastName == "":
				return errcode.ErrUserLastNameIsEmpty
			}
		}
		if pb.FamilyRelationship_name[int32(parentProfile.Relationship.Enum().Number())] == "" {
			return errors.New("parent relationship is not valid")
		}
		if err := validateParentPhoneNumber(parentProfile.ParentPhoneNumbers); err != nil {
			return err
		}
	}

	return nil
}

func (s *UserModifierService) UpdateParentsAndFamilyRelationship(ctx context.Context, req *pb.UpdateParentsAndFamilyRelationshipRequest) (*pb.UpdateParentsAndFamilyRelationshipResponse, error) {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	parentProfilePBs := make([]*pb.UpdateParentsAndFamilyRelationshipResponse_ParentProfile, 0, len(req.ParentProfiles))

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resourcePath := int64(organization.SchoolID().Int32())

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
	if err := validateUpdateParentsAndFamilyRelationshipRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	parentDomains := make(entity.Users, 0, len(req.ParentProfiles))
	for _, parentProfile := range req.ParentProfiles {
		if !isEnableUsername {
			parentProfile.Username = parentProfile.Email
		}

		parentProfile.Username = strings.TrimSpace(strings.ToLower(parentProfile.Username))

		parent := updateParentProfileToParentDomain(parentProfile)
		if err := entity.ValidParent(parent, isEnableUsername); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		parentDomains = append(parentDomains, parent)
	}

	if isEnableUsername {
		if err := ValidateUserNamesExistedInSystem(ctx, s.DomainUserRepo, s.DB, parentDomains); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
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

	err = database.ExecInTx(ctx, s.DB, func(ctx context.Context, tx pgx.Tx) error {
		if len(req.ParentProfiles) > 0 {
			parentIDs := []string{}
			userPhoneNumbers := []*entity.UserPhoneNumber{}

			// Check if student to assign exist
			findStudentResult, err := s.StudentRepo.FindStudentProfilesByIDs(ctx, tx, database.TextArray([]string{req.StudentId}))
			if err != nil {
				return errorx.ToStatusError(err)
			}
			if len(findStudentResult) == 0 {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot update parents associated with un-existing student in system: %s", req.StudentId))
			}
			student := findStudentResult[0]

			for _, profile := range req.ParentProfiles {
				parentIDs = append(parentIDs, profile.Id)
			}
			// Valid parent will be assigned to student
			existingParents, err := s.ParentRepo.GetByIds(ctx, tx, database.TextArray(parentIDs))
			if err != nil {
				return errorx.ToStatusError(errors.Wrap(err, "s.ParentRepo.GetByIds"))
			}
			if len(existingParents) != len(req.ParentProfiles) {
				return status.Error(codes.InvalidArgument, "a parent ID in request does not exist in db")
			}

			// Convert existing parents slice to map with id as key
			existingParentsMap := make(map[string]*entity.Parent, len(existingParents))
			for _, existingParent := range existingParents {
				existingParentsMap[existingParent.ID.String] = existingParent
			}

			parentsWithTags := map[entity.User][]entity.DomainTag{}
			// Check parents will be assigned to student are exists or not
			for _, profile := range req.ParentProfiles {
				existingParent := existingParentsMap[profile.Id]

				if existingParent == nil {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot assign non-existing parent to student: %s", profile.Id))
				}

				existingParent.UserName = database.Text(profile.Username)

				if existingParent.SchoolID.Int != student.SchoolID.Int {
					return status.Error(codes.InvalidArgument, fmt.Sprintf("parent %s not same school with student", profile.Id))
				}

				if profile.ExternalUserId != "" && (existingParent.ExternalUserID.String != profile.ExternalUserId) {
					if existingParent.ExternalUserID.String != "" {
						return status.Error(codes.InvalidArgument, fmt.Sprintf("cannot re-update external_user_id: %s", profile.Id))
					}

					// Valid external user id for new parents to create
					externalUserIDExistingParents, err := s.DomainUserRepo.GetByExternalUserIDs(ctx, tx, []string{profile.ExternalUserId})
					if err != nil {
						return status.Error(codes.Internal, fmt.Errorf("s.DomainUserRepo.GetByExternalUserIDs: %w", err).Error())
					}
					if len(externalUserIDExistingParents) > 0 {
						return status.Error(codes.AlreadyExists, fmt.Sprintf("cannot update parent with external user id existing in system: %s", strings.Join(externalUserIDExistingParents.ExternalUserIDs(), ", ")))
					}

					existingParent.ExternalUserID = database.Text(profile.ExternalUserId)
				}

				// If parent email edited
				if existingParent.Email.String != profile.Email {
					/*//Use centralized func to update email
					repo := userservice.RepoForDomainUser{
						DomainUserRepo:     s.DomainUserRepo,
						DomainUsrEmailRepo: s.DomainUsrEmailRepo,
						UserRepo:           s.UserRepo,
						OrganizationRepo:   s.OrganizationRepo,
					}
					userservice.UpdateEmail(ctx, tx, s.TenantManager, s.DB, repo, )*/
					// Check if edited email already exists
					emailExistingParents, err := s.UserRepo.GetByEmailInsensitiveCase(ctx, tx, []string{profile.Email})
					if err != nil {
						return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.GetByEmailInsensitiveCase: %w", err).Error())
					}
					if len(emailExistingParents) > 0 {
						return status.Error(codes.AlreadyExists, errcode.ErrUserEmailExists.Error())
					}
					// Update new email in DB
					existingParent.Email = database.Text(profile.Email)
					if !isEnableUsername {
						existingParent.LoginEmail = database.Text(profile.Email)
						err := s.UsrEmailRepo.UpdateEmail(ctx, tx, existingParent.ID, database.Text(strconv.Itoa(int(resourcePath))), existingParent.LegacyUser.LoginEmail)
						switch err {
						case nil:
							break
						case repository.ErrNoRowAffected:
							// it's ok to have no row affected
							break
						default:
							return status.Error(codes.Internal, errors.Wrap(err, "s.UserModifierService.UsrEmailRepo.UpdateEmail").Error())
						}
					}
					err = s.UserRepo.UpdateEmail(ctx, tx, &existingParent.LegacyUser)
					if err != nil {
						return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.UpdateEmail: %w", err).Error())
					}
				}
				// set User to update remarks
				if err := existingParent.Remarks.Set(profile.Remarks); err != nil {
					return err
				}
				existingParent.ParentAdditionalInfo = &entity.ParentAdditionalInfo{Relationship: profile.Relationship.String()}
				if profile.UserNameFields != nil {
					if err := multierr.Combine(
						existingParent.FirstName.Set(profile.UserNameFields.FirstName),
						existingParent.LastName.Set(profile.UserNameFields.LastName),
						existingParent.FullName.Set(CombineFirstNameAndLastNameToFullName(profile.UserNameFields.FirstName, profile.UserNameFields.LastName)),
						existingParent.FirstNamePhonetic.Set(profile.UserNameFields.FirstNamePhonetic),
						existingParent.LastNamePhonetic.Set(profile.UserNameFields.LastNamePhonetic),
						existingParent.FullNamePhonetic.Set(CombineFirstNameAndLastNameToFullName(profile.UserNameFields.FirstNamePhonetic, profile.UserNameFields.LastNamePhonetic)),
					); err != nil {
						return err
					}
				}

				if err = s.UserRepo.UpdateProfileV1(ctx, tx, &existingParent.LegacyUser); err != nil {
					return status.Error(codes.Internal, fmt.Errorf("s.UserRepo.UpdateProfileV1: %w", err).Error())
				}
				userPhoneNumbersEnt, err := pbParentPhoneNumberToUserPhoneNumbers(profile, fmt.Sprint(resourcePath))
				if err != nil {
					return err
				}
				userPhoneNumbers = append(userPhoneNumbers, userPhoneNumbersEnt...)

				// we convert data
				// from parent1 {tagid1,        tagid2,        tagid3}
				//   to parent1 {record_tagid1, record_tagid2, record_tagid3}
				parentTags := []entity.DomainTag{}
				for _, tagID := range profile.GetTagIds() {
					parentTags = append(parentTags, mapExistedTags[tagID])
				}
				parentProfile := grpc.NewUserProfileWithID(profile.GetId())
				parentsWithTags[parentProfile] = parentTags
			}

			// Upsert ignore conflict for relationships between student and parent
			err = s.assignParentsToStudent(ctx, tx, req.StudentId, existingParents...)
			if err != nil {
				return errorx.ToStatusError(err)
			}

			if err := s.StudentParentRepo.UpsertParentAccessPathByStudentIDs(ctx, tx, []string{req.StudentId}); err != nil {
				return err
			}

			taggedUsers, err := s.DomainTaggedUserRepo.GetByUserIDs(ctx, tx, parentIDs)
			if err != nil {
				return status.Error(codes.Internal, errors.Wrap(err, "DomainTaggedUserRepo.GetByUserIDs").Error())
			}

			if err := s.UpsertTaggedUsers(ctx, tx, parentsWithTags, taggedUsers); err != nil {
				return status.Error(codes.Internal, errors.Wrap(err, "UpsertTaggedUsers").Error())
			}

			if !isEnableUsername {
				// Import to identity platform
				tenantID, err := s.OrganizationRepo.GetTenantIDByOrgID(ctx, s.DB, strconv.FormatInt(resourcePath, 10))
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
				// Update new parents emails in firebase
				for _, profile := range req.ParentProfiles {
					err := s.UpdateUserEmailInIdentityPlatform(ctx, tenantID, profile.Id, profile.Email)
					if err != nil {
						zapLogger.Error(
							"cannot create users on identity platform",
							zap.Error(err),
							zap.Int64("organizationID", resourcePath),
							zap.String("tenantID", tenantID),
							zap.String("email", profile.Email),
						)
						switch err {
						case user.ErrUserNotFound:
							return status.Error(codes.NotFound, errcode.NewUserNotFoundErr(profile.Id).Error())
						default:
							return status.Error(codes.Internal, err.Error())
						}
					}
				}
			}
			// Update and Insert userPhoneNumber
			if err := s.UserPhoneNumberRepo.Upsert(ctx, tx, userPhoneNumbers); err != nil {
				return errorx.ToStatusError(err)
			}
			// Protobuf responses
			parentProfilePBs = append(parentProfilePBs, parentsToParentPBsInUpdateParentsAndFamilyRelationshipResponse(existingParents...)...)
			// publish parent assigned to student events
			userEvents := make([]*pb.EvtUser, 0, len(req.ParentProfiles))
			userEvents = append(userEvents, newParentAssignedToStudentEvent(student.ID.String, parentIDs)...)
			if err := s.publishUserEvent(ctx, constants.SubjectUserUpdated, userEvents...); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	response := &pb.UpdateParentsAndFamilyRelationshipResponse{
		StudentId:      req.StudentId,
		ParentProfiles: parentProfilePBs,
	}
	return response, nil
}

func pbParentPhoneNumberToUserPhoneNumbers(profile *pb.UpdateParentsAndFamilyRelationshipRequest_ParentProfile, resourcePath string) (entity.UserPhoneNumbers, error) {
	userPhoneNumbers := make(entity.UserPhoneNumbers, 0, len(profile.ParentPhoneNumbers))

	for _, userPhoneNumber := range profile.ParentPhoneNumbers {
		newParentPhoneNumber := &entity.UserPhoneNumber{}
		database.AllNullEntity(newParentPhoneNumber)
		userPhoneNumberID := idutil.ULIDNow()

		if userPhoneNumber.GetPhoneNumberId() != "" {
			userPhoneNumberID = userPhoneNumber.GetPhoneNumberId()
		}
		phoneNumberType, ok := pb.ParentPhoneNumber_ParentPhoneNumberType_name[int32(userPhoneNumber.PhoneNumberType)]
		if !ok {
			return nil, errors.New("don't have that type of parent phone")
		}

		if err := multierr.Combine(
			newParentPhoneNumber.ID.Set(userPhoneNumberID),
			newParentPhoneNumber.UserID.Set(profile.Id),
			newParentPhoneNumber.PhoneNumber.Set(userPhoneNumber.PhoneNumber),
			newParentPhoneNumber.PhoneNumberType.Set(phoneNumberType),
			newParentPhoneNumber.ResourcePath.Set(resourcePath),
		); err != nil {
			return nil, fmt.Errorf("pbParentPhoneNumberToUserPhoneNumbers multierr.Combine: %v", err)
		}

		userPhoneNumbers = append(userPhoneNumbers, newParentPhoneNumber)
	}
	return userPhoneNumbers, nil
}

func parentsToParentPBsInUpdateParentsAndFamilyRelationshipResponse(parents ...*entity.Parent) []*pb.UpdateParentsAndFamilyRelationshipResponse_ParentProfile {
	// Parent profile pb to response
	parentProfilePBs := make([]*pb.UpdateParentsAndFamilyRelationshipResponse_ParentProfile, 0, len(parents))
	for _, parent := range parents {
		parentProfilePB := &pb.UpdateParentsAndFamilyRelationshipResponse_ParentProfile{
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
					FullNamePhonetic:  parent.LegacyUser.FullNamePhonetic.String,
				},
				SchoolId: parent.SchoolID.Int,
			},
			Relationship: pb.FamilyRelationship(pb.FamilyRelationship_value[parent.ParentAdditionalInfo.Relationship]),
		}
		parentProfilePBs = append(parentProfilePBs, parentProfilePB)
	}
	return parentProfilePBs
}

func newParentAssignedToStudentEvent(studentID string, parentIDs []string) []*pb.EvtUser {
	parentAssignedToStudentEvents := make([]*pb.EvtUser, 0, len(parentIDs))

	for _, parentID := range parentIDs {
		parentRemovedFromStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_ParentAssignedToStudent_{
				ParentAssignedToStudent: &pb.EvtUser_ParentAssignedToStudent{
					StudentId: studentID,
					ParentId:  parentID,
				},
			},
		}

		parentAssignedToStudentEvents = append(parentAssignedToStudentEvents, parentRemovedFromStudentEvent)
	}

	return parentAssignedToStudentEvents
}
