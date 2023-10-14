package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/constants"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/golibs/unleashclient"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DomainParent struct {
	DB                 libdatabase.Ext
	JSM                nats.JetStreamManagement
	FirebaseAuthClient multitenant.TenantClient
	TenantManager      multitenant.TenantManager
	UnleashClient      unleashclient.ClientInstance
	Env                string

	OrganizationRepo OrganizationRepo

	UserRepo     UserRepo
	UsrEmailRepo interface {
		CreateMultiple(ctx context.Context, db libdatabase.QueryExecer, users entity.Users) (valueobj.HasUserIDs, error)
		UpdateEmail(ctx context.Context, db libdatabase.QueryExecer, user entity.User) error
	}
	UserGroupRepo interface {
		FindUserGroupByRoleName(ctx context.Context, db libdatabase.QueryExecer, roleName string) (entity.DomainUserGroup, error)
	}
	ParentRepo interface {
		GetUsersByExternalUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (entity.Users, error)
		UpsertMultiple(ctx context.Context, db libdatabase.QueryExecer, isEnableUsername bool, parentsToCreate ...aggregate.DomainParent) error
	}
	TagRepo interface {
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) (entity.DomainTags, error)
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainTags, error)
	}
	TaggedUserRepo interface {
		UpsertBatch(ctx context.Context, db libdatabase.QueryExecer, taggedUsers ...entity.DomainTaggedUser) error
		SoftDeleteByUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) error
	}
	UserAccessPathRepo interface {
		GetByUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (entity.DomainUserAccessPaths, error)
	}
	StudentParentRepo interface {
		SoftDeleteByParentIDs(ctx context.Context, db libdatabase.QueryExecer, parentIDs []string) error
		GetByParentIDs(ctx context.Context, db libdatabase.QueryExecer, parentIDs []string) (entity.DomainStudentParentRelationships, error)
	}
	InternalConfigurationRepo interface {
		GetByKey(ctx context.Context, db libdatabase.QueryExecer, configKey string) (entity.DomainConfiguration, error)
	}
	AssignParentToStudentsManager AssignParentToStudentsManager
	AuthUserUpserter              AuthUserUpserter
	UserPhoneNumberRepo           UserPhoneNumberRepo
}

func (service *DomainParent) UpsertMultiple(ctx context.Context, option unleash.DomainParentFeatureOption, parentsToUpsert ...aggregate.DomainParent) ([]aggregate.DomainParent, error) {
	parentsToCreate, parentsToUpdate, parentsToUpsert, err := service.DomainParentsToUpsert(ctx, service.DB, option.EnableUsername, parentsToUpsert...)
	if err != nil {
		return nil, err
	}
	if err := libdatabase.ExecInTx(ctx, service.DB, func(ctx context.Context, tx pgx.Tx) error {
		parentsToUpsert, err = service.UpsertMultipleParentsInTx(ctx, tx, parentsToCreate, parentsToUpdate, parentsToUpsert, option)
		return err
	}); err != nil {
		return nil, err
	}

	return parentsToUpsert, nil
}

