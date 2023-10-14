package service

import (
	"context"
	"regexp"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	libdatabase "github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	upb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/pkg/errors"
)

var allowListInOrderFlow = []string{
	upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_POTENTIAL.String(),
	upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_TEMPORARY.String(),
	upb.StudentEnrollmentStatus_STUDENT_ENROLLMENT_STATUS_NON_POTENTIAL.String(),
}

type StudentValidationManager struct {
	UserRepo interface {
		GetByUserNames(ctx context.Context, db libdatabase.QueryExecer, usernames []string) (entity.Users, error)
		GetByEmails(ctx context.Context, db libdatabase.QueryExecer, emails []string) (entity.Users, error)
		GetByEmailsInsensitiveCase(ctx context.Context, db libdatabase.QueryExecer, emails []string) (entity.Users, error)
		GetByExternalUserIDs(ctx context.Context, db libdatabase.QueryExecer, externalUserIDs []string) (entity.Users, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (entity.Users, error)
	}
	UserGroupRepo interface {
		FindUserGroupByRoleName(ctx context.Context, db libdatabase.QueryExecer, roleName string) (entity.DomainUserGroup, error)
	}
	LocationRepo interface {
		RetrieveLowestLevelLocations(ctx context.Context, db libdatabase.Ext, name string, limit int32, offset int32, locationIDs []string) (entity.DomainLocations, error)
	}
	GradeRepo interface {
		GetAll(ctx context.Context, db libdatabase.QueryExecer) ([]entity.DomainGrade, error)
	}
	SchoolRepo interface {
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) (entity.DomainSchools, error)
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainSchools, error)
	}
	SchoolCourseRepo interface {
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainSchoolCourses, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, schoolCourseIDs []string) (entity.DomainSchoolCourses, error)
	}
	PrefectureRepo interface {
		GetByPrefectureCodes(ctx context.Context, db libdatabase.QueryExecer, prefectureCodes []string) (entity.DomainPrefectures, error)
	}
	TagRepo interface {
		GetByPartnerInternalIDs(ctx context.Context, db libdatabase.QueryExecer, partnerInternalIDs []string) (entity.DomainTags, error)
		GetByIDs(ctx context.Context, db libdatabase.QueryExecer, ids []string) (entity.DomainTags, error)
	}
	InternalConfigurationRepo interface {
		GetByKey(ctx context.Context, db libdatabase.QueryExecer, configKey string) (entity.DomainConfiguration, error)
	}
	EnrollmentStatusHistoryRepo interface {
		GetByStudentIDs(ctx context.Context, db libdatabase.QueryExecer, studentIDs []string) (entity.DomainEnrollmentStatusHistories, error)
	}
	StudentRepo interface {
		GetUsersByExternalUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (entity.Users, error)
	}
}

