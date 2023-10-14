package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/alert"
	"github.com/manabie-com/backend/internal/golibs/auth/multitenant"
	"github.com/manabie-com/backend/internal/golibs/clients"
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
	fpb "github.com/manabie-com/backend/pkg/manabuf/fatima/v1"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	featureToggleIgnoreEmailValidation             = "User_StudentManagement_IgnoreStudentEmailValidation"
	featureToggleIgnoreInvalidRecordsCSVAndOpenAPI = "User_StudentManagement_IgnoreInvalidRecordsCSVAndOpenAPI"
)

type UserPhoneNumberRepo interface {
	UpsertMultiple(ctx context.Context, db libdatabase.QueryExecer, userPhoneNumbers ...entity.DomainUserPhoneNumber) error
	SoftDeleteByUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) error
}

// DomainStudent represents for a domain service that
// contains all logics to deal with domain student
type DomainStudent struct {
	DB                 libdatabase.Ext
	JSM                nats.JetStreamManagement
	FirebaseAuthClient multitenant.TenantClient
	TenantManager      multitenant.TenantManager

	OrganizationRepo            OrganizationRepo
	EnrollmentStatusHistoryRepo DomainEnrollmentStatusHistoryRepo
	UserAccessPathRepo          DomainUserAccessPathRepo
	FatimaClient                fpb.SubscriptionModifierServiceClient
	ConfigurationClient         clients.ConfigurationClientInterface
	UnleashClient               unleashclient.ClientInstance
	SlackClient                 alert.SlackFactory
	Env                         string

	UserPhoneNumberRepo UserPhoneNumberRepo

	DomainParentService interface {
		UpsertMultiple(ctx context.Context, option unleash.DomainParentFeatureOption, parentsToUpsert ...aggregate.DomainParent) ([]aggregate.DomainParent, error)
		DomainParentsToUpsert(ctx context.Context, db libdatabase.Ext, isEnableUsername bool, parentsToUpsert ...aggregate.DomainParent) ([]aggregate.DomainParent, []aggregate.DomainParent, []aggregate.DomainParent, error)
		UpsertMultipleParentsInTx(ctx context.Context, tx libdatabase.Tx, parentsToCreate, parentsToUpdate, parentsToUpsert aggregate.DomainParents, option unleash.DomainParentFeatureOption) ([]aggregate.DomainParent, error)
	}
	StudentRepo interface {
		GetUsersByExternalUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (entity.Users, error)
		UpsertMultiple(ctx context.Context, db libdatabase.QueryExecer, isEnableUsername bool, studentsToCreate ...aggregate.DomainStudent) error
	}
	UserRepo interface {
		GetByEmails(ctx context.Context, db libdatabase.QueryExecer, emails []string) (entity.Users, error)
		GetByUserNames(ctx context.Context, db libdatabase.QueryExecer, usernames []string) (entity.Users, error)
		GetByEmailsInsensitiveCase(ctx context.Context, db libdatabase.QueryExecer, emails []string) (entity.Users, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (entity.Users, error)
		GetByExternalUserIDs(ctx context.Context, db libdatabase.QueryExecer, externalUserIDs []string) (entity.Users, error)

		UpdateActivation(ctx context.Context, db libdatabase.QueryExecer, users entity.Users) error
	}
	UserGroupRepo interface {
		FindUserGroupByRoleName(ctx context.Context, db libdatabase.QueryExecer, roleName string) (entity.DomainUserGroup, error)
	}
	UserAddressRepo interface {
		UpsertMultiple(ctx context.Context, db libdatabase.QueryExecer, userAddresses ...entity.DomainUserAddress) error
		SoftDeleteByUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) error
	}
	SchoolHistoryRepo interface {
		UpsertMultiple(ctx context.Context, db libdatabase.QueryExecer, schoolHistories ...entity.DomainSchoolHistory) error
		SoftDeleteByStudentIDs(ctx context.Context, db libdatabase.QueryExecer, studentIDs []string) error
		SetCurrentSchoolByStudentIDsAndSchoolIDs(ctx context.Context, db libdatabase.QueryExecer, studentIDs, schoolIDs []string) error
	}
	LocationRepo interface {
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainLocations, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) (entity.DomainLocations, error)
		RetrieveLowestLevelLocations(ctx context.Context, db libdatabase.Ext, name string, limit int32, offset int32, locationIDs []string) (entity.DomainLocations, error)
	}
	GradeRepo interface {
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) ([]entity.DomainGrade, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) ([]entity.DomainGrade, error)
		GetAll(ctx context.Context, db libdatabase.QueryExecer) ([]entity.DomainGrade, error)
	}
	SchoolRepo interface {
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) (entity.DomainSchools, error)
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainSchools, error)
		GetByIDsAndGradeID(ctx context.Context, db libdatabase.QueryExecer, schoolIDs []string, gradeID string) (entity.DomainSchools, error)
	}
	SchoolCourseRepo interface {
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainSchoolCourses, error)
		GetByPartnerInternalIDsAndSchoolIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs, schoolIDs []string) (entity.DomainSchoolCourses, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, schoolCourseIDs []string) (entity.DomainSchoolCourses, error)
	}
	PrefectureRepo interface {
		GetByPrefectureCodes(ctx context.Context, db libdatabase.QueryExecer, prefectureCodes []string) (entity.DomainPrefectures, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) ([]entity.DomainPrefecture, error)
		GetAll(ctx context.Context, db libdatabase.QueryExecer) ([]entity.DomainPrefecture, error)
	}
	UsrEmailRepo interface {
		CreateMultiple(ctx context.Context, db libdatabase.QueryExecer, users entity.Users) (valueobj.HasUserIDs, error)
		UpdateEmail(ctx context.Context, db libdatabase.QueryExecer, user entity.User) error
	}
	TagRepo interface {
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainTags, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) (entity.DomainTags, error)
	}
	TaggedUserRepo interface {
		UpsertBatch(ctx context.Context, db libdatabase.QueryExecer, taggedUsers ...entity.DomainTaggedUser) error
		SoftDeleteByUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) error
	}
	CourseRepo interface {
		GetByCoursePartnerIDs(ctx context.Context, db libdatabase.QueryExecer, coursePartnerIDs []string) (entity.DomainCourses, error)
	}
	StudentPackage interface {
		GetByStudentIDs(ctx context.Context, db libdatabase.QueryExecer, studentIDs []string) (entity.DomainStudentPackages, error)
		GetByStudentCourseAndLocationIDs(ctx context.Context, db libdatabase.QueryExecer, studentID string, courseID string, locationIDs []string) (entity.DomainStudentPackages, error)
	}
	StudentParentRepo interface {
		SoftDeleteByStudentIDs(ctx context.Context, db libdatabase.QueryExecer, studentIDs []string) error
		GetByStudentIDs(ctx context.Context, db libdatabase.QueryExecer, studentIDs []string) (entity.DomainStudentParentRelationships, error)
	}

	StudentParentRelationshipManager StudentParentRelationshipManager
	StudentValidationManager         interface {
		FullyValidate(ctx context.Context, db libdatabase.Ext, students aggregate.DomainStudents, iEnableUsername bool) (aggregate.DomainStudents, aggregate.DomainStudents, []error)
	}
	InternalConfigurationRepo interface {
		GetByKey(ctx context.Context, db libdatabase.QueryExecer, configKey string) (entity.DomainConfiguration, error)
	}
	AuthUserUpserter AuthUserUpserter

	// NOTE: example
	// StudentSFRepo interface {
	// 	Get(client salesforce.SFClient, limit int, offset int) ([]entity.DomainStudent, error)
	// 	GetByID(client salesforce.SFClient, studentID string) (entity.DomainStudent, error)
	// 	Create(client salesforce.SFClient, student entity.DomainStudent) error
	// }
	FeatureManager interface {
		FeatureUsernameToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
	}
}

func (service *DomainStudent) UpsertStudentCoursePackage(ctx context.Context, student aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	for _, course := range student.Courses {
		if field.IsPresent(course.StudentPackageID()) {
			updateStudentPackageCourseReq := &fpb.EditTimeStudentPackageRequest{
				StudentPackageId: course.StudentPackageID().String(),
				StartAt:          timestamppb.New(course.StartAt().Time()),
				EndAt:            timestamppb.New(course.EndAt().Time()),
				LocationIds:      []string{course.LocationID().String()},
			}
			_, err := service.FatimaClient.EditTimeStudentPackage(ctx, updateStudentPackageCourseReq)
			if err != nil {
				zapLogger.Error(
					"cannot edit time of student package",
					zap.Error(err),
					zap.String("Repo", "FatimaClient.EditTimeStudentPackage"),
				)
				return err
			}
		} else {
			addStudentPackageCourseReq := &fpb.AddStudentPackageCourseRequest{
				CourseIds: []string{course.CourseID().String()},
				StudentId: student.UserID().String(),
				StartAt:   timestamppb.New(course.StartAt().Time()),
				EndAt:     timestamppb.New(course.EndAt().Time()),
				StudentPackageExtra: []*fpb.AddStudentPackageCourseRequest_AddStudentPackageExtra{
					{
						CourseId:   course.CourseID().String(),
						LocationId: course.LocationID().String(),
					},
				},
			}
			_, err := service.FatimaClient.AddStudentPackageCourse(ctx, addStudentPackageCourseReq)
			if err != nil {
				zapLogger.Error(
					"cannot add student package",
					zap.Error(err),
					zap.String("Repo", "FatimaClient.AddStudentPackageCourse"),
				)
				return err
			}
		}
	}
	return nil
}