func (service *DomainParent) UpsertMultipleWithChildren(ctx context.Context, option unleash.DomainParentFeatureOption, aggregateParents ...aggregate.DomainParentWithChildren) ([]aggregate.DomainParent, error) {
	zapLogger := ctxzap.Extract(ctx)
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "OrganizationFromContext")
	}

	parents := make([]aggregate.DomainParent, len(aggregateParents))
	for i, parent := range aggregateParents {
		parents[i] = parent.DomainParent
	}

	parentsToCreate, parentsToUpdate, parentsToUpsert, err := service.DomainParentsToUpsert(ctx, service.DB, option.EnableUsername, parents...)
	if err != nil {
		return nil, err
	}

	if err := libdatabase.ExecInTx(ctx, service.DB, func(ctx context.Context, tx pgx.Tx) error {
		parentsToUpsert, err = service.UpsertMultipleParentsInTx(ctx, tx, parentsToCreate, parentsToUpdate, parentsToUpsert, option)
		if err != nil {
			zapLogger.Error("service.DomainParentService.UpsertMultipleParentsInTx", zap.Error(err))
			return err
		}

		parentToUpdateIDs := aggregate.DomainParents(parentsToUpdate).ParentIDs()
		studentParentsToSoftDelete, err := service.StudentParentRepo.GetByParentIDs(ctx, tx, parentToUpdateIDs)
		if err != nil {
			zapLogger.Error("service.StudentParentRepo.GetByParentIDs", zap.Error(err))
			return err
		}
		for _, studentParent := range studentParentsToSoftDelete {
			if err := publishRemovedParentFromStudentEvent(ctx, service.JSM, studentParent, studentParent); err != nil {
				zapLogger.Error("publishRemovedParentFromStudentEvent", zap.Error(err))
				return err
			}
		}
		if err := service.StudentParentRepo.SoftDeleteByParentIDs(ctx, tx, parentToUpdateIDs); err != nil {
			zapLogger.Error("service.StudentParentRepo.SoftDeleteByParentIDs", zap.Error(err))
			return err
		}
		for idx, parent := range parentsToUpsert {
			parentWithChildren := aggregateParents[idx]
			for _, studentParentRelationship := range parentWithChildren.Children {
				err := service.AssignParentToStudentsManager(ctx, tx, organization, studentParentRelationship.Relationship(), parent, studentParentRelationship)
				if err != nil {
					zapLogger.Error("service.AssignParentToStudentsManager", zap.Error(err))
					return err
				}
				students, err := service.UserRepo.GetByIDs(ctx, tx, []string{studentParentRelationship.StudentID().String()})
				if err != nil {
					zapLogger.Error("service.UserRepo.GetByIDs", zap.Error(err))
					return err
				}
				for _, student := range students {
					if err := publishUpsertParentEvent(ctx, service.JSM, student, []aggregate.DomainParent{parent}); err != nil {
						return fmt.Errorf("publishUpsertParentEvent: %s", err.Error())
					}
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return parentsToUpsert, nil
}

func (service *DomainParent) UpsertMultipleParentsInTx(ctx context.Context, tx libdatabase.Tx, parentsToCreate, parentsToUpdate, parentsToUpsert aggregate.DomainParents, option unleash.DomainParentFeatureOption) ([]aggregate.DomainParent, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "OrganizationFromContext")
	}

	for _, parent := range parentsToUpdate {
		err := service.UsrEmailRepo.UpdateEmail(ctx, tx, parent)
		if err != nil {
			return nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.UsrEmailRepo.UpdateEmail"),
			}
		}
	}

	if err := service.ParentRepo.UpsertMultiple(ctx, tx, option.EnableUsername, parentsToUpsert...); err != nil {
		return nil, err
	}

	var userPhoneNumbers entity.DomainUserPhoneNumbers
	for _, parent := range parentsToUpsert {
		for _, userPhoneNumber := range parent.UserPhoneNumbers {
			userPhoneNumbers = append(userPhoneNumbers, entity.UserPhoneNumberWillBeDelegated{
				HasUserID:                parent,
				HasOrganizationID:        parent,
				UserPhoneNumberAttribute: userPhoneNumber,
			})
		}
	}
	if err := upsertUserPhoneNumbers(ctx, tx, service.UserPhoneNumberRepo, parentsToUpsert.Users(), userPhoneNumbers); err != nil {
		return nil, err
	}

	if err := service.upsertTaggedUsers(ctx, tx, parentsToUpsert...); err != nil {
		return nil, err
	}

	// Upsert users in auth platform
	// AuthUserUpserter must be used with service.DB, not current tx
	// Because users are updated in current tx before AuthUserUpserter invoke,
	// but it needs to query the user data before updating to validate
	if _, err := service.AuthUserUpserter(ctx, service.DB, organization, parentsToCreate.Users(), parentsToUpdate.Users(), option.DomainUserFeatureOption); err != nil {
		return nil, err
	}

	return parentsToUpsert, nil
}

func (service *DomainParent) DomainParentsToUpsert(ctx context.Context, db libdatabase.Ext, isEnableUsername bool, parentsToUpsert ...aggregate.DomainParent) ([]aggregate.DomainParent, []aggregate.DomainParent, []aggregate.DomainParent, error) {
	if err := service.validationUpsertParent(ctx, db, isEnableUsername, parentsToUpsert...); err != nil {
		return nil, nil, nil, err
	}

	parentsToCreate, parentsToUpdate, err := service.generateUserIDs(ctx, db, isEnableUsername, parentsToUpsert...)
	if err != nil {
		return nil, nil, nil, err
	}
	if err := service.validateUpdateSystemAndExternalUserID(ctx, db, parentsToUpdate); err != nil {
		return nil, nil, nil, err
	}

	parentsToUpsert, err = service.assignAggregate(ctx, db, parentsToUpsert...)
	if err != nil {
		return nil, nil, nil, err
	}

	return parentsToCreate, parentsToUpdate, parentsToUpsert, nil
}

func validateParentDuplicatedFields(parents ...aggregate.DomainParent) error {
	users := entity.Users{}
	for _, parent := range parents {
		users = append(users, parent)
	}

	return ValidateUserDuplicatedFields(users)
}

func (service *DomainParent) validateTags(ctx context.Context, parents ...aggregate.DomainParent) error {
	for i, parent := range parents {
		if len(parent.TaggedUsers) == 0 {
			continue
		}

		tags, err := service.TagRepo.GetByIDs(ctx, service.DB, parent.TaggedUsers.TagIDs())
		if err != nil {
			return err
		}

		if len(parent.TaggedUsers) != len(tags) {
			return entity.InvalidFieldError{
				EntityName: entity.ParentEntity,
				Index:      i,
				FieldName:  entity.ParentTagsField,
				Reason:     entity.Invalid,
			}
		}

		for _, tag := range tags {
			if !golibs.InArrayString(tag.TagType().String(), entity.ParentTags) {
				return entity.InvalidFieldError{
					EntityName: entity.ParentEntity,
					Index:      i,
					FieldName:  entity.ParentTagsField,
					Reason:     entity.Invalid,
				}
			}
		}
	}
	return nil
}

func (service *DomainParent) getCurrentUser(ctx context.Context) (entity.User, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	users, err := service.UserRepo.GetByIDs(ctx, service.DB, []string{currentUserID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, entity.NotFoundError{
			EntityName: entity.UserEntity,
			FieldName:  string(entity.UserFieldUserID),
			FieldValue: currentUserID,
		}
	}
	return users[0], nil
}

func (service *DomainParent) generateUserIDs(ctx context.Context, db libdatabase.Ext, isEnableUsername bool, parents ...aggregate.DomainParent) ([]aggregate.DomainParent, []aggregate.DomainParent, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, nil, entity.InternalError{
			RawErr: errors.Wrap(err, "OrganizationFromContext"),
		}
	}

	parentsToCreate := []aggregate.DomainParent{}
	parentsToUpdate := []aggregate.DomainParent{}

	usersToGenID := entity.Users{}

	for _, parent := range parents {
		if parent.UserID().String() != "" {
			parentsToUpdate = append(parentsToUpdate, parent)
		} else {
			parentsToCreate = append(parentsToCreate, parent)
			uid := idutil.ULIDNow()
			var userProfileLoginEmail valueobj.HasLoginEmail = parent.DomainParent
			if isEnableUsername {
				userProfileLoginEmail = &entity.UserProfileLoginEmailDelegate{
					Email: uid + constant.LoginEmailPostfix,
				}
			}
			usersToGenID = append(usersToGenID, entity.ParentWillBeDelegated{
				DomainParentProfile: parent.DomainParent,
				HasOrganizationID:   organization,
				HasUserID: &valueobj.RandomHasUserID{
					RandomUserID: field.NewString(uid),
				},
				HasLoginEmail: userProfileLoginEmail,
			})
		}
	}
	userIDs, err := service.UsrEmailRepo.CreateMultiple(ctx, db, usersToGenID)
	if err != nil {
		return nil, nil, err
	}
	for i := range parentsToCreate {
		var userProfileLoginEmail valueobj.HasLoginEmail = parentsToCreate[i].DomainParent
		if isEnableUsername {
			userProfileLoginEmail = &entity.UserProfileLoginEmailDelegate{
				Email: userIDs[i].UserID().String() + constant.LoginEmailPostfix,
			}
		}
		parentsToCreate[i].DomainParent = entity.ParentWillBeDelegated{
			DomainParentProfile: parentsToCreate[i].DomainParent,
			HasUserID:           userIDs[i],
			HasLoginEmail:       userProfileLoginEmail,
		}
		for j := range parents {
			if parentsToCreate[i].Email() == parents[j].Email() {
				parents[j].DomainParent = parentsToCreate[i].DomainParent
			}

			parentToCreateEmail := parentsToCreate[i].Email()
			parentToUpsertEmail := parents[j].Email()
			if isEnableUsername {
				parentToCreateEmail = parentsToCreate[i].UserName()
				parentToUpsertEmail = parents[j].UserName()
			}
			if parentToCreateEmail == parentToUpsertEmail {
				parents[j].DomainParent = parentsToCreate[i].DomainParent
			}
		}
	}
	return parentsToCreate, parentsToUpdate, nil
}

func (service *DomainParent) setUserGroupMembers(ctx context.Context, db libdatabase.Ext, organization valueobj.HasOrganizationID, parentsToCreate ...aggregate.DomainParent) error {
	parentUserGroup, err := service.UserGroupRepo.FindUserGroupByRoleName(ctx, db, constant.RoleParent)
	if err != nil {
		return err
	}

	for i, parent := range parentsToCreate {
		parentsToCreate[i].UserGroupMembers = append(parentsToCreate[i].UserGroupMembers, entity.UserGroupMemberWillBeDelegated{
			HasUserGroupID:    parentUserGroup,
			HasUserID:         parent,
			HasOrganizationID: organization,
		})
	}

	return nil
}

func (service *DomainParent) validationUpsertParent(ctx context.Context, db libdatabase.Ext, isEnableUsername bool, parentsToUpsert ...aggregate.DomainParent) error {
	for i := range parentsToUpsert {
		if err := entity.ValidParent(parentsToUpsert[i], isEnableUsername); err != nil {
			return err
		}
	}

	if err := validateParentDuplicatedFields(parentsToUpsert...); err != nil {
		return err
	}

	if err := service.validateExternUserIDUsedByOtherRole(ctx, parentsToUpsert); err != nil {
		return err
	}

	if err := validateParentPhoneNumbers(parentsToUpsert...); err != nil {
		return err
	}

	users := entity.Users{}
	for _, user := range parentsToUpsert {
		users = append(users, user)
	}
	if err := ValidateUserEmailsExistedInSystem(ctx, service.UserRepo, db, users); err != nil {
		return err
	}

	if err := ValidateUserNamesExistedInSystem(ctx, service.UserRepo, db, users); err != nil {
		return err
	}

	return service.validateTags(ctx, parentsToUpsert...)
}

func (service *DomainParent) assignAggregate(ctx context.Context, db libdatabase.Ext, parentsToUpsert ...aggregate.DomainParent) ([]aggregate.DomainParent, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, entity.InternalError{
			RawErr: errors.Wrap(err, "OrganizationFromContext"),
		}
	}
	currentUser, err := service.getCurrentUser(ctx)
	if err != nil {
		return nil, entity.InternalError{
			RawErr: errors.Wrap(err, "getCurrentUser"),
		}
	}

	for i := range parentsToUpsert {
		parentsToUpsert[i].DomainParent = &entity.ParentWillBeDelegated{
			DomainParentProfile: parentsToUpsert[i].DomainParent,
			HasOrganizationID:   organization,
			HasSchoolID:         organization,
			HasUserID:           parentsToUpsert[i].DomainParent,
			HasCountry:          currentUser,
			HasLoginEmail:       parentsToUpsert[i].DomainParent,
		}
		parentsToUpsert[i].LegacyUserGroups = entity.LegacyUserGroups{entity.DelegateToLegacyUserGroup(&entity.ParentLegacyUserGroup{}, organization, parentsToUpsert[i])}
	}

	if err := service.setUserGroupMembers(ctx, db, organization, parentsToUpsert...); err != nil {
		return nil, entity.InternalError{
			RawErr: errors.Wrap(err, "service.setUserGroupMembers"),
		}
	}
	service.setUserAccessPaths(organization, parentsToUpsert...)

	return parentsToUpsert, nil
}