//gocyclo:ignore
func (manager *StudentValidationManager) FullyValidate(ctx context.Context, db libdatabase.Ext, domainStudents aggregate.DomainStudents, isEnableUsername bool) (aggregate.DomainStudents, aggregate.DomainStudents, []error) {
	uniqueDomainStudents, errorCollections := removeDuplicatedFields(domainStudents)
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		e := entity.InternalError{
			RawErr: errors.Wrap(err, "interceptors.OrganizationFromContext"),
		}
		return nil, nil, append(errorCollections, e)
	}

	currentUser, err := manager.GetCurrentUser(ctx, db)
	if err != nil {
		errorCollections = append(errorCollections, err)
		return nil, nil, errorCollections
	}

	isOrderFlow, err := manager.IsInOrderFlow(ctx, db)
	if err != nil {
		errorCollections = append(errorCollections, err)
		return nil, nil, errorCollections
	}

	existingEntitiesRelatedStudent, err := manager.GetExistingEntitiesRelatedStudent(ctx, db, uniqueDomainStudents)
	if err != nil {
		errorCollections = append(errorCollections, err)
		return nil, nil, errorCollections
	}

	var studentsToCreate, studentsToUpdate aggregate.DomainStudents
	// Validation both create and update student
	for _, uniqueDomainStudent := range uniqueDomainStudents {
		// validation of gender, last name, first name, email, username
		if err := validateBothCreateAndUpdate(uniqueDomainStudent); err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}
		domainGrade, err := validateGrade(uniqueDomainStudent, existingEntitiesRelatedStudent.MapPartnerIDAndGrade)
		if err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}

		domainTags, err := validateTag(uniqueDomainStudent, existingEntitiesRelatedStudent.MapPartnerIDAndTag)
		if err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}

		domainSchoolHistories, err := validateSchoolHistories(uniqueDomainStudent, existingEntitiesRelatedStudent.MapPartnerIDAndSchool, existingEntitiesRelatedStudent.MapPartnerIDAndSchoolCourse)
		if err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}

		domainLocations, enrollmentStatusHistories, err := validateLocations(uniqueDomainStudent, existingEntitiesRelatedStudent.MapPartnerIDAndLowestLocation)
		if err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}

		domainUserAddress, err := validateUserAddress(uniqueDomainStudent, existingEntitiesRelatedStudent.MapPrefectureCodeAndPrefecture)
		if err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}

		domainPhoneNumbers, err := validatePhoneNumbers(uniqueDomainStudent)
		if err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}

		uniqueDomainStudent.Locations = domainLocations
		uniqueDomainStudent.EnrollmentStatusHistories = enrollmentStatusHistories
		uniqueDomainStudent.SchoolHistories = domainSchoolHistories
		uniqueDomainStudent.UserAddress = domainUserAddress
		uniqueDomainStudent.UserPhoneNumbers = domainPhoneNumbers
		uniqueDomainStudent.Tags = domainTags

		if err := validateEnrollmentStatusHistories(uniqueDomainStudent, isOrderFlow); err != nil {
			errorCollections = append(errorCollections, err)
			continue
		}
		// Validation only create student
		if uniqueDomainStudent.UserID().IsEmpty() {
			if err := validateExternalUserIDForCreating(uniqueDomainStudent, existingEntitiesRelatedStudent.MapExternalUserIDAndUser); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			if err := validateEmailForCreating(uniqueDomainStudent, existingEntitiesRelatedStudent.MapEmailAndUser); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			if err := validateUserNameForCreating(uniqueDomainStudent, existingEntitiesRelatedStudent.MapUserNameAndUser); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			if err := validateEnrollmentStatusHistoriesForCreating(uniqueDomainStudent); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			userID := field.NewString(idutil.ULIDNow())
			var userProfileLoginEmail valueobj.HasLoginEmail = uniqueDomainStudent.DomainStudent
			if isEnableUsername {
				userProfileLoginEmail = &entity.UserProfileLoginEmailDelegate{
					Email: userID.String() + constant.LoginEmailPostfix,
				}
			}
			uniqueDomainStudent.DomainStudent = entity.StudentWillBeDelegated{
				DomainStudentProfile: uniqueDomainStudent.DomainStudent,
				HasOrganizationID:    organization,
				HasUserID: &valueobj.RandomHasUserID{
					RandomUserID: userID,
				},
				HasGradeID:    domainGrade,
				HasSchoolID:   organization,
				HasCountry:    currentUser,
				HasLoginEmail: userProfileLoginEmail,
			}

			uniqueDomainStudent.UserAccessPaths = uniqueDomainStudent.Locations.ToUserAccessPath(uniqueDomainStudent)

			for idx, enrollmentStatusHistory := range uniqueDomainStudent.EnrollmentStatusHistories {
				uniqueDomainStudent.EnrollmentStatusHistories[idx] = entity.EnrollmentStatusHistoryWillBeDelegated{
					EnrollmentStatusHistory: enrollmentStatusHistory,
					HasUserID: &valueobj.RandomHasUserID{
						RandomUserID: userID,
					},
					HasLocationID:     enrollmentStatusHistory,
					HasOrganizationID: organization,
				}
			}

			for idx, userPhoneNumber := range uniqueDomainStudent.UserPhoneNumbers {
				uniqueDomainStudent.UserPhoneNumbers[idx] = entity.UserPhoneNumberWillBeDelegated{
					UserPhoneNumberAttribute: userPhoneNumber,
					HasUserID: &valueobj.RandomHasUserID{
						RandomUserID: userID,
					},
					HasOrganizationID: organization,
				}
			}

			uniqueDomainStudent.UserGroupMembers = append(uniqueDomainStudent.UserGroupMembers, entity.UserGroupMemberWillBeDelegated{
				HasUserGroupID:    existingEntitiesRelatedStudent.DomainUserGroup,
				HasUserID:         uniqueDomainStudent.DomainStudent,
				HasOrganizationID: organization,
			})
			uniqueDomainStudent.TaggedUsers = domainTags.ToTaggedUser(uniqueDomainStudent)
			studentsToCreate = append(studentsToCreate, uniqueDomainStudent)
		} else {
			// Validation only update student
			// Add user_id and organization_id to enrollment_status_history to validate
			for idx, enrollmentStatusHistory := range uniqueDomainStudent.EnrollmentStatusHistories {
				uniqueDomainStudent.EnrollmentStatusHistories[idx] = entity.EnrollmentStatusHistoryWillBeDelegated{
					EnrollmentStatusHistory: enrollmentStatusHistory,
					HasUserID:               uniqueDomainStudent.DomainStudent,
					HasLocationID:           enrollmentStatusHistory,
					HasOrganizationID:       organization,
				}
			}
			if err := validateUserIDForUpdating(uniqueDomainStudent, existingEntitiesRelatedStudent.MapUserIDAndUser); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			if err := validateExternalUserIDForUpdating(uniqueDomainStudent, existingEntitiesRelatedStudent.MapUserIDAndUser, existingEntitiesRelatedStudent.MapExternalUserIDAndUser, existingEntitiesRelatedStudent.MapExternalUserIDAndStudentUser); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			if err := validateEmailForUpdating(uniqueDomainStudent, existingEntitiesRelatedStudent.MapEmailAndUser); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			if err := validateUserNameForUpdating(uniqueDomainStudent, existingEntitiesRelatedStudent.MapUserNameAndUser); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			if err := validateEnrollmentStatusHistoriesForUpdating(uniqueDomainStudent, existingEntitiesRelatedStudent.ExistingEnrollmentStatusHistories, isOrderFlow); err != nil {
				errorCollections = append(errorCollections, err)
				continue
			}

			// Add organization to user_phone_number
			for idx, userPhoneNumber := range uniqueDomainStudent.UserPhoneNumbers {
				uniqueDomainStudent.UserPhoneNumbers[idx] = entity.UserPhoneNumberWillBeDelegated{
					UserPhoneNumberAttribute: userPhoneNumber,
					HasUserID:                uniqueDomainStudent.DomainStudent,
					HasOrganizationID:        organization,
				}
			}
			// Add organization to student
			uniqueDomainStudent.DomainStudent = entity.StudentWillBeDelegated{
				DomainStudentProfile: uniqueDomainStudent.DomainStudent,
				HasOrganizationID:    organization,
				HasUserID:            uniqueDomainStudent.DomainStudent,
				HasGradeID:           domainGrade,
				HasCountry:           currentUser,
				HasSchoolID:          organization,
				HasLoginEmail:        uniqueDomainStudent.DomainStudent,
			}
			uniqueDomainStudent.UserAccessPaths = uniqueDomainStudent.Locations.ToUserAccessPath(uniqueDomainStudent)
			uniqueDomainStudent.TaggedUsers = domainTags.ToTaggedUser(uniqueDomainStudent)

			// Reassign user access path for Nats
			if len(uniqueDomainStudent.UserAccessPaths) == 0 {
				enrollmentStatusHistoriesWithUniqLocation := existingEntitiesRelatedStudent.ExistingEnrollmentStatusHistories.GetByUserIDWithUniqLocation(uniqueDomainStudent.UserID())
				for _, enrollmentStatusHistory := range enrollmentStatusHistoriesWithUniqLocation {
					uniqueDomainStudent.UserAccessPaths = append(uniqueDomainStudent.UserAccessPaths, entity.UserAccessPathWillBeDelegated{
						HasUserID:         uniqueDomainStudent.DomainStudent,
						HasLocationID:     enrollmentStatusHistory,
						HasOrganizationID: organization,
					})
				}
			}

			uniqueDomainStudent.UserGroupMembers = append(uniqueDomainStudent.UserGroupMembers, entity.UserGroupMemberWillBeDelegated{
				HasUserGroupID:    existingEntitiesRelatedStudent.DomainUserGroup,
				HasUserID:         uniqueDomainStudent.DomainStudent,
				HasOrganizationID: organization,
			})
			studentsToUpdate = append(studentsToUpdate, uniqueDomainStudent)
		}
	}
	return studentsToCreate, studentsToUpdate, errorCollections
}