func (service *DomainStudent) UpsertMultipleWithErrorCollection(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
	var studentsToUpsert aggregate.DomainStudents
	studentsToCreate, studentsToUpdate, errorCollection := service.StudentValidationManager.FullyValidate(ctx, service.DB, domainStudents, option.EnableUsername)
	studentsToUpsert = append(studentsToUpsert, studentsToCreate...)
	studentsToUpsert = append(studentsToUpsert, studentsToUpdate...)

	createdUserEmails, err := service.UsrEmailRepo.CreateMultiple(ctx, service.DB, studentsToCreate.Users())
	if err != nil {
		return nil, []error{
			entity.InternalError{
				RawErr: errors.Wrap(err, "service.UsrEmailRepo.CreateMultiple"),
			},
		}
	}

	for i := range studentsToCreate {
		var userProfileLoginEmail valueobj.HasLoginEmail = studentsToCreate[i].DomainStudent
		if option.EnableUsername {
			userProfileLoginEmail = &entity.UserProfileLoginEmailDelegate{
				Email: createdUserEmails[i].UserID().String() + constant.LoginEmailPostfix,
			}
		}
		studentsToCreate[i].DomainStudent = entity.StudentWillBeDelegated{
			DomainStudentProfile: studentsToCreate[i].DomainStudent,
			HasGradeID:           studentsToCreate[i].DomainStudent,
			HasUserID:            createdUserEmails[i],
			HasLoginEmail:        userProfileLoginEmail,
		}
	}

	if err := libdatabase.ExecInTx(ctx, service.DB, func(ctx context.Context, tx pgx.Tx) error {
		upsertedStudents, err := service.upsertMultipleStudent(ctx, tx, option, studentsToCreate, studentsToUpdate, studentsToUpsert...)
		if err != nil {
			return err
		}
		studentsToUpsert = upsertedStudents
		return nil
	}); err != nil {
		return nil, []error{
			entity.InternalError{
				RawErr: errors.Wrap(err, "service.upsertMultipleStudent"),
			},
		}
	}
	return studentsToUpsert, errorCollection
}

func (service *DomainStudent) UpsertMultiple(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToUpsert ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error) {
	zapLogger := ctxzap.Extract(ctx)

	if err := service.validationUpsertStudent(ctx, option.EnableUsername, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"UpsertMultiple.validationUpsertStudent",
			zap.Error(err),
		)
		return nil, err
	}

	studentsToCreate, studentsToUpdate, err := service.generateUserIDs(ctx, option.EnableUsername, studentsToUpsert...)
	if err != nil {
		zapLogger.Error(
			"UpsertMultiple.generateUserIDs",
			zap.Error(err),
		)
		return nil, err
	}

	if err = service.ValidateUpdateSystemAndExternalUserID(ctx, studentsToUpdate); err != nil {
		zapLogger.Error(
			"UpsertMultiple.ValidateUpdateSystemAndExternalUserID",
			zap.Error(err),
		)
		return nil, err
	}

	studentsToUpsert, err = service.assignAggregate(ctx, studentsToUpsert...)
	if err != nil {
		zapLogger.Error(
			"UpsertMultiple.assignAggregate",
			zap.Error(err),
		)
		return nil, err
	}
	if err := libdatabase.ExecInTx(ctx, service.DB, func(ctx context.Context, tx pgx.Tx) error {
		upsertedStudents, err := service.upsertMultipleStudent(ctx, tx, option, studentsToCreate, studentsToUpdate, studentsToUpsert...)
		if err != nil {
			zapLogger.Error(
				"UpsertMultiple.upsertMultipleStudent",
				zap.Error(err),
			)
			return err
		}
		studentsToUpsert = upsertedStudents
		return nil
	}); err != nil {
		return nil, err
	}
	return studentsToUpsert, nil
}

func (service *DomainStudent) UpsertMultipleWithAssignedParent(ctx context.Context, studentsToUpsert aggregate.DomainStudentWithAssignedParents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudentWithAssignedParents, error) {
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "OrganizationFromContext")
	}

	if err := service.validationUpsertStudent(ctx, option.EnableUsername, studentsToUpsert.Students()...); err != nil {
		return nil, err
	}

	aggregateStudentsToUpsert := studentsToUpsert.Students()
	studentsToCreate, studentsToUpdate, err := service.generateUserIDs(ctx, option.EnableUsername, aggregateStudentsToUpsert...)
	if err != nil {
		return nil, err
	}

	if err = service.ValidateUpdateSystemAndExternalUserID(ctx, studentsToUpdate); err != nil {
		return nil, err
	}

	aggregateStudentsToUpsert, err = service.assignAggregate(ctx, aggregateStudentsToUpsert...)
	if err != nil {
		return nil, err
	}

	if err := libdatabase.ExecInTx(ctx, service.DB, func(ctx context.Context, tx pgx.Tx) error {
		aggregateStudentsToUpsert, err := service.upsertMultipleStudent(ctx, tx, option, studentsToCreate, studentsToUpdate, aggregateStudentsToUpsert...)
		if err != nil {
			return fmt.Errorf("upsertMultipleStudent: %s", err.Error())
		}

		for i, student := range studentsToUpsert {
			if len(student.Parents) > 0 {
				student.DomainStudent = aggregateStudentsToUpsert[i]
				for _, parent := range student.Parents {
					parent.UserAccessPaths = student.UserAccessPaths
				}
				parentsToCreate, parentsToUpdate, parentsToUpsert, err := service.DomainParentService.DomainParentsToUpsert(ctx, tx, option.EnableUsername, student.Parents...)
				if err != nil {
					return fmt.Errorf("service.DomainParentService.DomainParentsToUpsert: %s", err.Error())
				}
				domainParentOption := unleash.DomainParentFeatureOption{
					DomainUserFeatureOption: option.DomainUserFeatureOption,
				}
				_, err = service.DomainParentService.UpsertMultipleParentsInTx(ctx, tx, parentsToCreate, parentsToUpdate, parentsToUpsert, domainParentOption)
				if err != nil {
					return fmt.Errorf("service.DomainParentService.UpsertMultipleParentsInTx: %s", err.Error())
				}
				studentParentsToDelete, err := service.StudentParentRepo.GetByStudentIDs(ctx, tx, []string{student.UserID().String()})
				if err != nil {
					return fmt.Errorf("service.StudentParentRepo.GetByStudentIDs: %s", err.Error())
				}
				if err := service.StudentParentRepo.SoftDeleteByStudentIDs(ctx, tx, []string{student.UserID().String()}); err != nil {
					return fmt.Errorf("service.StudentParentRepo.SoftDeleteByStudentIDs: %s", err.Error())
				}
				for _, parent := range student.Parents {
					err := service.StudentParentRelationshipManager(ctx, tx, organization, field.NewString(string(constant.FamilyRelationshipOther)), student, parent)
					if err != nil {
						return fmt.Errorf("StudentParentRelationshipManager: %s", err.Error())
					}
				}
				for _, studentParent := range studentParentsToDelete {
					if err := publishRemovedParentFromStudentEvent(ctx, service.JSM, studentParent, studentParent); err != nil {
						return fmt.Errorf("publishRemovedParentFromStudentEvent: %s", err.Error())
					}
				}
				if err := publishUpsertParentEvent(ctx, service.JSM, student.DomainStudent, student.Parents); err != nil {
					return fmt.Errorf("publishUpsertParentEvent: %s", err.Error())
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return studentsToUpsert, nil
}

func (service *DomainStudent) UpdateUserActivation(ctx context.Context, users entity.Users) error {
	userDBs, err := service.UserRepo.GetByIDs(ctx, service.DB, users.UserIDs())
	if err != nil {
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "userRepo.GetByIDs"),
		}
	}

	if len(userDBs) != len(users.UserIDs()) {
		return errcode.Error{
			Code: errcode.InvalidData,
			Err:  fmt.Errorf("user ids are not valid"),
		}
	}

	err = libdatabase.ExecInTx(ctx, service.DB, func(ctx context.Context, tx pgx.Tx) error {
		if len(users) < 1 {
			return nil
		}

		err := service.UserRepo.UpdateActivation(ctx, tx, users)
		if err != nil {
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  fmt.Errorf("userRepo.UpdateActivation: %w", err),
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (service *DomainStudent) assignAggregate(ctx context.Context, studentsToUpsert ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error) {
	zapLogger := ctxzap.Extract(ctx)
	currentUser, err := service.getCurrentUser(ctx)
	if err != nil {
		zapLogger.Error(
			"assignAggregate.getCurrentUser",
			zap.Error(err),
		)
		// TODO: should throw specific error
		return nil, err
	}
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		zapLogger.Error(
			"assignAggregate.OrganizationFromContext",
			zap.Error(err),
		)
		return nil, errors.Wrap(err, "OrganizationFromContext")
	}

	for i := range studentsToUpsert {
		studentsToUpsert[i].DomainStudent = &entity.StudentWillBeDelegated{
			DomainStudentProfile: studentsToUpsert[i].DomainStudent,
			HasOrganizationID:    organization,
			HasSchoolID:          organization,
			HasCountry:           currentUser,
			HasGradeID:           studentsToUpsert[i].DomainStudent,
			HasUserID:            studentsToUpsert[i].DomainStudent,
			HasLoginEmail:        studentsToUpsert[i].DomainStudent,
		}
		studentsToUpsert[i].LegacyUserGroups = entity.LegacyUserGroups{entity.DelegateToLegacyUserGroup(&entity.StudentLegacyUserGroup{}, organization, studentsToUpsert[i])}
	}

	if err := service.setUserAccessPaths(ctx, organization, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"assignAggregate.setUserAccessPaths",
			zap.Error(err),
		)
		return nil, err
	}

	if err := service.setUserGroupMembers(ctx, organization, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"assignAggregate.setUserGroupMembers",
			zap.Error(err),
		)
		return nil, err
	}

	service.setEnrollmentStatusHistories(organization, studentsToUpsert...)

	if err := service.validateEnrollmentStatusHistories(ctx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"assignAggregate.validateEnrollmentStatusHistories",
			zap.Error(err),
		)
		return nil, err
	}

	if err := setUserPhoneNumbers(studentsToUpsert...); err != nil {
		zapLogger.Error(
			"assignAggregate.setUserPhoneNumbers",
			zap.Error(err),
		)
		return nil, err
	}

	return studentsToUpsert, nil
}

func (service *DomainStudent) validationUpsertStudent(ctx context.Context, isEnableUsername bool, studentsToUpsert ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)

	for idx := range studentsToUpsert {
		if err := aggregate.ValidStudent(studentsToUpsert[idx], isEnableUsername); err != nil {
			zapLogger.Error(
				"validationUpsertStudent.ValidStudent",
				zap.Error(err),
			)
			return err
		}
	}

	if err := validateStudentDuplicatedFields(studentsToUpsert...); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateStudentDuplicatedFields",
			zap.Error(err),
		)
		return err
	}

	users := entity.Users{}
	for _, user := range studentsToUpsert {
		users = append(users, user)
	}

	if err := service.validateExternalUserIDExistedInSystem(ctx, users); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateExternalUserIDExistedInSystem",
			zap.Error(err),
		)
		return err
	}

	if err := ValidateUserEmailsExistedInSystem(ctx, service.UserRepo, service.DB, users); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.ValidateUserEmailsExistedInSystem",
			zap.Error(err),
		)
		return err
	}

	if isEnableUsername {
		if err := ValidateUserNamesExistedInSystem(ctx, service.UserRepo, service.DB, users); err != nil {
			zapLogger.Error(
				"validationUpsertStudent.ValidateUserNamesExistedInSystem",
				zap.Error(err),
			)
			return err
		}
	}

	if err := service.validateGrade(ctx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateGrade",
			zap.Error(err),
		)
		return err
	}

	if err := validateStudentPhoneNumbers(studentsToUpsert...); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateStudentPhoneNumbers",
			zap.Error(err),
		)
		return err
	}

	if err := service.validateSchoolHistories(ctx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateSchoolHistories",
			zap.Error(err),
		)
		return err
	}

	if err := service.validateUserAddress(ctx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateUserAddress",
			zap.Error(err),
		)
		return err
	}

	if err := service.validateTags(ctx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateTags",
			zap.Error(err),
		)
		return err
	}

	if err := service.validateStudentsLocations(ctx, service.DB, studentsToUpsert); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateStudentsLocations",
			zap.Error(err),
		)
		return err
	}
	if err := validateEnrollmentStatusHistoriesBeforeCreating(ctx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"validationUpsertStudent.validateEnrollmentStatusHistoriesBeforeCreating",
			zap.Error(err),
		)
		return err
	}
	return nil
}