func validateParentPhoneNumbers(parents ...aggregate.DomainParent) error {
	for _, parent := range parents {
		if err := ValidateUserPhoneNumbers(parent.UserPhoneNumbers, parent.IndexAttr); err != nil {
			return err
		}
	}

	return nil
}

func (service *DomainParent) upsertTaggedUsers(ctx context.Context, db libdatabase.QueryExecer, parents ...aggregate.DomainParent) error {
	var taggedUsers entity.DomainTaggedUsers
	userIDs := []string{}
	for _, parent := range parents {
		userIDs = append(userIDs, parent.UserID().String())
		for _, taggedUser := range parent.TaggedUsers {
			taggedUsers = append(taggedUsers, &entity.TaggedUserWillBeDelegated{
				HasTagID:          taggedUser,
				HasUserID:         parent,
				HasOrganizationID: parent,
			})
		}
	}

	if err := service.TaggedUserRepo.SoftDeleteByUserIDs(ctx, db, userIDs); err != nil {
		return entity.InternalError{
			RawErr: errors.Wrap(err, "service.TaggedUserRepo.SoftDeleteByUserIDs"),
		}
	}

	if len(taggedUsers) == 0 {
		return nil
	}

	if err := service.TaggedUserRepo.UpsertBatch(ctx, db, taggedUsers...); err != nil {
		return entity.InternalError{
			RawErr: errors.Wrap(err, "service.TaggedUserRepo.UpsertMultiple"),
		}
	}

	return nil
}