func (manager *StudentValidationManager) GetExistingEntitiesRelatedStudent(ctx context.Context, db libdatabase.Ext, domainStudents aggregate.DomainStudents) (aggregate.ExistingEntitiesRelatedStudent, error) {
	existingEntitiesRelatedStudent := aggregate.ExistingEntitiesRelatedStudent{}

	mapPartnerIDAndGrade, err := manager.GetMapPartnerIDAndGrade(ctx, db)
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapPartnerIDAndTag, err := manager.GetMapPartnerIDAndTag(ctx, db, domainStudents.TagPartnerIDs())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapPartnerIDAndSchool, err := manager.GetMapPartnerIDAndSchool(ctx, db, domainStudents.SchoolPartnerIDs())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapPartnerIDAndSchoolCourse, err := manager.GetMapPartnerIDAndSchoolCourse(ctx, db, domainStudents.SchoolCoursePartnerIDs())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapEmailAndUser, err := manager.GetMapEmailAndUserByEmails(ctx, db, domainStudents.Users().Emails())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapUserIDAndUser, err := manager.GetMapUserIDAndUserByUserIDs(ctx, db, domainStudents.Users().UserIDs())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapExternalUserIDAndUser, err := manager.GetMapExternalUserIDAndUserByExternalUserIDs(ctx, db, domainStudents.Users().ExternalUserIDs())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapExternalUserIDAndStudentUser, err := manager.GetMapExternalUserIDAndStudentUserByExternalUserIDs(ctx, db, domainStudents.Users().ExternalUserIDs())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapPartnerIDAndLowestLocation, err := manager.GetMapPartnerIDAndLowestLocation(ctx, db)
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	existingEnrollmentStatusHistories, err := manager.EnrollmentStatusHistoryRepo.GetByStudentIDs(ctx, db, domainStudents.StudentIDs())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}
	mapPrefectureCodeAndPrefecture, err := manager.GetMapPrefectureCodeAndPrefecture(ctx, db, domainStudents.PrefectureCodes())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	mapUserNameAndUser, err := manager.GetMapUserNameAndUserByUserName(ctx, db, domainStudents.Users().LowerCasedUserNames())
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	domainUserGroup, err := manager.UserGroupRepo.FindUserGroupByRoleName(ctx, db, constant.RoleStudent)
	if err != nil {
		return existingEntitiesRelatedStudent, err
	}

	existingEntitiesRelatedStudent = aggregate.ExistingEntitiesRelatedStudent{
		MapPartnerIDAndGrade:              mapPartnerIDAndGrade,
		MapPartnerIDAndTag:                mapPartnerIDAndTag,
		MapPartnerIDAndSchool:             mapPartnerIDAndSchool,
		MapPartnerIDAndSchoolCourse:       mapPartnerIDAndSchoolCourse,
		MapEmailAndUser:                   mapEmailAndUser,
		MapUserIDAndUser:                  mapUserIDAndUser,
		MapExternalUserIDAndUser:          mapExternalUserIDAndUser,
		MapExternalUserIDAndStudentUser:   mapExternalUserIDAndStudentUser,
		MapPartnerIDAndLowestLocation:     mapPartnerIDAndLowestLocation,
		ExistingEnrollmentStatusHistories: existingEnrollmentStatusHistories,
		MapPrefectureCodeAndPrefecture:    mapPrefectureCodeAndPrefecture,
		MapUserNameAndUser:                mapUserNameAndUser,
		DomainUserGroup:                   domainUserGroup,
	}

	return existingEntitiesRelatedStudent, nil
}

func (manager *StudentValidationManager) GetMapPartnerIDAndGrade(ctx context.Context, db libdatabase.QueryExecer) (map[string]entity.DomainGrade, error) {
	grades, err := manager.GradeRepo.GetAll(ctx, db)
	if err != nil {
		return nil, err
	}

	mapPartnerIDAndGrade := make(map[string]entity.DomainGrade)
	for _, grade := range grades {
		mapPartnerIDAndGrade[grade.PartnerInternalID().String()] = grade
	}
	return mapPartnerIDAndGrade, nil
}

func (manager *StudentValidationManager) GetMapEmailAndUserByEmails(ctx context.Context, db libdatabase.QueryExecer, emails []string) (map[string]entity.User, error) {
	existingUsers, err := manager.UserRepo.GetByEmailsInsensitiveCase(ctx, db, emails)
	if err != nil {
		return nil, err
	}
	mapEmailAndUser := make(map[string]entity.User, len(existingUsers))
	for _, user := range existingUsers {
		mapEmailAndUser[user.Email().String()] = user
	}
	return mapEmailAndUser, nil
}

func (manager *StudentValidationManager) GetMapUserNameAndUserByUserName(ctx context.Context, db libdatabase.QueryExecer, usernames []string) (map[string]entity.User, error) {
	existingUsers, err := manager.UserRepo.GetByUserNames(ctx, db, usernames)
	if err != nil {
		return nil, err
	}
	mapUserNameAndUser := make(map[string]entity.User, len(existingUsers))
	for _, existingUser := range existingUsers {
		mapUserNameAndUser[existingUser.UserName().String()] = existingUser
	}
	return mapUserNameAndUser, nil
}

func (manager *StudentValidationManager) GetMapUserIDAndUserByUserIDs(ctx context.Context, db libdatabase.QueryExecer, userIDs []string) (map[string]entity.User, error) {
	existingUsers, err := manager.UserRepo.GetByIDs(ctx, db, userIDs)
	if err != nil {
		return nil, err
	}

	mapUserIDAndUser := make(map[string]entity.User, len(existingUsers))
	for _, user := range existingUsers {
		mapUserIDAndUser[user.UserID().String()] = user
	}
	return mapUserIDAndUser, nil
}

func (manager *StudentValidationManager) GetMapPartnerIDAndLowestLocation(ctx context.Context, db libdatabase.Ext) (map[string]entity.DomainLocation, error) {
	mapPartnerIDAndLowestLocation := make(map[string]entity.DomainLocation)
	lowestLevelLocations, err := manager.LocationRepo.RetrieveLowestLevelLocations(ctx, db, "", 0, 0, nil)
	if err != nil {
		return nil, err
	}
	for _, location := range lowestLevelLocations {
		mapPartnerIDAndLowestLocation[location.PartnerInternalID().String()] = location
	}
	return mapPartnerIDAndLowestLocation, nil
}

func (manager *StudentValidationManager) GetMapExternalUserIDAndUserByExternalUserIDs(ctx context.Context, db libdatabase.QueryExecer, externalUserIDs []string) (map[string]entity.User, error) {
	existingUsers, err := manager.UserRepo.GetByExternalUserIDs(ctx, db, externalUserIDs)
	if err != nil {
		return nil, err
	}

	mapExternalUserIDAndUser := make(map[string]entity.User, len(existingUsers))
	for _, user := range existingUsers {
		mapExternalUserIDAndUser[user.ExternalUserID().String()] = user
	}
	return mapExternalUserIDAndUser, nil
}

func (manager *StudentValidationManager) GetMapExternalUserIDAndStudentUserByExternalUserIDs(ctx context.Context, db libdatabase.Ext, ds []string) (map[string]entity.User, error) {
	existingUsers, err := manager.StudentRepo.GetUsersByExternalUserIDs(ctx, db, ds)
	if err != nil {
		return nil, err
	}

	mapExternalUserIDAndUser := make(map[string]entity.User, len(existingUsers))
	for _, user := range existingUsers {
		mapExternalUserIDAndUser[user.ExternalUserID().String()] = user
	}
	return mapExternalUserIDAndUser, nil
}

func (manager *StudentValidationManager) IsInOrderFlow(ctx context.Context, db libdatabase.Ext) (bool, error) {
	config, err := manager.InternalConfigurationRepo.GetByKey(ctx, db, constant.KeyEnrollmentStatusHistoryConfig)
	if err != nil {
		return false, err
	}
	return config.ConfigValue().String() == constant.ConfigValueOff, nil
}

func (manager *StudentValidationManager) GetMapPartnerIDAndTag(ctx context.Context, db libdatabase.QueryExecer, tagPartnerIDs []string) (map[string]entity.DomainTag, error) {
	tags, err := manager.TagRepo.GetByPartnerInternalIDs(ctx, db, tagPartnerIDs)
	if err != nil {
		return nil, err
	}

	mapPartnerIDAndTag := make(map[string]entity.DomainTag)
	for _, tag := range tags {
		mapPartnerIDAndTag[tag.PartnerInternalID().String()] = tag
	}
	return mapPartnerIDAndTag, nil
}

func (manager *StudentValidationManager) GetMapPartnerIDAndSchool(ctx context.Context, db libdatabase.QueryExecer, schoolPartnerIDs []string) (map[string]entity.DomainSchool, error) {
	schools, err := manager.SchoolRepo.GetByPartnerInternalIDs(ctx, db, schoolPartnerIDs)

	if err != nil {
		return nil, err
	}

	mapPartnerIDAndSchool := make(map[string]entity.DomainSchool)
	for _, school := range schools {
		mapPartnerIDAndSchool[school.PartnerInternalID().String()] = school
	}
	return mapPartnerIDAndSchool, nil
}