func (service *DomainStudent) upsertMultipleStudent(ctx context.Context, tx libdatabase.Ext, option unleash.DomainStudentFeatureOption, studentsToCreate, studentsToUpdate aggregate.DomainStudents, _studentsToUpsert ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error) {
	zapLogger := ctxzap.Extract(ctx)
	var studentsToUpsert aggregate.DomainStudents = _studentsToUpsert
	now := time.Now()

	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "OrganizationFromContext")
	}

	if !option.EnableIgnoreUpdateEmail && !option.EnableUsername {
		for _, student := range studentsToUpdate {
			err = service.UsrEmailRepo.UpdateEmail(ctx, tx, student)

			if err != nil {
				zapLogger.Error(
					"cannot update email",
					zap.Error(err),
					zap.String("Repo", "UsrEmailRepo.UpdateEmail"),
				)
				return nil, errcode.Error{
					Code: errcode.InternalError,
					Err:  errors.Wrap(err, "service.UsrEmailRepo.UpdateEmail"),
				}
			}
		}
	}

	zapLogger.Debug(
		"--end service.UsrEmailRepo.UpdateEmail--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()

	if err := service.StudentRepo.UpsertMultiple(ctx, tx, option.EnableUsername, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"cannot upsert students",
			zap.Error(err),
			zap.String("Repo", "StudentRepo.UpsertMultiple"),
		)
		return nil, entity.InternalError{
			RawErr: errors.Wrap(err, "service.StudentRepo.UpsertMultiple"),
		}
	}

	zapLogger.Debug(
		"--end service.StudentRepo.UpsertMultiple--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()

	if err := service.upsertUserAddresses(ctx, tx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"upsertMultipleStudent.upsertUserAddresses",
			zap.Error(err),
		)
		return nil, entity.InternalError{
			RawErr: errors.Wrap(err, "service.upsertUserAddresses"),
		}
	}
	zapLogger.Debug(
		"--end  service.upsertUserAddresses--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()

	var userPhoneNumbers entity.DomainUserPhoneNumbers
	for _, student := range studentsToUpsert {
		userPhoneNumbers = append(userPhoneNumbers, student.UserPhoneNumbers...)
	}
	if err := upsertUserPhoneNumbers(ctx, tx, service.UserPhoneNumberRepo, studentsToUpsert.Users(), userPhoneNumbers); err != nil {
		zapLogger.Error(
			"upsertMultipleStudent.upsertUserPhoneNumbers",
			zap.Error(err),
		)
		return nil, entity.InternalError{
			RawErr: errors.Wrap(err, "service.upsertUserPhoneNumbers"),
		}
	}
	zapLogger.Debug(
		"--end  upsertUserPhoneNumbers--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()

	if err := service.upsertSchoolHistories(ctx, tx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"upsertMultipleStudent.upsertSchoolHistories",
			zap.Error(err),
		)
		return nil, entity.InternalError{
			RawErr: errors.Wrap(err, "service.upsertSchoolHistories"),
		}
	}

	zapLogger.Debug(
		"--end  upsertSchoolHistories--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()

	if option.EnableExperimentalBulkInsertEnrollmentStatusHistories {
		if err := service.EnrollmentStatusHistoryRepo.BulkInsert(ctx, tx, studentsToCreate.EnrollmentStatusHistories()); err != nil {
			return nil, err
		}
		if err := service.upsertEnrollmentStatusHistories(ctx, tx, studentsToUpdate...); err != nil {
			return nil, err
		}
	} else {
		if err := service.upsertEnrollmentStatusHistories(ctx, tx, studentsToUpsert...); err != nil {
			zapLogger.Error(
				"upsertMultipleStudent.upsertEnrollmentStatusHistories",
				zap.Error(err),
			)
			return nil, err
		}
	}
	zapLogger.Debug(
		"--end  upsertEnrollmentStatusHistories--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()

	if err := service.UpsertTaggedUsers(ctx, tx, studentsToUpsert...); err != nil {
		zapLogger.Error(
			"upsertMultipleStudent.UpsertTaggedUsers",
			zap.Error(err),
		)
		return nil, err
	}

	zapLogger.Debug(
		"--end  UpsertTaggedUsers--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()

	if len(studentsToUpsert) > 0 && !option.DisableAutoDeactivateAndReactivateStudent {
		enrollmentStatuses := []string{pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String()}
		if option.EnableAutoDeactivateAndReactivateStudentV2 {
			enrollmentStatuses = []string{
				pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_WITHDRAWN.String(),
				pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_GRADUATED.String(),
				pb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
			}
		}
		if err := service.EnrollmentStatusHistoryRepo.UpdateStudentStatusBasedEnrollmentStatus(ctx, tx, studentsToUpsert.StudentIDs(), enrollmentStatuses); err != nil {
			return nil, err
		}
	}
	zapLogger.Debug(
		"--end insert DeactivateAndReactivateStudents--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)
	now = time.Now()
	// Upsert users in auth platform
	// AuthUserUpserter must be used with service.DB, not current tx
	// Because users are updated in current tx before AuthUserUpserter invoke,
	// but it needs to query the user data before updating to validate
	if _, err := service.AuthUserUpserter(ctx, service.DB, organization, studentsToCreate.Users(), studentsToUpdate.Users(), option.DomainUserFeatureOption); err != nil {
		return nil, err
	}
	zapLogger.Debug(
		"--end send to firebase--",
		zap.Int64("time-end", time.Since(now).Milliseconds()),
	)

	/*// Import to identity platform
	if err := service.upsertUserInIdentityPlatform(ctx, tx, skipUpdateEmail, organization, studentsToCreate.Users(), studentsToUpdate.Users()); err != nil {
		return nil, err
	}*/

	evtUsers := toCreatedStudentEvent(organization.OrganizationID().String(), studentsToCreate...)
	err = publishDomainUserEvent(ctx, service.JSM, constants.SubjectUserCreated, evtUsers...)
	if err != nil {
		return nil, err
	}

	// For CSV Import and OpenAPI inputs, the locations in input can be empty when updating.
	// We can not send event to downstream service with locations in aggregate because aggregate
	// is initialized by input, so it's empty although users are created with locations before
	// We should think how to keep data of aggregate update-to-date with db in update case
	// TODO: Can remove this one after apply func UpsertMultipleWithErrorCollection for all port

	enrollmentStatusHistories, err := service.EnrollmentStatusHistoryRepo.GetByStudentIDs(ctx, tx, studentsToUpdate.StudentIDs())
	if err != nil {
		return nil, err
	}

	enrollmentStatusHistoriesByUserID := groupEnrollmentStatusHistoriesByUserID(enrollmentStatusHistories)
	for idx := range studentsToUpdate {
		histories := enrollmentStatusHistoriesByUserID[studentsToUpdate[idx].UserID()]
		studentsToUpdate[idx].EnrollmentStatusHistories = histories
	}

	evtUsers = toUpdatedUserEvent(studentsToUpdate...)
	err = publishDomainUserEvent(ctx, service.JSM, constants.SubjectUserUpdated, evtUsers...)
	if err != nil {
		return nil, err
	}
	for _, student := range studentsToUpsert {
		if len(student.Courses) > 0 {
			if err = service.UpsertStudentCoursePackage(ctx, student); err != nil {
				return nil, err
			}
		}
	}
	return studentsToUpsert, nil
}

// groupEnrollmentStatusHistoriesByUserID groups enrollment status histories by user ID
func groupEnrollmentStatusHistoriesByUserID(enrollmentStatusHistories entity.DomainEnrollmentStatusHistories) map[field.String]entity.DomainEnrollmentStatusHistories {
	groupedHistoriesByUserID := make(map[field.String]entity.DomainEnrollmentStatusHistories)

	for _, enrollmentStatusHistory := range enrollmentStatusHistories {
		userID := enrollmentStatusHistory.UserID()

		// If this user ID hasn't been added to the grouped histories yet, add an empty slice for it
		if _, ok := groupedHistoriesByUserID[userID]; !ok {
			groupedHistoriesByUserID[userID] = make(entity.DomainEnrollmentStatusHistories, 0)
		}

		// Append the current history to the appropriate user ID key in the map
		groupedHistoriesByUserID[userID] = append(groupedHistoriesByUserID[userID], enrollmentStatusHistory)
	}

	return groupedHistoriesByUserID
}

func updateUserEmailsInIdentityPlatform(ctx context.Context, identityPlatformTenantManager multitenant.TenantManager, tenantID string, usersToUpdate ...entity.User) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()

	tenantClient, err := identityPlatformTenantManager.TenantClient(ctx, tenantID)
	if err != nil {
		zapLogger.Warnw(
			"cannot get tenant client",
			"tenantID", tenantID,
			"err", err.Error(),
		)
		return errors.Wrap(err, "service.TenantManager.TenantClient")
	}

	for _, student := range usersToUpdate {
		err = UpdateUserEmail(ctx, tenantClient, student.UserID().String(), student.LoginEmail().String())
		if err != nil {
			zapLogger.Error(
				"updateUserEmailsInIdentityPlatform.UpdateUserEmail",
				zap.Error(err),
			)
			return errors.Wrap(err, "UpdateUserEmail in Identity Platform")
		}
	}
	return nil
}

func createUsersInIdentityPlatform(ctx context.Context, identityPlatformTenantManager multitenant.TenantManager, tenantID string, organizationID int64, usersToCreate ...entity.User) (entity.LegacyUsers, error) {
	backwardCompatibleAuthUsers := entity.LegacyUsers{}

	if len(usersToCreate) == 0 {
		return backwardCompatibleAuthUsers, nil
	}

	for _, student := range usersToCreate {
		backwardCompatibleAuthUser := &entity.LegacyUser{
			ID:         libdatabase.Text(student.UserID().String()),
			LoginEmail: libdatabase.Text(student.LoginEmail().String()),
			UserAdditionalInfo: entity.UserAdditionalInfo{
				Password: student.Password().String(),
			},
		}
		backwardCompatibleAuthUsers = append(backwardCompatibleAuthUsers, backwardCompatibleAuthUser)
	}

	err := CreateUsersInIdentityPlatform(ctx, identityPlatformTenantManager, tenantID, backwardCompatibleAuthUsers, organizationID)
	if err != nil {
		return backwardCompatibleAuthUsers, errors.Wrap(err, "CreateUsersInIdentityPlatform")
	}
	return backwardCompatibleAuthUsers, nil
}

func (service *DomainStudent) getCurrentUser(ctx context.Context) (entity.User, error) {
	zapLogger := ctxzap.Extract(ctx)
	currentUserID := interceptors.UserIDFromContext(ctx)
	users, err := service.UserRepo.GetByIDs(ctx, service.DB, []string{currentUserID})
	if err != nil {
		zapLogger.Error(
			"cannot get current user",
			zap.Error(err),
			zap.String("Repo", "UserRepo.GetByIDs"),
			zap.Strings("userIDs", []string{currentUserID}),
		)
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.UserRepo.GetByIDs"),
		}
	}
	if len(users) == 0 {
		return nil, errcode.Error{
			Code:     errcode.NotFound,
			Resource: "user",
		}
	}
	return users[0], nil
}

func (service *DomainStudent) generateUserIDs(ctx context.Context, isEnableUsername bool, students ...aggregate.DomainStudent) ([]aggregate.DomainStudent, []aggregate.DomainStudent, error) {
	zapLogger := ctxzap.Extract(ctx)
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		zapLogger.Error(
			"generateUserIDs.OrganizationFromContext",
			zap.Error(err),
		)
		return nil, nil, errors.Wrap(err, "OrganizationFromContext")
	}

	studentsToCreate := []aggregate.DomainStudent{}
	studentsToUpdate := []aggregate.DomainStudent{}
	usersToGenID := entity.Users{}

	for _, student := range students {
		if student.UserID().String() != "" {
			studentsToUpdate = append(studentsToUpdate, student)
		} else {
			studentsToCreate = append(studentsToCreate, student)
			uid := idutil.ULIDNow()

			var userProfileLoginEmail valueobj.HasLoginEmail = student.DomainStudent
			if isEnableUsername {
				userProfileLoginEmail = &entity.UserProfileLoginEmailDelegate{
					Email: uid + constant.LoginEmailPostfix,
				}
			}
			usersToGenID = append(usersToGenID, entity.StudentWillBeDelegated{
				DomainStudentProfile: student.DomainStudent,
				HasOrganizationID:    organization,
				HasUserID: &valueobj.RandomHasUserID{
					RandomUserID: field.NewString(uid),
				},
				HasLoginEmail: userProfileLoginEmail,
			})
		}
	}

	createdUserEmails, err := service.UsrEmailRepo.CreateMultiple(ctx, service.DB, usersToGenID)
	if err != nil {
		zapLogger.Error(
			"cannot create user emails",
			zap.String("Repo", "UsrEmailRepo.CreateMultiple"),
			zap.Error(err),
		)
		return nil, nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.UsrEmailRepo.CreateMultiple"),
		}
	}

	for i := range studentsToCreate {
		var userProfileLoginEmail valueobj.HasLoginEmail = studentsToCreate[i].DomainStudent
		if isEnableUsername {
			userProfileLoginEmail = &entity.UserProfileLoginEmailDelegate{
				Email: createdUserEmails[i].UserID().String() + constant.LoginEmailPostfix,
			}
		}
		studentsToCreate[i].DomainStudent = entity.StudentWillBeDelegated{
			DomainStudentProfile: studentsToCreate[i].DomainStudent,
			HasGradeID:           studentsToCreate[i].DomainStudent,
			HasUserID:            createdUserEmails[i],
			HasLoginEmail:        userProfileLoginEmail,
		}
		for j := range students {
			studentToCreateEmail := studentsToCreate[i].Email()
			studentToUpsertEmail := students[j].Email()
			if isEnableUsername {
				studentToCreateEmail = studentsToCreate[i].UserName()
				studentToUpsertEmail = students[j].UserName()
			}
			if studentToCreateEmail == studentToUpsertEmail {
				students[j].DomainStudent = studentsToCreate[i].DomainStudent
			}
		}
	}

	if len(studentsToUpdate) == 0 {
		return studentsToCreate, studentsToUpdate, nil
	}

	return studentsToCreate, studentsToUpdate, nil
}