func (service *DomainParent) setUserAccessPaths(organization valueobj.HasOrganizationID, parentsToUpsert ...aggregate.DomainParent) {
	for i, parent := range parentsToUpsert {
		userAccessPaths := entity.DomainUserAccessPaths{}
		for _, userAccessPath := range parent.UserAccessPaths {
			userAccessPaths = append(userAccessPaths, entity.UserAccessPathWillBeDelegated{
				HasLocationID:     userAccessPath,
				HasUserID:         parent,
				HasOrganizationID: organization,
			})
		}
		parentsToUpsert[i].UserAccessPaths = userAccessPaths
	}
}

func (service *DomainParent) validateUpdateSystemAndExternalUserID(ctx context.Context, db libdatabase.Ext, parentsToUpdate aggregate.DomainParents) error {
	userIDsToUpdate := parentsToUpdate.ParentIDs()
	existedUsers, err := service.UserRepo.GetByIDs(ctx, db, userIDsToUpdate)
	if err != nil {
		return entity.InternalError{
			RawErr: errors.Wrap(err, "service.UserRepo.GetByIDs"),
		}
	}
	if len(existedUsers) != len(userIDsToUpdate) {
		existedUserIDs := existedUsers.UserIDs()
		for _, parent := range parentsToUpdate {
			if !golibs.InArrayString(parent.UserID().String(), existedUserIDs) {
				return entity.InvalidFieldError{
					EntityName: entity.ParentEntity,
					FieldName:  string(entity.UserFieldUserID),
					Index:      parent.IndexAttr,
					Reason:     entity.Invalid,
				}
			}
		}
	}

	/*
		if the external_user_id was updated, it would be failed:
		- create student with `external_user_id`, so `external_user_id` have value
			-> update student with another `external_user_id` will be failed.
		- create student without `external_user_id`, so `external_user_id` is null
			-> update again student with another `external_user_id` will be success.
	*/
	for _, parent := range parentsToUpdate {
		if !field.IsPresent(parent.ExternalUserID()) {
			continue
		}

		for _, existedUser := range existedUsers {
			if !parent.UserID().Equal(existedUser.UserID()) {
				continue
			}

			if !field.IsPresent(existedUser.ExternalUserID()) {
				continue
			}

			if parent.ExternalUserID() != existedUser.ExternalUserID() {
				return entity.InvalidFieldError{
					EntityName: entity.ParentEntity,
					Index:      parent.IndexAttr,
					FieldName:  string(entity.UserFieldExternalUserID),
					Reason:     entity.NotMatching,
				}
			}
		}
	}

	return nil
}