func (manager *StudentValidationManager) GetMapPartnerIDAndSchoolCourse(ctx context.Context, db libdatabase.QueryExecer, schoolPartnerIDs []string) (map[string]entity.DomainSchoolCourse, error) {
	schoolCourses, err := manager.SchoolCourseRepo.GetByPartnerInternalIDs(ctx, db, schoolPartnerIDs)
	if err != nil {
		return nil, err
	}

	mapPartnerIDAndSchoolCourse := make(map[string]entity.DomainSchoolCourse)
	for _, schoolCourse := range schoolCourses {
		mapPartnerIDAndSchoolCourse[schoolCourse.PartnerInternalID().String()] = schoolCourse
	}
	return mapPartnerIDAndSchoolCourse, nil
}

func (manager *StudentValidationManager) GetCurrentUser(ctx context.Context, db libdatabase.QueryExecer) (entity.User, error) {
	currentUserID := interceptors.UserIDFromContext(ctx)
	users, err := manager.UserRepo.GetByIDs(ctx, db, []string{currentUserID})
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return nil, entity.NotFoundError{
			EntityName: entity.UserEntity,
			FieldValue: currentUserID,
			FieldName:  string(entity.UserFieldUserID),
		}
	}
	return users[0], nil
}

func (manager *StudentValidationManager) GetMapPrefectureCodeAndPrefecture(ctx context.Context, db libdatabase.QueryExecer, prefectureCodes []string) (map[string]entity.DomainPrefecture, error) {
	if len(prefectureCodes) == 0 {
		return nil, nil
	}
	prefectures, err := manager.PrefectureRepo.GetByPrefectureCodes(ctx, db, prefectureCodes)
	if err != nil {
		return nil, err
	}

	mapPrefectureCodeAndPrefecture := make(map[string]entity.DomainPrefecture)
	for _, prefecture := range prefectures {
		mapPrefectureCodeAndPrefecture[prefecture.PrefectureCode().String()] = prefecture
	}
	return mapPrefectureCodeAndPrefecture, nil
}

func validateBothCreateAndUpdate(student aggregate.DomainStudent) error {
	if err := entity.ValidateUserGender(student); err != nil {
		return err
	}
	if err := entity.ValidateUserFirstName(student); err != nil {
		return err
	}
	if err := entity.ValidateUserLastName(student); err != nil {
		return err
	}
	// Validate required fields and pattern email
	if err := entity.ValidateUserEmail(student); err != nil {
		return err
	}
	if err := entity.ValidateStudentContactPreference(student); err != nil {
		return err
	}
	return entity.ValidateUserName(student)
}

func removeDuplicatedFields(domainStudents aggregate.DomainStudents) (aggregate.DomainStudents, []error) {
	uniqueDomainStudents, errorCollections := removeDuplicatedUserID(domainStudents)

	uniqueDomainStudents, listOfErrors := removeDuplicatedExternalUserID(uniqueDomainStudents)
	if len(listOfErrors) > 0 {
		errorCollections = append(errorCollections, listOfErrors...)
	}
	uniqueDomainStudents, listOfErrors = removeDuplicatedEmail(uniqueDomainStudents)
	if len(listOfErrors) > 0 {
		errorCollections = append(errorCollections, listOfErrors...)
	}
	uniqueDomainStudents, listOfErrors = removeDuplicatedUserName(uniqueDomainStudents)
	if len(listOfErrors) > 0 {
		errorCollections = append(errorCollections, listOfErrors...)
	}
	return uniqueDomainStudents, errorCollections
}

func validateGrade(student aggregate.DomainStudent, mapPartnerIDAndGrade map[string]entity.DomainGrade) (entity.DomainGrade, error) {
	if student.Grade.PartnerInternalID().IsEmpty() {
		return nil, entity.MissingMandatoryFieldError{
			FieldName:  entity.StudentGradeField,
			EntityName: entity.StudentEntity,
			Index:      student.IndexAttr,
		}
	}

	grade, ok := mapPartnerIDAndGrade[student.Grade.PartnerInternalID().String()]
	if !ok {
		return nil, entity.NotFoundError{
			EntityName: entity.StudentEntity,
			FieldName:  entity.StudentGradeField,
			Index:      student.IndexAttr,
			FieldValue: student.Grade.PartnerInternalID().String(),
		}
	}
	return grade, nil
}

func validateLocations(student aggregate.DomainStudent, mapPartnerIDAndLowestLocation map[string]entity.DomainLocation) (entity.DomainLocations, entity.DomainEnrollmentStatusHistories, error) {
	// Duplicate location_id
	if index := getDuplicatedIndex(student.Locations.PartnerInternalIDs()); index > -1 {
		return nil, nil, entity.DuplicatedFieldError{
			DuplicatedField: entity.StudentLocationsField,
			Index:           student.IndexAttr,
			EntityName:      entity.StudentEntity,
		}
	}

	domainLocations := make(entity.DomainLocations, 0, len(student.Locations.PartnerInternalIDs()))
	for _, partnerInternalID := range student.Locations.PartnerInternalIDs() {
		if partnerInternalID == "" {
			return nil, nil, entity.InvalidFieldError{
				FieldName:  entity.StudentLocationsField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				Reason:     entity.Empty,
			}
		}
		// Not found location as lowest location in system
		location, ok := mapPartnerIDAndLowestLocation[partnerInternalID]
		if !ok {
			return nil, nil, entity.NotFoundError{
				FieldName:  entity.StudentLocationsField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
			}
		}
		domainLocations = append(domainLocations, location)
	}

	enrollmentStatusHistories := make(entity.DomainEnrollmentStatusHistories, 0, len(domainLocations))
	for idx, location := range domainLocations {
		// Imposable to have empty enrollment status history when updating student
		if len(student.EnrollmentStatusHistories) != 0 {
			enrollmentStatusHistories = append(enrollmentStatusHistories, entity.EnrollmentStatusHistoryWillBeDelegated{
				// Ensure location and enrollment status history are in the same order in port layer
				EnrollmentStatusHistory: student.EnrollmentStatusHistories[idx],
				HasLocationID:           location,
			})
		}
	}
	return domainLocations, enrollmentStatusHistories, nil
}

func validateTag(student aggregate.DomainStudent, mapPartnerIDAndTag map[string]entity.DomainTag) (entity.DomainTags, error) {
	tags := entity.DomainTags{}
	duplicatedTag := make(map[string]struct{})

	allowListTypeTag := []string{entity.UserTagTypeStudent, entity.UserTagTypeStudentDiscount}

	for _, tag := range student.Tags {
		if _, ok := duplicatedTag[tag.PartnerInternalID().String()]; ok {
			return tags, entity.DuplicatedFieldError{
				DuplicatedField: entity.StudentTagsField,
				EntityName:      entity.StudentEntity,
				Index:           student.IndexAttr,
			}
		}
		mappedTag, ok := mapPartnerIDAndTag[tag.PartnerInternalID().String()]
		if !ok {
			return nil, entity.NotFoundError{
				FieldName:  entity.StudentTagsField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				FieldValue: tag.PartnerInternalID().String(),
			}
		}

		if !golibs.InArrayString(mappedTag.TagType().String(), allowListTypeTag) {
			return nil, entity.InvalidFieldError{
				FieldName:  entity.StudentTagsField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				Reason:     entity.NotMatchingConstants,
			}
		}

		duplicatedTag[tag.PartnerInternalID().String()] = struct{}{}
		tags = append(tags, mappedTag)
	}
	return tags, nil
}