func (service *DomainStudent) validateTags(ctx context.Context, students ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	for _, student := range students {
		if len(student.TaggedUsers) == 0 {
			continue
		}
		tags, err := service.TagRepo.GetByIDs(ctx, service.DB, student.TaggedUsers.TagIDs())
		if err != nil {
			zapLogger.Error(
				"cannot get user tags",
				zap.Error(err),
				zap.String("Repo", "TagRepo.GetByIDs"),
				zap.Strings("tagIDs", student.TaggedUsers.TagIDs()),
			)
			return err
		}
		if len(student.TaggedUsers) != len(tags) {
			zapLogger.Error(
				"TaggedUsers from client is not equal to tags in BD",
				zap.String("Function", "validateTags"),
				zap.Strings("TaggedUsers from client", student.TaggedUsers.TagIDs()),
				zap.Strings("tags in client", tags.TagIDs()),
			)
			return entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentTagsField,
				Index:      student.IndexAttr,
			}
		}

		for _, tag := range tags {
			if !golibs.InArrayString(tag.TagType().String(), entity.StudentTags) {
				zapLogger.Error(
					"TagType is wrong",
					zap.String("Function", "validateTags"),
					zap.String("TagType", tag.TagType().String()),
					zap.Strings(" StudentTags", entity.StudentTags),
				)
				return entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentTagsField,
					Index:      student.IndexAttr,
					Reason:     entity.InvalidTagType,
				}
			}
		}
	}
	return nil
}

func setUserPhoneNumbers(students ...aggregate.DomainStudent) error {
	for idx, student := range students {
		if len(student.UserPhoneNumbers) == 0 {
			continue
		}

		registeredPhoneNumber := make(map[string]struct{})
		userPhoneNumbers := make(entity.DomainUserPhoneNumbers, 0, len(student.UserPhoneNumbers))
		for _, phoneNumber := range student.UserPhoneNumbers {
			if !(phoneNumber.PhoneNumber().IsEmpty()) {
				if _, ok := registeredPhoneNumber[phoneNumber.PhoneNumber().String()]; ok {
					return errcode.Error{
						Code:      errcode.DuplicatedData,
						Err:       errors.New("duplicated home_phone_number"),
						FieldName: fmt.Sprintf("students[%d].phone_number.home_phone_number", idx),
						Index:     idx,
					}
				}
				registeredPhoneNumber[phoneNumber.PhoneNumber().String()] = struct{}{}
			}

			userPhoneNumbers = append(userPhoneNumbers, entity.UserPhoneNumberWillBeDelegated{
				UserPhoneNumberAttribute: phoneNumber,
				HasUserID:                student,
				HasOrganizationID:        student,
			})
		}

		students[idx].UserPhoneNumbers = userPhoneNumbers
	}
	return nil
}

func (service *DomainStudent) setUserGroupMembers(ctx context.Context, organization valueobj.HasOrganizationID, studentsToCreate ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	studentUserGroup, err := service.UserGroupRepo.FindUserGroupByRoleName(ctx, service.DB, constant.RoleStudent)
	if err != nil {
		zapLogger.Error(
			"cannot get user groups",
			zap.Error(err),
			zap.String("Repo", "UserGroupRepo.FindUserGroupByRoleName"),
			zap.String("roleName", constant.RoleStudent),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.UserGroupRepo.FindUserGroupByRoleName"),
		}
	}

	if err == nil {
		for i, student := range studentsToCreate {
			studentsToCreate[i].UserGroupMembers = append(studentsToCreate[i].UserGroupMembers, entity.UserGroupMemberWillBeDelegated{
				HasUserGroupID:    studentUserGroup,
				HasUserID:         student,
				HasOrganizationID: organization,
			})
		}
	}
	return nil
}