func (service *DomainParent) GetUsersByExternalIDs(ctx context.Context, externalIDs []string) (entity.Users, error) {
	users, err := service.UserRepo.GetByExternalUserIDs(ctx, service.DB, externalIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.UserRepo.GetByExternalUserIDs")
	}
	return users, nil
}

func (service *DomainParent) GetUsersByEmails(ctx context.Context, emails []string) (entity.Users, error) {
	users, err := service.UserRepo.GetByEmails(ctx, service.DB, emails)
	if err != nil {
		return nil, errors.Wrap(err, "service.UserRepo.GetByEmails")
	}
	return users, nil
}

func (service *DomainParent) GetStudentsAccessPaths(ctx context.Context, studentIDs []string) (entity.DomainUserAccessPaths, error) {
	accessPaths, err := service.UserAccessPathRepo.GetByUserIDs(ctx, service.DB, studentIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.UserAccessPathRepo.GetByUserIDs")
	}
	return accessPaths, nil
}

func (service *DomainParent) GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
	tags, err := service.TagRepo.GetByPartnerInternalIDs(ctx, service.DB, externalIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.TagRepo.GetByPartnerInternalIDs")
	}

	return tags, nil
}

func publishUpsertParentEvent(ctx context.Context, jsm nats.JetStreamManagement, student entity.User, parents aggregate.DomainParents) error {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "publishUpsertParentEvent.OrganizationFromContext")
	}

	createParentEvents := toCreatedParentEvent(organization, student, parents)
	err = publishDomainUserEvent(ctx, jsm, constants.SubjectUserCreated, createParentEvents...)
	if err != nil {
		return errors.Wrap(err, "service.publishDomainUserEvent")
	}

	assignedParentToStudentEvents := toParentAssignedToStudentEvent(student, parents)
	err = publishDomainUserEvent(ctx, jsm, constants.SubjectUserUpdated, assignedParentToStudentEvents...)
	if err != nil {
		return errors.Wrap(err, "service.publishDomainUserEvent")
	}

	return nil
}