func validateUserAddress(student aggregate.DomainStudent, mapPrefectureCodeAndPrefecture map[string]entity.DomainPrefecture) (entity.DomainUserAddress, error) {
	if student.Prefecture == nil {
		return student.UserAddress, nil
	}

	if !student.Prefecture.PrefectureCode().IsEmpty() {
		prefecture, ok := mapPrefectureCodeAndPrefecture[student.Prefecture.PrefectureCode().String()]
		if !ok {
			return nil, entity.NotFoundError{
				FieldName:  entity.StudentUserAddressPrefectureField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
			}
		}

		return entity.UserAddressWillBeDelegated{
			UserAddressAttribute: student.UserAddress,
			HasPrefectureID:      prefecture,
			HasUserID:            student.DomainStudent,
			HasOrganizationID:    student.DomainStudent,
		}, nil
	}
	return student.UserAddress, nil
}

func validatePhoneNumbers(student aggregate.DomainStudent) (entity.DomainUserPhoneNumbers, error) {
	index := student.IndexAttr
	userPhoneNumbers := make(entity.DomainUserPhoneNumbers, 0, len(student.UserPhoneNumbers))
	registeredPhoneNumber := make(map[string]struct{})

	pattern := regexp.MustCompile(constant.PhoneNumberPattern)
	fieldName := "phone_number"

	for _, phoneNumber := range student.UserPhoneNumbers {
		if !phoneNumber.PhoneNumber().IsEmpty() {
			phoneNumberStr := phoneNumber.PhoneNumber().String()
			switch phoneNumber.Type().String() {
			case entity.StudentPhoneNumber:
				fieldName = entity.StudentFieldStudentPhoneNumber
			case entity.StudentHomePhoneNumber:
				fieldName = entity.StudentFieldHomePhoneNumber
			case entity.ParentPrimaryPhoneNumber:
				fieldName = entity.StudentFieldPrimaryPhoneNumber
			case entity.ParentSecondaryPhoneNumber:
				fieldName = entity.ParentSecondaryPhoneNumber
			}

			if ok := pattern.MatchString(phoneNumberStr); !ok {
				return nil, entity.InvalidFieldError{
					FieldName:  fieldName,
					EntityName: entity.StudentEntity,
					Index:      index,
					Reason:     entity.NotMatchingPattern,
				}
			}

			if _, ok := registeredPhoneNumber[phoneNumberStr]; ok {
				return nil, entity.DuplicatedFieldError{
					DuplicatedField: fieldName,
					EntityName:      entity.StudentEntity,
					Index:           index,
				}
			}
			registeredPhoneNumber[phoneNumberStr] = struct{}{}
		}

		userPhoneNumbers = append(userPhoneNumbers, entity.UserPhoneNumberWillBeDelegated{
			UserPhoneNumberAttribute: phoneNumber,
		})
	}

	return userPhoneNumbers, nil
}

func validateSchoolHistories(student aggregate.DomainStudent, mapPartnerIDAndSchool map[string]entity.DomainSchool, mapPartnerIDAndSchoolCourse map[string]entity.DomainSchoolCourse) (entity.DomainSchoolHistories, error) {
	schoolHistories := entity.DomainSchoolHistories{}
	duplicatedSchool := make(map[string]struct{})
	duplicatedSchoolCourse := make(map[string]struct{})
	duplicatedSchoolLevel := make(map[string]struct{})
	for idx, schoolInfo := range student.SchoolInfos {
		if schoolInfo.PartnerInternalID().IsEmpty() {
			return schoolHistories, entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  entity.StudentSchoolField,
					EntityName: entity.StudentEntity,
					Index:      student.IndexAttr,
					Reason:     entity.Empty,
				},
				NestedFieldName: entity.SchoolHistories,
				NestedIndex:     idx,
			}
		}

		schoolHistoryDelegated := entity.SchoolHistoryWillBeDelegated{}
		// start date and end date are optional, so len(school_histories) maybe less than len(school_infos)
		schoolHistory := student.SchoolHistories[idx]
		if !schoolHistory.StartDate().Time().IsZero() && !schoolHistory.EndDate().Time().IsZero() {
			if schoolHistory.StartDate().Time().After(schoolHistory.EndDate().Time()) {
				return schoolHistories, entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						FieldName:  entity.StudentSchoolHistoryStartDateField,
						EntityName: entity.StudentEntity,
						Index:      student.IndexAttr,
						Reason:     entity.StartDateAfterEndDate,
					},
					NestedFieldName: entity.StudentSchoolHistoryField,
					NestedIndex:     idx,
				}
			}
		}

		schoolHistoryDelegated.SchoolHistoryAttribute = schoolHistory
		// validate duplicated school_id
		schoolPartnerID := schoolInfo.PartnerInternalID().String()
		if _, ok := duplicatedSchool[schoolPartnerID]; ok {
			return schoolHistories, entity.DuplicatedFieldError{
				DuplicatedField: entity.StudentSchoolField,
				EntityName:      entity.StudentEntity,
				Index:           student.IndexAttr,
			}
		}

		// check existed_school_course
		school, ok := mapPartnerIDAndSchool[schoolPartnerID]
		if !ok {
			return schoolHistories, entity.NotFoundError{
				FieldName:  entity.StudentSchoolField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				FieldValue: schoolPartnerID,
			}
		}
		if school.IsArchived().Boolean() {
			return schoolHistories, entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				Reason:     entity.Archived,
			}
		}

		if _, ok := duplicatedSchoolLevel[school.SchoolLevelID().String()]; ok {
			return schoolHistories, entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				Reason:     entity.AlreadyRegistered,
			}
		}
		duplicatedSchool[schoolPartnerID] = struct{}{}
		duplicatedSchoolLevel[school.SchoolLevelID().String()] = struct{}{}

		schoolCourseDelegated := entity.SchoolCourseWillBeDelegated{
			SchoolCourseAttribute: entity.DefaultDomainSchoolCourse{},
		}

		// school course is optional, so len(school_histories) maybe less than len(school_infos)
		schoolCoursePartnerID := student.SchoolCourses[idx].PartnerInternalID().String()
		if schoolCoursePartnerID == "" {
			schoolHistories = append(schoolHistories, entity.SchoolHistoryWillBeDelegated{
				SchoolHistoryAttribute: schoolHistoryDelegated,
				HasSchoolInfoID:        school,
				HasSchoolCourseID:      schoolCourseDelegated,
				HasUserID:              student.DomainStudent,
				HasOrganizationID:      school,
			})
			continue
		}
		// check duplicated school course
		if _, ok := duplicatedSchoolCourse[schoolCoursePartnerID]; ok {
			return schoolHistories, entity.DuplicatedFieldError{
				DuplicatedField: entity.StudentSchoolCourseField,
				EntityName:      entity.StudentEntity,
				Index:           student.IndexAttr,
			}
		}

		// check existed school_course
		schoolCourse, ok := mapPartnerIDAndSchoolCourse[schoolCoursePartnerID]
		if !ok {
			return schoolHistories, entity.NotFoundError{
				FieldName:  entity.StudentSchoolCourseField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
			}
		}
		if schoolCourse.IsArchived().Boolean() {
			return schoolHistories, entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolCourseField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				Reason:     entity.Archived,
			}
		}

		if !school.SchoolID().Equal(schoolCourse.SchoolID()) {
			return schoolHistories, entity.InvalidFieldError{
				FieldName:  entity.StudentSchoolCourseField,
				EntityName: entity.StudentEntity,
				Index:      student.IndexAttr,
				Reason:     entity.NotMatching,
			}
		}
		duplicatedSchoolCourse[schoolCoursePartnerID] = struct{}{}
		schoolCourseDelegated.SchoolCourseAttribute = schoolCourse
		schoolHistories = append(schoolHistories, entity.SchoolHistoryWillBeDelegated{
			SchoolHistoryAttribute: schoolHistoryDelegated,
			HasSchoolInfoID:        school,
			HasSchoolCourseID:      schoolCourseDelegated,
			HasUserID:              student.DomainStudent,
			HasOrganizationID:      school,
		})
	}

	return schoolHistories, nil
}