func validateStudentPhoneNumbers(studentsToCreate ...aggregate.DomainStudent) error {
	for _, student := range studentsToCreate {
		if err := ValidateUserPhoneNumbers(student.UserPhoneNumbers, student.IndexAttr); err != nil {
			return err
		}
	}
	return nil
}

func (service *DomainStudent) validateGrade(ctx context.Context, studentsToCreate ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	for _, student := range studentsToCreate {
		if !field.IsPresent(student.GradeID()) {
			zapLogger.Error(
				"field grades is not present",
				zap.String("Function", "validateGrade"),
				zap.String("GradeID", student.GradeID().String()),
			)
			return entity.MissingMandatoryFieldError{
				FieldName:  entity.StudentGradeField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
			}
		}
		gradeIDs := []string{student.GradeID().String()}
		grades, err := service.GradeRepo.GetByIDs(ctx, service.DB, gradeIDs)
		if err != nil {
			zapLogger.Error(
				"cannot get grades",
				zap.Error(err),
				zap.String("Repo", "GradeRepo.GetByIDs"),
				zap.Strings("gradeIDs", gradeIDs),
			)
			return err
		}
		if len(grades) != 1 {
			zapLogger.Error(
				"grades is not equal to 1",
				zap.String("Function", "validateGrade"),
				zap.String("len of grades", strconv.Itoa(len(grades))),
			)
			return entity.NotFoundError{
				EntityName: entity.GradeEntity,
				FieldName:  string(entity.StudentGradeField),
				FieldValue: strings.Join(gradeIDs, ";"),
				Index:      student.IndexAttr,
			}
		}
	}

	return nil
}

func toUpdatedUserEvent(studentsToUpdate ...aggregate.DomainStudent) []*pb.EvtUser {
	updateStudentEvents := make([]*pb.EvtUser, 0, len(studentsToUpdate))

	for _, student := range studentsToUpdate {
		updateStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_UpdateStudent_{
				UpdateStudent: &pb.EvtUser_UpdateStudent{
					StudentId:                 student.UserID().String(),
					DeviceToken:               student.DeviceToken().String(),
					AllowNotification:         student.AllowNotification().Boolean(),
					Name:                      student.FullName().String(),
					StudentFirstName:          student.FirstName().String(),
					StudentLastName:           student.LastName().String(),
					StudentFirstNamePhonetic:  student.FirstNamePhonetic().String(),
					StudentLastNamePhonetic:   student.LastNamePhonetic().String(),
					LocationIds:               field.Strings(student.EnrollmentStatusHistories.LocationIDs()).Strings(),
					EnrollmentStatusHistories: toPbEnrollmentStatusHistories(student.EnrollmentStatusHistories),
				},
			},
		}
		if student.UserAddress != nil {
			updateStudentEvent.GetUpdateStudent().UserAddress = toPbUserAddress(student.UserAddress)
		}
		updateStudentEvents = append(updateStudentEvents, updateStudentEvent)
	}

	return updateStudentEvents
}

func toPbUserAddress(userAddress entity.DomainUserAddress) *pb.UserAddress {
	return &pb.UserAddress{
		AddressId:    userAddress.UserAddressID().String(),
		AddressType:  pb.AddressType(pb.AddressType_value[userAddress.AddressType().String()]),
		PostalCode:   userAddress.PostalCode().String(),
		Prefecture:   userAddress.PrefectureID().String(),
		City:         userAddress.City().String(),
		FirstStreet:  userAddress.FirstStreet().String(),
		SecondStreet: userAddress.SecondStreet().String(),
	}
}

func toCreatedStudentEvent(schoolID string, studentsToCreate ...aggregate.DomainStudent) []*pb.EvtUser {
	createStudentEvents := make([]*pb.EvtUser, 0, len(studentsToCreate))

	for _, student := range studentsToCreate {
		createStudentEvent := &pb.EvtUser{
			Message: &pb.EvtUser_CreateStudent_{
				CreateStudent: &pb.EvtUser_CreateStudent{
					StudentId:                student.UserID().String(),
					StudentName:              student.FullName().String(),
					SchoolId:                 schoolID,
					LocationIds:              student.UserAccessPaths.LocationIDs(),
					StudentFirstName:         student.FirstName().String(),
					StudentLastName:          student.LastName().String(),
					StudentFirstNamePhonetic: student.FirstNamePhonetic().String(),
					StudentLastNamePhonetic:  student.LastNamePhonetic().String(),
				},
			},
		}
		if student.UserAddress != nil {
			createStudentEvent.GetCreateStudent().UserAddress = toPbUserAddress(student.UserAddress)
		}
		if student.EnrollmentStatusHistories != nil {
			createStudentEvent.GetCreateStudent().EnrollmentStatusHistories = toPbEnrollmentStatusHistories(student.EnrollmentStatusHistories)
		}
		createStudentEvents = append(createStudentEvents, createStudentEvent)
	}

	return createStudentEvents
}

func toPbEnrollmentStatusHistories(enrollmentStatusHistories entity.DomainEnrollmentStatusHistories) []*pb.EnrollmentStatusHistory {
	statusHistories := make([]*pb.EnrollmentStatusHistory, 0, len(enrollmentStatusHistories))
	for _, enrollmentStatusHistory := range enrollmentStatusHistories {
		status := enrollmentStatusHistory.EnrollmentStatus().String()
		enrollmentStatus := &pb.EnrollmentStatusHistory{
			StudentId:        enrollmentStatusHistory.UserID().String(),
			LocationId:       enrollmentStatusHistory.LocationID().String(),
			EnrollmentStatus: pb.StudentEnrollmentStatus(pb.StudentEnrollmentStatus_value[status]),
		}

		if field.IsPresent(enrollmentStatusHistory.StartDate()) {
			enrollmentStatus.StartDate = timestamppb.New(enrollmentStatusHistory.StartDate().Time())
		}
		if field.IsPresent(enrollmentStatusHistory.EndDate()) {
			enrollmentStatus.EndDate = timestamppb.New(enrollmentStatusHistory.EndDate().Time())
		}

		statusHistories = append(statusHistories, enrollmentStatus)
	}
	return statusHistories
}

func (service *DomainStudent) validateUserAddress(ctx context.Context, studentsToCreate ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	for _, student := range studentsToCreate {
		if student.UserAddress == nil {
			continue
		}

		if field.IsPresent(student.UserAddress.PrefectureID()) {
			prefectureIDs := []string{student.UserAddress.PrefectureID().String()}
			prefectures, err := service.PrefectureRepo.GetByIDs(ctx, service.DB, prefectureIDs)
			if err != nil {
				zapLogger.Error(
					"cannot get prefectures",
					zap.Error(err),
					zap.String("Repo", "PrefectureRepo.GetByIDs"),
					zap.Strings("prefectureIDs", prefectureIDs),
				)
				return err
			}
			if len(prefectureIDs) != len(prefectures) {
				return entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentUserAddressPrefectureField,
					Index:      student.IndexAttr,
				}
			}

			// len(prefectures) cannot be zero because of prefectureCodes var above
		}
	}
	return nil
}

func (service *DomainStudent) setUserAccessPaths(ctx context.Context, organization valueobj.HasOrganizationID, studentsToCreate ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx).Sugar()
	for idx, student := range studentsToCreate {
		if len(student.EnrollmentStatusHistories) > 0 {
			userAccessPaths := entity.DomainUserAccessPaths{}
			for j, enrollmentStatusHistory := range student.EnrollmentStatusHistories {
				// validate empty locationID
				if enrollmentStatusHistory.LocationID().String() == "" {
					zapLogger.Error(
						"location is empty",
						zap.String("Function", "setUserAccessPaths"),
					)
					return errcode.Error{
						Code:      errcode.InvalidData,
						FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].location", idx, j),
						Index:     idx,
					}
				}
				// validate duplicate locationID
				if golibs.InArrayString(enrollmentStatusHistory.LocationID().String(), userAccessPaths.LocationIDs()) {
					zapLogger.Error(
						"location is duplicate",
						zap.String("Function", "setUserAccessPaths"),
						zap.String("enrollmentStatusHistory.LocationID", enrollmentStatusHistory.LocationID().String()),
						zap.Strings("userAccessPaths.LocationIDs", userAccessPaths.LocationIDs()),
					)
					return errcode.Error{
						Code:      errcode.DuplicatedData,
						FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].location", idx, j),
						Index:     idx,
					}
				}
				userAccessPaths = append(userAccessPaths, entity.UserAccessPathWillBeDelegated{
					HasLocationID:     enrollmentStatusHistory,
					HasUserID:         student,
					HasOrganizationID: organization,
				})
			}
			locations, err := service.LocationRepo.GetByIDs(ctx, service.DB, userAccessPaths.LocationIDs())
			if err != nil {
				zapLogger.Error(
					"cannot get locations",
					zap.Error(err),
					zap.String("Repo", "LocationRepo.GetByIDs"),
					zap.Strings("locationIDs", userAccessPaths.LocationIDs()),
				)
				return errcode.Error{
					Code: errcode.InternalError,
					Err:  errors.Wrap(err, "service.LocationRepo.GetByIDs"),
				}
			}
			locationIDs := locations.LocationIDs()
			for j, locationID := range userAccessPaths.LocationIDs() {
				if !golibs.InArrayString(locationID, locationIDs) {
					zapLogger.Error(
						"locations from client and BD is diff",
						zap.String("Function", "setUserAccessPaths"),
						zap.String("LocationID Client", locationID),
						zap.Strings("locationIDs DB", locationIDs),
					)
					return errcode.Error{
						Code:      errcode.InvalidData,
						FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].location", idx, j),
						Index:     idx,
					}
				}
			}
			for j, location := range locations {
				if location.IsArchived().Boolean() {
					zapLogger.Error(
						"locations is IsArchived",
						zap.String("Function", "setUserAccessPaths"),
						zap.String("locationID", location.LocationID().String()),
						zap.Bool("IsArchived", location.IsArchived().Boolean()),
					)
					return errcode.Error{
						Code:      errcode.InvalidData,
						FieldName: fmt.Sprintf("students[%d].enrollment_status_histories[%d].location", idx, j),
						Index:     idx,
					}
				}
			}
			studentsToCreate[idx].UserAccessPaths = userAccessPaths
			continue
		}
		locations, err := service.LocationRepo.GetByIDs(ctx, service.DB, student.UserAccessPaths.LocationIDs())
		if err != nil {
			zapLogger.Error(
				"cannot get locations",
				zap.Error(err),
				zap.String("Repo", "LocationRepo.GetByIDs"),
				zap.Strings("locationIDs", student.UserAccessPaths.LocationIDs()),
			)
			return errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.LocationRepo.GetByIDs"),
			}
		}

		if len(student.UserAccessPaths) != len(locations) {
			zapLogger.Error(
				"len of locations from client and BD is diff",
				zap.String("Function", "setUserAccessPaths"),
				zap.Int("len of LocationID Client", len(student.UserAccessPaths)),
				zap.Int("len of locationIDs DB", len(locations)),
			)
			return errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: fmt.Sprintf("students[%d].locations", idx),
				Index:     idx,
			}
		}
		userAccessPaths := entity.DomainUserAccessPaths{}
		for j, location := range locations {
			if location.IsArchived().Boolean() {
				zapLogger.Error(
					"location is Archived in old api",
					zap.String("Function", "setUserAccessPaths"),
					zap.String("LocationID", location.LocationID().String()),
					zap.Bool("IsArchived", location.IsArchived().Boolean()),
				)
				return errcode.Error{
					Code:      errcode.InvalidData,
					FieldName: fmt.Sprintf("students[%d].locations[%d]", idx, j),
					Index:     idx,
				}
			}
			userAccessPaths = append(userAccessPaths, entity.UserAccessPathWillBeDelegated{
				HasLocationID:     location,
				HasUserID:         student,
				HasOrganizationID: organization,
			})
		}

		studentsToCreate[idx].UserAccessPaths = userAccessPaths
	}
	return nil
}