func publishRemovedParentFromStudentEvent(ctx context.Context, jsm nats.JetStreamManagement, student valueobj.HasStudentID, parent valueobj.HasParentID) error {
	removedParentFromStudentEvent := toParentRemovedFromStudentEvent(student, parent)
	err := publishDomainUserEvent(ctx, jsm, constants.SubjectUserUpdated, []*pb.EvtUser{removedParentFromStudentEvent}...)
	if err != nil {
		return errors.Wrap(err, "service.publishDomainUserEvent")
	}

	return nil
}

func toCreatedParentEvent(orgID valueobj.HasOrganizationID, student entity.User, parents aggregate.DomainParents) []*pb.EvtUser {
	createParentEvents := make([]*pb.EvtUser, 0, len(parents))
	for _, parent := range parents {
		createParentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_CreateParent_{
				CreateParent: &pb.EvtUser_CreateParent{
					StudentId:   student.UserID().String(),
					ParentId:    parent.UserID().String(),
					StudentName: student.FullName().String(),
					SchoolId:    orgID.OrganizationID().String(),
				},
			},
		}
		createParentEvents = append(createParentEvents, createParentEvent)
	}

	return createParentEvents
}

func toParentAssignedToStudentEvent(student entity.User, parents aggregate.DomainParents) []*pb.EvtUser {
	parentAssignedToStudentEvents := make([]*pb.EvtUser, 0, len(parents))

	for _, parent := range parents {
		parentAssignedToStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_ParentAssignedToStudent_{
				ParentAssignedToStudent: &pb.EvtUser_ParentAssignedToStudent{
					StudentId: student.UserID().String(),
					ParentId:  parent.UserID().String(),
				},
			},
		}

		parentAssignedToStudentEvents = append(parentAssignedToStudentEvents, parentAssignedToStudentEvent)
	}

	return parentAssignedToStudentEvents
}

func toParentRemovedFromStudentEvent(student valueobj.HasStudentID, parent valueobj.HasParentID) *pb.EvtUser {
	parentRemovedFromStudentEvent := &pb.EvtUser{
		Message: &pb.EvtUser_ParentRemovedFromStudent_{
			ParentRemovedFromStudent: &pb.EvtUser_ParentRemovedFromStudent{
				StudentId: student.StudentID().String(),
				ParentId:  parent.ParentID().String(),
			},
		},
	}
	return parentRemovedFromStudentEvent
}

func (service *DomainParent) IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool {
	return unleash.IsFeatureUserNameStudentParentEnabled(service.UnleashClient, service.Env, organization)
}

func (service *DomainParent) IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error) {
	config, err := service.InternalConfigurationRepo.GetByKey(ctx, service.DB, constant.KeyAuthUsernameConfig)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return false, nil
		}
		return false, err
	}
	return config.ConfigValue().String() == constant.ConfigValueOn, nil
}

func (service *DomainParent) validateExternUserIDUsedByOtherRole(ctx context.Context, parents aggregate.DomainParents) error {
	externalUsers, err := service.UserRepo.GetByExternalUserIDs(ctx, service.DB, parents.ParentExternalUserIDs())
	if err != nil {
		return err
	}

	existingStudents, err := service.ParentRepo.GetUsersByExternalUserIDs(ctx, service.DB, parents.ParentExternalUserIDs())
	if err != nil {
		return err
	}

	for idx, parent := range parents {
		idxExternalUserIDByExistingUsers := utils.IndexOf(externalUsers.ExternalUserIDs(), parent.ExternalUserID().String())
		idxExternalUserIDByExistingStudents := utils.IndexOf(existingStudents.ExternalUserIDs(), parent.ExternalUserID().String())

		if idxExternalUserIDByExistingUsers == -1 {
			continue
		}

		if idxExternalUserIDByExistingStudents == -1 {
			return entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.ParentEntity,
				Index:      idx,
			}
		}
	}

	return nil
}