func validateUserNameForCreating(reqStudent aggregate.DomainStudent, mapUserNameAndUser map[string]entity.User) error {
	if len(mapUserNameAndUser) == 0 {
		return nil
	}
	if _, ok := mapUserNameAndUser[reqStudent.UserName().String()]; ok {
		// Don't allow to create user with username existing
		return entity.ExistingDataError{
			FieldName:  string(entity.UserFieldUserName),
			EntityName: entity.StudentEntity,
			Index:      reqStudent.IndexAttr,
		}
	}
	return nil
}

func validateUserNameForUpdating(reqStudent aggregate.DomainStudent, mapUserNameAndUser map[string]entity.User) error {
	if existingUser, ok := mapUserNameAndUser[reqStudent.UserName().String()]; ok {
		// Don't allow to update username when username is existing in other user
		if !reqStudent.UserID().Equal(existingUser.UserID()) {
			return entity.ExistingDataError{
				FieldName:  string(entity.UserFieldUserName),
				EntityName: entity.StudentEntity,
				Index:      reqStudent.IndexAttr,
			}
		}
	}
	return nil
}

func validateEmailForCreating(reqStudent aggregate.DomainStudent, mapEmailAndUser map[string]entity.User) error {
	if len(mapEmailAndUser) == 0 {
		return nil
	}
	if _, ok := mapEmailAndUser[reqStudent.Email().String()]; ok {
		// Don't allow to create user with email existing
		return entity.ExistingDataError{
			FieldName:  string(entity.UserFieldEmail),
			EntityName: entity.StudentEntity,
			Index:      reqStudent.IndexAttr,
		}
	}
	// Skip don't match email if email is empty. Because empty email was validated in func ValidateUserEmail
	return nil
}

func validateEmailForUpdating(reqStudent aggregate.DomainStudent, mapEmailAndUser map[string]entity.User) error {
	if existingUser, ok := mapEmailAndUser[reqStudent.Email().String()]; ok {
		// Don't allow to update email when email is existing in other user
		if !reqStudent.UserID().Equal(existingUser.UserID()) {
			return entity.ExistingDataError{
				FieldName:  string(entity.UserFieldEmail),
				EntityName: entity.StudentEntity,
				Index:      reqStudent.IndexAttr,
			}
		}
	}
	return nil
}

func validateUserIDForUpdating(reqStudent aggregate.DomainStudent, mapUserIDAndUser map[string]entity.User) error {
	if _, ok := mapUserIDAndUser[reqStudent.UserID().String()]; !ok {
		return entity.NotFoundError{
			FieldName:  string(entity.UserFieldUserID),
			EntityName: entity.StudentEntity,
			Index:      reqStudent.IndexAttr,
			FieldValue: reqStudent.UserID().String(),
		}
	}
	return nil
}

func validateExternalUserIDForUpdating(reqStudent aggregate.DomainStudent, mapUserIDAndUser map[string]entity.User, mapExternalUserIDAndUser map[string]entity.User, mapExternalUserIDAndStudentUser map[string]entity.User) error {
	if existingUser, ok := mapExternalUserIDAndUser[reqStudent.ExternalUserID().String()]; ok {
		if _, ok := mapExternalUserIDAndStudentUser[reqStudent.ExternalUserID().String()]; !ok {
			// when external user id of this student is existed in mapExternalUserIDAndUser
			// but not in mapExternalUserIDAndStudentUser
			// -> that means this external user id is belong to other type of users like staff, parent
			return entity.ExistingDataError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.StudentEntity,
				Index:      reqStudent.IndexAttr,
			}
		}

		if existingUser.UserID().Equal(reqStudent.UserID()) {
			return nil
		}

		// Don't allow to update external_user_id when external_user_id is existing in user
		return entity.ExistingDataError{
			FieldName:  string(entity.UserFieldExternalUserID),
			EntityName: entity.StudentEntity,
			Index:      reqStudent.IndexAttr,
		}
	}
	if existingUser, ok := mapUserIDAndUser[reqStudent.UserID().String()]; ok {
		// Allow to update external_user_id when existing user have the empty external_user_id
		if existingUser.ExternalUserID().IsEmpty() {
			return nil
		}
		// Don't allow to update external_user_id when req student is not equal existing user
		if !reqStudent.ExternalUserID().Equal(existingUser.ExternalUserID()) {
			return entity.UpdateFieldError{
				FieldName:  string(entity.UserFieldExternalUserID),
				EntityName: entity.StudentEntity,
				Index:      reqStudent.IndexAttr,
			}
		}
	}
	return nil
}

func validateExternalUserIDForCreating(reqStudent aggregate.DomainStudent, mapExternalUserIDAndUser map[string]entity.User) error {
	if len(mapExternalUserIDAndUser) == 0 {
		return nil
	}
	// External_user_id is not mandatory
	if reqStudent.ExternalUserID().IsEmpty() {
		return nil
	}
	if _, ok := mapExternalUserIDAndUser[reqStudent.ExternalUserID().String()]; ok {
		// Don't allow to create user with external_user_id existing
		return entity.ExistingDataError{
			FieldName:  string(entity.UserFieldExternalUserID),
			EntityName: entity.StudentEntity,
			Index:      reqStudent.IndexAttr,
		}
	}
	return nil
}

func removeDuplicatedExternalUserID(domainStudents aggregate.DomainStudents) (aggregate.DomainStudents, []error) {
	uniqueExternalUserID := make(map[string]bool)
	errorCollection := make([]error, 0)
	uniqueExternalUserIDStudents := make(aggregate.DomainStudents, 0)

	for _, student := range domainStudents {
		// External_user_id is not mandatory
		if student.ExternalUserID().IsEmpty() {
			uniqueExternalUserIDStudents = append(uniqueExternalUserIDStudents, student)
			continue
		}
		if !uniqueExternalUserID[student.ExternalUserID().String()] {
			uniqueExternalUserID[student.ExternalUserID().String()] = true
			uniqueExternalUserIDStudents = append(uniqueExternalUserIDStudents, student)
		} else {
			errorCollection = append(errorCollection, entity.DuplicatedFieldError{
				DuplicatedField: string(entity.UserFieldExternalUserID),
				Index:           student.IndexAttr,
				EntityName:      entity.StudentEntity,
			})
		}
	}
	return uniqueExternalUserIDStudents, errorCollection
}