func (service *DomainStudent) validateSchoolHistories(ctx context.Context, studentsToCreate ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	for _, student := range studentsToCreate {
		if len(student.SchoolHistories) == 0 {
			continue
		}

		schoolIDs := make([]string, 0, len(student.SchoolHistories))
		schoolIDBySchoolCourseID := map[string]string{}
		for idx, schoolHistory := range student.SchoolHistories {
			// validate empty schoolID
			schoolID := schoolHistory.SchoolID().String()
			if schoolID == "" {
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						EntityName: entity.StudentEntity,
						FieldName:  entity.StudentSchoolField,
						Index:      student.IndexAttr,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     idx,
				}
			}
			// validate duplicate schoolID
			if golibs.InArrayString(schoolID, schoolIDs) {
				return entity.DuplicatedFieldErrorWithArrayNestedField{
					DuplicatedFieldError: entity.DuplicatedFieldError{
						EntityName:      entity.StudentEntity,
						DuplicatedField: entity.StudentSchoolField,
						Index:           student.IndexAttr,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     idx,
				}
			}
			schoolIDs = append(schoolIDs, schoolID)

			if field.IsPresent(schoolHistory.StartDate()) && field.IsPresent(schoolHistory.EndDate()) {
				if schoolHistory.StartDate().Time().After(schoolHistory.EndDate().Time()) {
					return entity.InvalidFieldErrorWithArrayNestedField{
						InvalidFieldError: entity.InvalidFieldError{
							EntityName: entity.StudentEntity,
							FieldName:  entity.StartDateFieldEnrollmentStatusHistory,
							Index:      student.IndexAttr,
							Reason:     entity.StartDateAfterCurrentDate,
						},
						NestedFieldName: entity.StudentSchoolHistoryField,
						NestedIndex:     idx,
					}
				}
			}
			schoolCourseID := schoolHistory.SchoolCourseID().String()
			if schoolCourseID != "" {
				// validate duplicate school_courses
				if _, ok := schoolIDBySchoolCourseID[schoolCourseID]; ok {
					return entity.DuplicatedFieldErrorWithArrayNestedField{
						DuplicatedFieldError: entity.DuplicatedFieldError{
							EntityName:      entity.StudentEntity,
							DuplicatedField: entity.StudentSchoolCourseField,
							Index:           student.IndexAttr,
						},
						NestedFieldName: entity.StudentSchoolHistoryField,
						NestedIndex:     idx,
					}
				}
				schoolIDBySchoolCourseID[schoolCourseID] = schoolID
			}
		}

		schools, err := service.SchoolRepo.GetByIDs(ctx, service.DB, schoolIDs)
		if err != nil {
			zapLogger.Error(
				"cannot get schools",
				zap.Error(err),
				zap.String("Repo", "SchoolRepo.GetByIDs"),
				zap.Strings("schoolIDs", schoolIDs),
			)
			return err
		}

		existingSchoolIDs := schools.SchoolIDs()
		for idx, schoolID := range schoolIDs {
			if !golibs.InArrayString(schoolID, existingSchoolIDs) {
				return entity.NotFoundErrorWithArrayNestedField{
					NotFoundError: entity.NotFoundError{
						EntityName: entity.StudentEntity,
						FieldName:  entity.StudentSchoolField,
						FieldValue: schoolID,
						Index:      student.IndexAttr,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     idx,
				}
			}
		}

		levelWithSchool := map[string]struct{}{}
		for j, school := range schools {
			if school.IsArchived().Boolean() {
				zapLogger.Error(
					"school is Archived",
					zap.String("Function", "validateSchoolHistories"),
					zap.Bool("IsArchived", school.IsArchived().Boolean()),
					zap.String("SchoolID", school.SchoolID().String()),
				)
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						EntityName: entity.StudentEntity,
						FieldName:  entity.StudentSchoolField,
						Index:      student.IndexAttr,
						Reason:     entity.Archived,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     j,
				}
			}

			if _, ok := levelWithSchool[school.SchoolLevelID().String()]; ok {
				zapLogger.Error(
					"duplicated school level id",
					zap.String("Function", "validateSchoolHistories"),
					zap.String("SchoolLevelID", school.SchoolLevelID().String()),
				)
				return entity.DuplicatedFieldErrorWithArrayNestedField{
					DuplicatedFieldError: entity.DuplicatedFieldError{
						EntityName:      entity.StudentEntity,
						DuplicatedField: entity.StudentSchoolHistorySchoolLevel,
						Index:           student.IndexAttr,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     j,
				}
			}
			levelWithSchool[school.SchoolLevelID().String()] = struct{}{}
		}

		schoolCourseIDs := make([]string, 0, len(schoolIDBySchoolCourseID))
		for schoolCourseID := range schoolIDBySchoolCourseID {
			schoolCourseIDs = append(schoolCourseIDs, schoolCourseID)
		}
		schoolCourses, err := service.SchoolCourseRepo.GetByIDs(ctx, service.DB, schoolCourseIDs)
		if err != nil {
			zapLogger.Error(
				"cannot get school courses",
				zap.Error(err),
				zap.String("Repo", "SchoolCourseRepo.GetByIDs"),
				zap.Strings("schoolCourseIDs", schoolCourseIDs),
			)
			return err
		}
		// validate schoolCourse exist in DB
		existingSchoolCourseIDs := schoolCourses.SchoolCourseIDs()
		for _, schoolCourseID := range schoolCourseIDs {
			if !golibs.InArrayString(schoolCourseID, existingSchoolCourseIDs) {
				idx := utils.IndexOf(schoolIDs, schoolIDBySchoolCourseID[schoolCourseID])
				return entity.NotFoundErrorWithArrayNestedField{
					NotFoundError: entity.NotFoundError{
						EntityName: entity.StudentEntity,
						FieldName:  entity.StudentSchoolCourseField,
						FieldValue: schoolCourseID,
						Index:      student.IndexAttr,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     idx,
				}
			}
		}
		// validate schoolCourse belongs to school
		for _, schoolCourse := range schoolCourses {
			schoolCourseID := schoolCourse.SchoolCourseID().String()
			schoolID := schoolCourse.SchoolID().String()
			if _, ok := schoolIDBySchoolCourseID[schoolCourseID]; !ok || schoolIDBySchoolCourseID[schoolCourseID] != schoolID {
				idx := utils.IndexOf(schoolIDs, schoolIDBySchoolCourseID[schoolCourseID])
				zapLogger.Error(
					"school course does not belong to school",
					zap.String("Function", "validateSchoolHistories"),
					zap.String("schoolCourseID", schoolCourseID),
					zap.String("schoolID", schoolID),
					zap.String("schoolIDBySchoolCourseID[schoolCourseID]", schoolIDBySchoolCourseID[schoolCourseID]),
				)
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						EntityName: entity.StudentEntity,
						FieldName:  entity.StudentSchoolCourseField,
						Index:      student.IndexAttr,
						Reason:     entity.SchoolCourseDoesNotBelongToSchool,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     idx,
				}
			}
		}
	}
	return nil
}

func (service *DomainStudent) validateExternalUserIDExistedInSystem(ctx context.Context, users entity.Users) error {
	zapLogger := ctxzap.Extract(ctx)
	externalUserIDs := users.ExternalUserIDs()
	existingUsers, err := service.UserRepo.GetByExternalUserIDs(ctx, service.DB, externalUserIDs)
	if err != nil {
		zapLogger.Error(
			"cannot get users",
			zap.Error(err),
			zap.String("Repo", "UserRepo.GetByExternalUserIDs"),
			zap.Strings("externalUserIDs", externalUserIDs),
		)
		return err
	}

	existingStudents, err := service.StudentRepo.GetUsersByExternalUserIDs(ctx, service.DB, externalUserIDs)
	if err != nil {
		zapLogger.Error(
			"cannot get users",
			zap.Error(err),
			zap.String("Repo", "UserRepo.GetByExternalUserIDs"),
			zap.Strings("externalUserIDs", externalUserIDs),
		)
		return err
	}

	for _, user := range existingUsers {
		if user.ExternalUserID().String() == "" {
			continue
		}

		idxExternalUserIDByExistingUsers := utils.IndexOf(externalUserIDs, user.ExternalUserID().String())
		idxExternalUserIDByExistingStudents := utils.IndexOf(existingStudents.ExternalUserIDs(), user.ExternalUserID().String())

		if idxExternalUserIDByExistingUsers != -1 {
			if idxExternalUserIDByExistingStudents == -1 {
				return entity.ExistingDataError{
					FieldName:  string(entity.UserFieldExternalUserID),
					EntityName: entity.UserEntity,
					Index:      idxExternalUserIDByExistingUsers,
				}
			}
		}

		if user.UserID().Equal(users[idxExternalUserIDByExistingUsers].UserID()) {
			continue
		}
		zapLogger.Error(
			"existed external_user_id",
			zap.String("Function", "validateExternalUserIDExistedInSystem"),
			zap.Strings("existingUsers", users.UserIDs()),
			zap.Strings("externalUserIDs", externalUserIDs),
		)
		return entity.ExistingDataError{
			FieldName:  string(entity.UserFieldExternalUserID),
			EntityName: entity.UserEntity,
			Index:      idxExternalUserIDByExistingUsers,
		}
	}

	return nil
}

func validateStudentDuplicatedFields(students ...aggregate.DomainStudent) error {
	users := entity.Users{}
	for _, student := range students {
		users = append(users, student)
	}

	return ValidateUserDuplicatedFields(users)
}

func (service *DomainStudent) upsertSchoolHistories(ctx context.Context, db libdatabase.QueryExecer, students ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	var schoolHistories entity.DomainSchoolHistories

	studentIDs := []string{}

	for _, student := range students {
		studentIDs = append(studentIDs, student.UserID().String())
		if len(student.SchoolHistories) == 0 {
			continue
		}
		schoolIDs := student.SchoolHistories.SchoolIDs()
		schools, err := service.SchoolRepo.GetByIDsAndGradeID(ctx, db, schoolIDs, student.GradeID().String())
		if err != nil {
			zapLogger.Error(
				"cannot get schools",
				zap.Error(err),
				zap.String("Repo", "SchoolRepo.GetByIDsAndGradeID"),
				zap.Strings("userIDs", schoolIDs),
				zap.String("gradeID", student.GradeID().String()),
			)
			return errors.Wrap(err, "service.SchoolRepo.GetByIDsAndGradeID")
		}

		for _, schoolHistory := range student.SchoolHistories {
			isCurrentSchool := false
			if utils.IndexOf(schools.SchoolIDs(), schoolHistory.SchoolID().String()) != -1 {
				isCurrentSchool = true
			}

			schoolHistories = append(schoolHistories, entity.SchoolHistoryWillBeDelegated{
				SchoolHistoryAttribute: schoolHistory,
				HasSchoolInfoID:        schoolHistory,
				HasSchoolCourseID:      schoolHistory,
				HasUserID:              student,
				HasOrganizationID:      student,
				SchoolHistoryAttributeIsCurrentSchool: &entity.SchoolHistoryCurrentSchool{
					IsCurrentSchool: isCurrentSchool,
				},
			})
		}
	}

	if err := service.SchoolHistoryRepo.SoftDeleteByStudentIDs(ctx, db, studentIDs); err != nil {
		zapLogger.Error(
			"cannot soft delete school history",
			zap.Error(err),
			zap.String("Repo", "SchoolHistoryRepo.SoftDeleteByStudentIDs"),
		)
		return errors.Wrap(err, "service.SchoolHistoryRepo.SoftDeleteByStudentIDs")
	}

	if len(schoolHistories) == 0 {
		return nil
	}

	if err := service.SchoolHistoryRepo.UpsertMultiple(ctx, db, schoolHistories...); err != nil {
		zapLogger.Error(
			"cannot upsert multiple school history",
			zap.Error(err),
			zap.String("Repo", "SchoolHistoryRepo.UpsertMultiple"),
		)
		return errors.Wrap(err, "service.SchoolHistoryRepo.UpsertMultiple")
	}

	return nil
}

func upsertUserPhoneNumbers(ctx context.Context, db libdatabase.QueryExecer, userPhoneNumberRepo UserPhoneNumberRepo, users entity.Users, userPhoneNumbers entity.DomainUserPhoneNumbers) error {
	zapLogger := ctxzap.Extract(ctx)
	if err := userPhoneNumberRepo.SoftDeleteByUserIDs(ctx, db, users.UserIDs()); err != nil {
		zapLogger.Error(
			"cannot soft delete user phone numbers",
			zap.Error(err),
			zap.String("Repo", "UserPhoneNumberRepo.SoftDeleteByUserIDs"),
		)
		return err
	}

	if len(userPhoneNumbers) == 0 {
		return nil
	}

	if err := userPhoneNumberRepo.UpsertMultiple(ctx, db, userPhoneNumbers...); err != nil {
		zapLogger.Error(
			"cannot upsert multiple user phone numbers",
			zap.Error(err),
			zap.String("Repo", "UserPhoneNumberRepo.UpsertMultiple"),
		)
		return err
	}

	return nil
}

func (service *DomainStudent) upsertUserAddresses(ctx context.Context, db libdatabase.QueryExecer, students ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	var userAddresses entity.DomainUserAddresses
	userIDs := []string{}
	for _, student := range students {
		userIDs = append(userIDs, student.UserID().String())
		if student.UserAddress == nil {
			continue
		}
		userAddresses = append(userAddresses, entity.UserAddressWillBeDelegated{
			UserAddressAttribute: student.UserAddress,
			HasOrganizationID:    student,
			HasUserID:            student,
			HasPrefectureID:      student.UserAddress,
		})
	}

	if err := service.UserAddressRepo.SoftDeleteByUserIDs(ctx, db, userIDs); err != nil {
		zapLogger.Error(
			"cannot soft delete user addresses",
			zap.Error(err),
			zap.String("Repo", "UserAddressRepo.SoftDeleteByUserIDs"),
		)
		return err
	}

	if len(userAddresses) == 0 {
		return nil
	}

	if err := service.UserAddressRepo.UpsertMultiple(ctx, db, userAddresses...); err != nil {
		zapLogger.Error(
			"cannot upsert multiple user address",
			zap.Error(err),
			zap.String("Repo", "UserAddressRepo.UpsertMultiple"),
		)
		return err
	}

	return nil
}

func (service *DomainStudent) UpsertTaggedUsers(ctx context.Context, db libdatabase.QueryExecer, students ...aggregate.DomainStudent) error {
	zapLogger := ctxzap.Extract(ctx)
	var taggedUsers entity.DomainTaggedUsers
	userIDs := []string{}
	for _, student := range students {
		userIDs = append(userIDs, student.UserID().String())
		for _, taggedUser := range student.TaggedUsers {
			taggedUsers = append(taggedUsers, &entity.TaggedUserWillBeDelegated{
				HasTagID:          taggedUser,
				HasUserID:         student.DomainStudent,
				HasOrganizationID: student.DomainStudent,
			})
		}
	}

	if err := service.TaggedUserRepo.SoftDeleteByUserIDs(ctx, db, userIDs); err != nil {
		zapLogger.Error(
			"cannot soft delete tagged users",
			zap.Error(err),
			zap.String("Repo", "TaggedUserRepo.SoftDeleteByUserIDs"),
		)
		return errors.Wrap(err, "service.TaggedUserRepo.SoftDeleteByUserIDs")
	}

	if len(taggedUsers) == 0 {
		return nil
	}

	if err := service.TaggedUserRepo.UpsertBatch(ctx, db, taggedUsers...); err != nil {
		zapLogger.Error(
			"cannot upsert tagged users",
			zap.Error(err),
			zap.String("Repo", "TaggedUserRepo.UpsertBatch"),
		)
		return errors.Wrap(err, "service.TaggedUserRepo.UpsertMultiple")
	}

	return nil
}

func (service *DomainStudent) GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error) {
	users, err := service.UserRepo.GetByExternalUserIDs(ctx, service.DB, externalUserIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.UserRepo.GetByExternalUserIDs")
	}

	return users, nil
}

func (service *DomainStudent) GetSchoolsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchools, error) {
	schools, err := service.SchoolRepo.GetByPartnerInternalIDs(ctx, service.DB, externalIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.SchoolRepo.GetByPartnerInternalIDs")
	}

	return schools, nil
}

func (service *DomainStudent) GetSchoolCoursesByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error) {
	schoolCourses, err := service.SchoolCourseRepo.GetByPartnerInternalIDs(ctx, service.DB, externalIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.SchoolCourseRepo.GetByPartnerInternalIDs")
	}

	return schoolCourses, nil
}

func (service *DomainStudent) GetGradesByExternalIDs(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error) {
	grades, err := service.GradeRepo.GetByPartnerInternalIDs(ctx, service.DB, externalIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.GradeRepo.GetByPartnerInternalIDs")
	}

	return grades, nil
}

func (service *DomainStudent) GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error) {
	tags, err := service.TagRepo.GetByPartnerInternalIDs(ctx, service.DB, externalIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.TagRepo.GetByPartnerInternalIDs")
	}

	return tags, nil
}

func (service *DomainStudent) GetLocationsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainLocations, error) {
	locations, err := service.LocationRepo.GetByPartnerInternalIDs(ctx, service.DB, externalIDs)
	if err != nil {
		return nil, errors.Wrap(err, "service.LocationRepo.GetByPartnerInternalIDs")
	}

	return locations, nil
}

func (service *DomainStudent) GetPrefecturesByCodes(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error) {
	prefectures, err := service.PrefectureRepo.GetByPrefectureCodes(ctx, service.DB, codes)
	if err != nil {
		return nil, errors.Wrap(err, "service.PrefectureRepo.GetByPrefectureCodes")
	}

	return prefectures, nil
}