func removeDuplicatedUserName(domainStudents aggregate.DomainStudents) (aggregate.DomainStudents, []error) {
	uniqueUserNames := make(map[string]bool)
	errorCollection := make([]error, 0)
	uniqueUserNameStudents := make(aggregate.DomainStudents, 0)

	for _, student := range domainStudents {
		// check empty username in the func entity.ValidateUserUserName
		if student.UserName().IsEmpty() {
			uniqueUserNameStudents = append(uniqueUserNameStudents, student)
			continue
		}
		if !uniqueUserNames[student.UserName().String()] {
			uniqueUserNames[student.UserName().String()] = true
			uniqueUserNameStudents = append(uniqueUserNameStudents, student)
		} else {
			errorCollection = append(errorCollection, entity.DuplicatedFieldError{
				DuplicatedField: string(entity.UserFieldUserName),
				Index:           student.IndexAttr,
				EntityName:      entity.StudentEntity,
			})
		}
	}
	return uniqueUserNameStudents, errorCollection
}

func removeDuplicatedEmail(domainStudents aggregate.DomainStudents) (aggregate.DomainStudents, []error) {
	uniqueEmails := make(map[string]bool)
	errorCollection := make([]error, 0)
	uniqueEmailStudents := make(aggregate.DomainStudents, 0)

	for _, student := range domainStudents {
		// check empty email in the func entity.ValidateUserEmail
		if student.Email().IsEmpty() {
			uniqueEmailStudents = append(uniqueEmailStudents, student)
			continue
		}

		if !uniqueEmails[student.Email().String()] {
			uniqueEmails[student.Email().String()] = true
			uniqueEmailStudents = append(uniqueEmailStudents, student)
		} else {
			errorCollection = append(errorCollection, entity.DuplicatedFieldError{
				DuplicatedField: string(entity.UserFieldEmail),
				Index:           student.IndexAttr,
				EntityName:      entity.StudentEntity,
			})
		}
	}
	return uniqueEmailStudents, errorCollection
}

func removeDuplicatedUserID(domainStudents aggregate.DomainStudents) (aggregate.DomainStudents, []error) {
	uniqueUserID := make(map[string]bool)
	errorCollection := make([]error, 0)
	uniqueUserIDStudents := make(aggregate.DomainStudents, 0)

	for _, student := range domainStudents {
		if student.UserID().IsEmpty() {
			uniqueUserIDStudents = append(uniqueUserIDStudents, student)
			continue
		}

		if !uniqueUserID[student.UserID().String()] {
			uniqueUserID[student.UserID().String()] = true
			uniqueUserIDStudents = append(uniqueUserIDStudents, student)
		} else {
			errorCollection = append(errorCollection, entity.DuplicatedFieldError{
				DuplicatedField: string(entity.UserFieldUserID),
				Index:           student.IndexAttr,
				EntityName:      entity.StudentEntity,
			})
		}
	}
	return uniqueUserIDStudents, errorCollection
}

func validateEnrollmentStatusHistoriesForCreating(reqStudent aggregate.DomainStudent) error {
	enrollmentStatusHistories := reqStudent.EnrollmentStatusHistories
	index := reqStudent.IndexAttr

	if len(enrollmentStatusHistories) == 0 {
		return entity.MissingMandatoryFieldError{
			FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
			EntityName: entity.StudentEntity,
			Index:      index,
		}
	}

	thereIsAtLeastOneActivatedEnrollmentStatus := containsActivatedEnrollmentStatus(enrollmentStatusHistories.EnrollmentStatuses())

	for idx, enrollmentStatusHistory := range enrollmentStatusHistories {
		if enrollmentStatusHistory.EnrollmentStatus().IsEmpty() {
			return entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      index,
					Reason:     entity.Empty,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     idx,
			}
		}

		if enrollmentStatusHistory.LocationID().IsEmpty() {
			return entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  entity.StudentLocationsField,
					EntityName: entity.StudentEntity,
					Index:      index,
					Reason:     entity.Empty,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     idx,
			}
		}

		// Check req create student have to have at least one status is activated
		if enrollmentStatusHistory.EnrollmentStatus().String() == entity.StudentEnrollmentStatusTemporary {
			if !thereIsAtLeastOneActivatedEnrollmentStatus {
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
						EntityName: entity.StudentEntity,
						Index:      index,
						Reason:     entity.MissingActivatedEnrollmentStatus,
					},
					NestedFieldName: entity.EnrollmentStatusHistories,
					NestedIndex:     idx,
				}
			}
		}
	}
	return nil
}

func containsActivatedEnrollmentStatus(enrollmentStatuses field.Strings) bool {
	for _, enrollmentStatus := range enrollmentStatuses {
		if enrollmentStatus.String() != entity.StudentEnrollmentStatusTemporary {
			return true
		}
	}
	return false
}

func validateEnrollmentStatusHistories(reqStudent aggregate.DomainStudent, isInOrderFlow bool) error {
	enrollmentStatusHistories := reqStudent.EnrollmentStatusHistories
	index := reqStudent.IndexAttr
	for idx, reqEnrollmentStatusHistory := range enrollmentStatusHistories {
		// Only allow to create student with potential/temporary/non-potential status in order flow
		if isInOrderFlow {
			if !golibs.InArrayString(reqEnrollmentStatusHistory.EnrollmentStatus().String(), allowListInOrderFlow) {
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
						EntityName: entity.StudentEntity,
						Index:      index,
						Reason:     entity.NotInAllowListEnrollmentStatus,
					},
					NestedFieldName: entity.EnrollmentStatusHistories,
					NestedIndex:     idx,
				}
			}
		}

		if reqEnrollmentStatusHistory.EnrollmentStatus().IsEmpty() {
			if !reqEnrollmentStatusHistory.LocationID().IsEmpty() {
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
						EntityName: entity.StudentEntity,
						Index:      index,
						Reason:     entity.Empty,
					},
					NestedFieldName: entity.EnrollmentStatusHistories,
					NestedIndex:     idx,
				}
			}
		} else {
			if reqEnrollmentStatusHistory.LocationID().IsEmpty() {
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						FieldName:  entity.StudentLocationsField,
						EntityName: entity.StudentEntity,
						Index:      index,
						Reason:     entity.Empty,
					},
					NestedFieldName: entity.EnrollmentStatusHistories,
					NestedIndex:     idx,
				}
			}
		}

		// Potential/Temporary/Non-Potential status start date can not be after current date
		if !golibs.InArrayString(reqEnrollmentStatusHistory.EnrollmentStatus().String(), allowListInOrderFlow) {
			continue
		}
		roundedStartDateTime := utils.TruncateTimeToStartOfDay(reqEnrollmentStatusHistory.StartDate().Time())
		roundedCurrentDateTime := utils.TruncateTimeToStartOfDay(time.Now())

		if roundedStartDateTime.After(roundedCurrentDateTime) {
			return entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
					EntityName: entity.StudentEntity,
					Index:      index,
					Reason:     entity.StartDateAfterCurrentDate,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     idx,
			}
		}
	}
	return nil
}

func validateEnrollmentStatusHistoriesForUpdating(reqStudent aggregate.DomainStudent, existingEnrollmentStatusHistories entity.DomainEnrollmentStatusHistories, isOrderFlow bool) error {
	// Allow to update student without enrollment status history
	if len(reqStudent.EnrollmentStatusHistories) == 0 && len(reqStudent.Locations) == 0 {
		return nil
	}

	activatedEnrollmentStatusInBD := existingEnrollmentStatusHistories.GetActivatedByRequest(reqStudent.UserID(), reqStudent.EnrollmentStatusHistories)

	for idx, reqEnrollmentStatusHistory := range reqStudent.EnrollmentStatusHistories {
		existingEnrollmentStatusHistory := existingEnrollmentStatusHistories.GetExactly(reqEnrollmentStatusHistory)
		// Skip validation if enrollment status history don't change anything
		if existingEnrollmentStatusHistory != nil {
			continue
		}

		allEnrollmentStatusHistories := existingEnrollmentStatusHistories.GetAllByUserIDLocationID(reqEnrollmentStatusHistory.UserID(), reqEnrollmentStatusHistory.LocationID())
		/* If there is no enrollment status history at the locationID, skip validation.
		Because don't have existing enrollment status history in database.
		That means this is a new enrollment status history at new location.*/
		if len(allEnrollmentStatusHistories) == 0 {
			continue
		}

		activatedEnrollmentStatusHistory := existingEnrollmentStatusHistories.GetActivatedByUserIDLocationID(reqEnrollmentStatusHistory.UserID(), reqEnrollmentStatusHistory.LocationID())

		if activatedEnrollmentStatusHistory == nil {
			// if there is no activated enrollment status history at the locationID, get the latest enrollment status history
			activatedEnrollmentStatusHistory = existingEnrollmentStatusHistories.GetLatestByUserIDLocationID(reqEnrollmentStatusHistory.UserID(), reqEnrollmentStatusHistory.LocationID())
		}

		if reqEnrollmentStatusHistory.EnrollmentStatus().String() == entity.StudentEnrollmentStatusTemporary {
			// Skip validation if req enrollment status history and existing enrollment status history are temporary
			if activatedEnrollmentStatusHistory.EnrollmentStatus().String() == entity.StudentEnrollmentStatusTemporary {
				continue
			}
			thereIsAtLeastOneActivatedEnrollmentStatus := containsActivatedEnrollmentStatus(append(activatedEnrollmentStatusInBD.EnrollmentStatuses(), reqEnrollmentStatusHistory.EnrollmentStatus()))
			// Req update student and existingEnrollmentStatusHistory have to have at least one status is activated
			if !thereIsAtLeastOneActivatedEnrollmentStatus {
				return entity.InvalidFieldErrorWithArrayNestedField{
					InvalidFieldError: entity.InvalidFieldError{
						FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
						EntityName: entity.StudentEntity,
						Index:      reqStudent.IndexAttr,
						Reason:     entity.MissingActivatedEnrollmentStatus,
					},
					NestedFieldName: entity.EnrollmentStatusHistories,
					NestedIndex:     idx,
				}
			}
		}
		// non erp status can not change to others at order flow
		if isOrderFlow && !golibs.InArrayString(activatedEnrollmentStatusHistory.EnrollmentStatus().String(), ERPEnrollmentStatus) {
			return entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      reqStudent.IndexAttr,
					Reason:     entity.ChangingNonERPStatusToOtherStatusAtOrderFlow,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     idx,
			}
		}
		if err := validateEntityEnrollmentStatusHistoryForUpdating(activatedEnrollmentStatusHistory, reqEnrollmentStatusHistory, reqStudent.IndexAttr, idx); err != nil {
			return err
		}
	}
	return nil
}

func validateEntityEnrollmentStatusHistoryForUpdating(activatedEnrollmentStatus entity.DomainEnrollmentStatusHistory, reqEnrollmentStatus entity.DomainEnrollmentStatusHistory, studentIndex int, nestedIndex int) error {
	activatedStatus := activatedEnrollmentStatus.EnrollmentStatus().String()
	reqStatus := reqEnrollmentStatus.EnrollmentStatus().String()
	activatedStartDateTruncated := utils.TruncateTimeToStartOfDay(activatedEnrollmentStatus.StartDate().Time())
	reqStartDateTruncated := utils.TruncateTimeToStartOfDay(reqEnrollmentStatus.StartDate().Time())
	timeNowTruncated := utils.TruncateTimeToStartOfDay(time.Now())
	switch activatedStatus {
	// status potential and non-potential can change to any status in reqEnrollmentStatus
	case entity.StudentEnrollmentStatusPotential:
		break
		// status non-potential can't change to any status
	case entity.StudentEnrollmentStatusNonPotential:
		if reqStatus != entity.StudentEnrollmentStatusNonPotential {
			return entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryEnrollmentStatus),
					EntityName: entity.StudentEntity,
					Index:      studentIndex,
					Reason:     entity.ChangingNonPotentialToOtherStatus,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     nestedIndex,
			}
		}
	}

	// Skip if nothing change
	if activatedStatus == reqStatus {
		// Nothing change, skip validation
		if activatedStartDateTruncated.Equal(reqStartDateTruncated) {
			return nil
		}
		// Don't allow status don't change, if start_date don't changed
		if !activatedStartDateTruncated.Equal(reqStartDateTruncated) {
			return entity.InvalidFieldErrorWithArrayNestedField{
				InvalidFieldError: entity.InvalidFieldError{
					FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
					EntityName: entity.StudentEntity,
					Index:      studentIndex,
					Reason:     entity.ChangingStartDateWithoutChangingStatus,
				},
				NestedFieldName: entity.EnrollmentStatusHistories,
				NestedIndex:     nestedIndex,
			}
		}
		return nil
	}
	// Allow to change status in a day
	if reqStartDateTruncated.Equal(timeNowTruncated) && activatedStartDateTruncated.Equal(timeNowTruncated) {
		return nil
	}

	if activatedStartDateTruncated.After(reqStartDateTruncated) {
		return entity.InvalidFieldErrorWithArrayNestedField{
			InvalidFieldError: entity.InvalidFieldError{
				FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
				EntityName: entity.StudentEntity,
				Index:      studentIndex,
				Reason:     entity.ActivatedStartDateAfterReqStartDate,
			},
			NestedFieldName: entity.EnrollmentStatusHistories,
			NestedIndex:     nestedIndex,
		}
	}

	// Enrollment status is changing
	if activatedStartDateTruncated.Equal(reqStartDateTruncated) {
		return entity.InvalidFieldErrorWithArrayNestedField{
			InvalidFieldError: entity.InvalidFieldError{
				FieldName:  string(entity.FieldEnrollmentStatusHistoryStartDate),
				EntityName: entity.StudentEntity,
				Index:      studentIndex,
				Reason:     entity.ChangingStatusWithoutChangingStartDate,
			},
			NestedFieldName: entity.EnrollmentStatusHistories,
			NestedIndex:     nestedIndex,
		}
	}
	return nil
}