func (service *DomainStudent) validateStudentsLocations(ctx context.Context, db libdatabase.Ext, studentsToUpsert aggregate.DomainStudents) error {
	studentIDs := studentsToUpsert.StudentIDs()
	zapLogger := ctxzap.Extract(ctx).Sugar()
	if len(studentIDs) == 0 {
		return nil
	}

	lowestLevelLocations, err := service.LocationRepo.RetrieveLowestLevelLocations(ctx, db, "", 0, 0, nil)
	if err != nil {
		zapLogger.Error(
			"cannot get lowest locations",
			zap.Error(err),
			zap.String("Repo", "LocationRepo.RetrieveLowestLevelLocations"),
		)
		return err
	}

	for _, student := range studentsToUpsert {
		if err := validateLocationTypeForUpsertStudent(student.UserAccessPaths, lowestLevelLocations, student.IndexAttr); err != nil {
			zapLogger.Error(
				"compare locations from client and lowest level locations is wrong",
				zap.Error(err),
				zap.String("Function", "validateLocationTypeForUpsertStudent"),
				zap.Strings("student.UserAccessPaths.LocationIDs", student.UserAccessPaths.LocationIDs()),
				zap.Strings("lowestLevelLocations", lowestLevelLocations.LocationIDs()),
			)
			return err
		}
	}

	return nil
}

func validateLocationTypeForUpsertStudent(userAccessPaths entity.DomainUserAccessPaths, lowestLevelLocations entity.DomainLocations, idx int) error {
	for i, userAccessPath := range userAccessPaths {
		if !golibs.InArrayString(userAccessPath.LocationID().String(), lowestLevelLocations.LocationIDs()) {
			return entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentLocationTypeField,
					Reason:     entity.LocationIsNotLowestLocation,
					Index:      idx,
				},
				NestedFieldName: entity.StudentLocationsField,
				NestedIndex:     i,
			}
		}
	}
	return nil
}

func (service *DomainStudent) ValidateUpdateSystemAndExternalUserID(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error {
	zapLogger := ctxzap.Extract(ctx)
	userIDsToUpdate := studentsToUpdate.StudentIDs()
	existedUsers, err := service.UserRepo.GetByIDs(ctx, service.DB, userIDsToUpdate)
	if err != nil {
		zapLogger.Error(
			"cannot get users",
			zap.Error(err),
			zap.String("Repo", "UserRepo.GetByIDs"),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.UserRepo.GetByIDs"),
		}
	}

	if len(existedUsers) != len(userIDsToUpdate) {
		existedUserIDs := existedUsers.UserIDs()
		for idx, userID := range userIDsToUpdate {
			if !golibs.InArrayString(userID, existedUserIDs) {
				zapLogger.Error(
					"len(existedUsers) != len(userIDsToUpdate)",
					zap.Error(err),
				)
				return errcode.Error{
					Code:      errcode.InvalidData,
					FieldName: fmt.Sprintf("students[%d].user_id", idx),
					Index:     idx,
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
	for idx, student := range studentsToUpdate {
		for _, existedUser := range existedUsers {
			if !student.UserID().Equal(existedUser.UserID()) {
				continue
			}

			if existedUser.ExternalUserID().IsEmpty() {
				continue
			}

			if student.ExternalUserID() != existedUser.ExternalUserID() {
				return errcode.Error{
					Code:      errcode.UpdateFieldFail,
					FieldName: fmt.Sprintf("students[%d].external_user_id", idx),
					Index:     idx,
				}
			}
		}
	}

	return nil
}

func NewLegacyAuthUserUpserter(_ UserRepo, organizationRepo OrganizationRepo, firebaseTenantClient multitenant.TenantClient, identityPlatformTenantManager multitenant.TenantManager) AuthUserUpserter {
	return func(ctx context.Context, db libdatabase.QueryExecer, organization entity.DomainOrganization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) (entity.LegacyUsers, error) {
		zapLogger := ctxzap.Extract(ctx)

		tenantID, err := organizationRepo.GetTenantIDByOrgID(ctx, db, organization.OrganizationID().String())
		if err != nil {
			zapLogger.Error(
				"cannot get tenant id",
				zap.Error(err),
				zap.String("organizationID", organization.OrganizationID().String()),
			)
			return nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "new(repository.OrganizationRepo).GetTenantIDByOrgID"),
			}
		}

		backwardCompatibleAuthUsers, err := createUsersInIdentityPlatform(ctx, identityPlatformTenantManager, tenantID, int64(organization.SchoolID().Int32()), usersToCreate...)
		if err != nil {
			zapLogger.Error(
				"cannot create users on identity platform",
				zap.Error(err),
				zap.String("organizationID", organization.OrganizationID().String()),
				zap.String("tenantID", tenantID),
				zap.Strings("emails", backwardCompatibleAuthUsers.Limit(10).Emails()),
			)
			return nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.createUsersInIdentityPlatform"),
			}
		}
		// enable username -> don't update login email on identity platform to email
		if option.EnableUsername {
			return nil, nil
		}
		if option.EnableIgnoreUpdateEmail {
			return nil, nil
		}

		if err := updateUserEmailsInIdentityPlatform(ctx, identityPlatformTenantManager, tenantID, usersToUpdate...); err != nil {
			zapLogger.Error(
				"cannot update users on identity platform",
				zap.Error(err),
				zap.Int64("organizationID", int64(organization.SchoolID().Int32())),
				zap.String("tenantID", tenantID),
			)
			return nil, errcode.Error{
				Code: errcode.InternalError,
				Err:  errors.Wrap(err, "service.updateUserEmailsInIdentityPlatform"),
			}
		}
		return nil, nil
	}
}

//lint:ignore U1000 Ignore this unused function for quick reverting if needed
func (service *DomainStudent) upsertUserInIdentityPlatform(ctx context.Context, tx libdatabase.Ext, organization *interceptors.Organization, usersToCreate entity.Users, usersToUpdate entity.Users, option unleash.DomainUserFeatureOption) error {
	zapLogger := ctxzap.Extract(ctx)

	tenantID, err := service.OrganizationRepo.GetTenantIDByOrgID(ctx, tx, organization.OrganizationID().String())
	if err != nil {
		zapLogger.Error(
			"cannot get tenant id",
			zap.Error(err),
			zap.String("organizationID", organization.OrganizationID().String()),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "new(repository.OrganizationRepo).GetTenantIDByOrgID"),
		}
	}

	backwardCompatibleAuthUsers, err := createUsersInIdentityPlatform(ctx, service.TenantManager, tenantID, int64(organization.SchoolID().Int32()), usersToCreate...)
	if err != nil {
		zapLogger.Error(
			"cannot create users on identity platform",
			zap.Error(err),
			zap.String("organizationID", organization.OrganizationID().String()),
			zap.String("tenantID", tenantID),
			zap.Strings("emails", backwardCompatibleAuthUsers.Limit(10).Emails()),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.createUsersInIdentityPlatform"),
		}
	}

	if option.EnableIgnoreUpdateEmail {
		return nil
	}

	if err := updateUserEmailsInIdentityPlatform(ctx, service.TenantManager, tenantID, usersToUpdate...); err != nil {
		zapLogger.Error(
			"cannot update users on identity platform",
			zap.Error(err),
			zap.Int64("organizationID", int64(organization.SchoolID().Int32())),
			zap.String("tenantID", tenantID),
		)
		return errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.updateUserEmailsInIdentityPlatform"),
		}
	}
	return nil
}

func (service *DomainStudent) GetEmailWithStudentID(ctx context.Context, studentIDs []string) (map[string]entity.User, error) {
	users, err := service.UserRepo.GetByIDs(ctx, service.DB, studentIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  fmt.Errorf("service.UserRepo.GetByIDs: %w", err),
		}
	}

	userByUserID := make(map[string]entity.User)
	for _, user := range users {
		userByUserID[user.UserID().String()] = user
	}
	return userByUserID, nil
}

func (service *DomainStudent) IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization valueobj.HasOrganizationID) bool {
	return unleash.IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(service.UnleashClient, service.Env, organization)
}

func (service *DomainStudent) IsFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled(organization valueobj.HasOrganizationID) bool {
	isFeatureEnabled, err := service.UnleashClient.IsFeatureEnabledOnOrganization(featureToggleIgnoreInvalidRecordsCSVAndOpenAPI, service.Env, organization.OrganizationID().String())
	if err != nil {
		isFeatureEnabled = false
	}
	return isFeatureEnabled
}

func (service *DomainStudent) IsExperimentalBulkInsertEnrollmentStatusHistories(organization valueobj.HasOrganizationID) bool {
	return unleash.IsExperimentalBulkInsertEnrollmentStatusHistories(service.UnleashClient, service.Env, organization)
}

func (service *DomainStudent) IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool {
	return unleash.IsFeatureUserNameStudentParentEnabled(service.UnleashClient, service.Env, organization)
}

func (service *DomainStudent) IsFeatureIgnoreInvalidRecordsOpenAPIEnabled(organization valueobj.HasOrganizationID) bool {
	return unleash.IsFeatureIgnoreInvalidRecordsOpenAPI(service.UnleashClient, service.Env, organization)
}

func (service *DomainStudent) GetExistingUsersByExternalUserIDs(ctx context.Context, users entity.Users) (map[string]entity.User, error) {
	existingUsers, err := service.UserRepo.GetByExternalUserIDs(ctx, service.DB, users.ExternalUserIDs())

	mapExternalUserIDAndUser := make(map[string]entity.User, len(existingUsers))

	if err != nil {
		return mapExternalUserIDAndUser, errors.Wrap(err, "service.GetExistingUsersByExternalUserIDs.GetByExternalUserIDs")
	}
	for _, user := range existingUsers {
		mapExternalUserIDAndUser[user.ExternalUserID().String()] = user
	}
	return mapExternalUserIDAndUser, nil
}

func (service *DomainStudent) IsDisableAutoDeactivateStudents(organization valueobj.HasOrganizationID) bool {
	return unleash.IsDisableAutoDeactivateStudents(service.UnleashClient, service.Env, organization)
}

func (service *DomainStudent) IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error) {
	config, err := service.InternalConfigurationRepo.GetByKey(ctx, service.DB, constant.KeyAuthUsernameConfig)
	if err != nil {
		if strings.Contains(err.Error(), pgx.ErrNoRows.Error()) {
			return false, nil
		}
		return false, err
	}
	return config.ConfigValue().String() == constant.ConfigValueOn, nil
}
