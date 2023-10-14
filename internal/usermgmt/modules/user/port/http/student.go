package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DomainStudentService struct {
	DB            database.Ext
	DomainStudent interface {
		GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error)
		GetGradesByExternalIDs(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error)
		GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error)
		GetLocationsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainLocations, error)
		GetSchoolsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchools, error)
		GetSchoolCoursesByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error)
		GetPrefecturesByCodes(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error)
		GetEmailWithStudentID(ctx context.Context, studentIDs []string) (map[string]entity.User, error)

		UpsertMultiple(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToCreate ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error)
		UpsertMultipleWithErrorCollection(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error)

		IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool
		IsFeatureIgnoreInvalidRecordsOpenAPIEnabled(organization valueobj.HasOrganizationID) bool
		IsFeatureAutoDeactivateAndReactivateStudentsV2Enabled(organization valueobj.HasOrganizationID) bool
		IsDisableAutoDeactivateStudents(organization valueobj.HasOrganizationID) bool
		IsExperimentalBulkInsertEnrollmentStatusHistories(organization valueobj.HasOrganizationID) bool
		IsAuthUsernameConfigEnabled(ctx context.Context) (bool, error)
	}
	FeatureManager interface {
		FeatureUsernameToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
		FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(ctx context.Context, org valueobj.HasOrganizationID, option unleash.DomainStudentFeatureOption) unleash.DomainStudentFeatureOption
	}
}

func (port *DomainStudentService) UpsertStudents(c *gin.Context) {
	zapLogger := ctxzap.Extract(c.Request.Context())
	organization, err := interceptors.OrganizationFromContext(c.Request.Context())
	if err != nil {
		zapLogger.Error(err.Error(), zap.Error(err))
		ResponseListErrors(c, []error{err})
		return
	}
	isEnableIgnoreError := port.DomainStudent.IsFeatureIgnoreInvalidRecordsOpenAPIEnabled(organization)
	var req UpsertStudentsRequest
	if err := ParseJSONPayload(c.Request, &req); err != nil {
		if isEnableIgnoreError {
			zapLogger.Error(err.Error(), zap.Error(err))
			ResponseListErrors(c, []error{err})
			return
		}

		ResponseError(c, err)
		return
	}

	option := unleash.DomainStudentFeatureOption{
		DomainUserFeatureOption: unleash.DomainUserFeatureOption{
			EnableIgnoreUpdateEmail: true,
		},
	}
	option = port.FeatureManager.FeatureUsernameToStudentFeatureOption(c.Request.Context(), organization, option)
	option = port.FeatureManager.FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(c.Request.Context(), organization, option)
	option = port.FeatureManager.FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(c.Request.Context(), organization, option)
	option = port.FeatureManager.FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(c.Request.Context(), organization, option)

	domainStudentsResp := aggregate.DomainStudents{}
	if isEnableIgnoreError {
		resp, listErrors := port.upsertStudentWithErrorCollections(c.Request.Context(), req.Students, option)
		if len(listErrors) > 0 {
			ResponseListErrors(c, listErrors)
			return
		}
		copy(domainStudentsResp, resp)
		domainStudentsResp = resp
	} else {
		studentsToUpsert, err := port.ToDomainStudentsAgg(c.Request.Context(), req.Students, option.EnableUsername)
		if err != nil {
			ResponseError(c, err)
			return
		}

		if !option.EnableUsername {
			studentsToUpsert, err = port.fillExistedEmailOfUsers(c.Request.Context(), studentsToUpsert)
			if err != nil {
				ResponseError(c, err)
				return
			}
		}

		domainStudentsResp, err = port.DomainStudent.UpsertMultiple(c.Request.Context(), option, studentsToUpsert...)
		if err != nil {
			ResponseError(c, err)
			return
		}
	}

	data := make([]map[string]interface{}, 0, len(domainStudentsResp))
	for _, student := range domainStudentsResp {
		data = append(data, map[string]interface{}{
			"user_id":          student.UserID().String(),
			"external_user_id": student.ExternalUserID().String(),
		})
	}

	// TODO: update response later
	c.JSON(http.StatusOK, Response{
		Data:    data,
		Code:    20000,
		Message: "success",
	})
}

func (port *DomainStudentService) ToDomainStudentsAgg(ctx context.Context, studentProfiles []StudentProfile, isEnableUsername bool) ([]aggregate.DomainStudent, error) {
	userIDs, err := port.toUserIDs(ctx, studentProfiles)
	if err != nil {
		return nil, err
	}

	gradeIDs, err := port.toGradeIDs(ctx, studentProfiles)
	if err != nil {
		return nil, err
	}

	studentsToUpsert := []aggregate.DomainStudent{}

	for idx, student := range studentProfiles {
		if userIDs[idx] != "" {
			student.UserID = field.NewString(userIDs[idx])
		}
		// trim external_user_id
		student.ExternalUserID = field.NewString(strings.TrimSpace(student.ExternalUserID.String()))

		student.FullName = field.NewString(utils.CombineFirstNameAndLastNameToFullName(student.FirstName.String(), student.LastName.String()))
		if student.FirstNamePhonetic.String() == "" {
			student.FirstNamePhonetic = field.NewNullString()
		}
		if student.LastNamePhonetic.String() == "" {
			student.LastNamePhonetic = field.NewNullString()
		}
		student.FullNamePhonetic = field.NewString(utils.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(student.FirstNamePhonetic.String(), student.LastNamePhonetic.String()))
		if student.FullNamePhonetic.String() == "" {
			student.FullNamePhonetic = field.NewNullString()
		}
		if !isEnableUsername {
			student.UserName = student.Email
		}
		student.LoginEmail = student.Email

		if field.IsPresent(student.Grade) {
			student.GradeID = field.NewString(gradeIDs[idx])
		}

		taggedUsers, err := port.toDomainTaggedUsers(ctx, student.Tags)
		if err != nil {
			return nil, err
		}

		userAccessPaths, err := port.toDomainUserAccessPaths(ctx, student.Locations)
		if err != nil {
			return nil, err
		}

		schoolHistories, err := port.toDomainSchoolHistories(ctx, student.SchoolHistories)
		if err != nil {
			err, ok := err.(errcode.Error)
			if ok {
				err.FieldName = fmt.Sprintf("students[%d].%s", idx, err.FieldName)
			}
			return nil, err
		}

		domainStudentAgg := aggregate.DomainStudent{
			DomainStudent:   DomainStudentImpl{StudentProfile: student},
			UserAccessPaths: userAccessPaths,
			TaggedUsers:     taggedUsers,
			SchoolHistories: schoolHistories,
			IndexAttr:       idx,
		}
		if len(student.EnrollmentStatusHistories) != 0 {
			domainStudentAgg.EnrollmentStatusHistories, domainStudentAgg.UserAccessPaths, err = port.toDomainEnrollmentStatusHistories(ctx, student.EnrollmentStatusHistories, userAccessPaths)
			if err != nil {
				return nil, err
			}
		}

		if student.Address != nil {
			domainStudentAgg.UserAddress, err = port.toUserAddress(ctx, *student.Address)
			if err != nil {
				return nil, err
			}
		}

		if student.UserPhoneNumber != nil {
			userPhoneNumbers := entity.DomainUserPhoneNumbers{}
			if field.IsPresent(student.UserPhoneNumber.PhoneNumber) {
				userPhoneNumbers = append(userPhoneNumbers, toDomainUserPhoneNumbers(student.UserPhoneNumber.PhoneNumber.String(), entity.UserPhoneNumberTypeStudentPhoneNumber))
			}
			if field.IsPresent(student.UserPhoneNumber.HomePhoneNumber) {
				userPhoneNumbers = append(userPhoneNumbers, toDomainUserPhoneNumbers(student.UserPhoneNumber.HomePhoneNumber.String(), entity.UserPhoneNumberTypeStudentHomePhoneNumber))
			}
			domainStudentAgg.UserPhoneNumbers = userPhoneNumbers
		}

		studentsToUpsert = append(studentsToUpsert, domainStudentAgg)
	}

	return studentsToUpsert, nil
}

func (port *DomainStudentService) toUserIDs(ctx context.Context, studentProfiles []StudentProfile) ([]string, error) {
	externalUserIDs := []string{}
	for i, s := range studentProfiles {
		trimmedExternalUserID := strings.TrimSpace(s.ExternalUserID.String())
		if trimmedExternalUserID == "" {
			return nil, errcode.Error{
				FieldName: fmt.Sprintf("students[%d].external_user_id", i),
				Code:      errcode.MissingMandatory,
			}
		}
		externalUserIDs = append(externalUserIDs, trimmedExternalUserID)
	}

	existingUsers, err := port.DomainStudent.GetUsersByExternalIDs(ctx, externalUserIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.DomainStudent.GetUsersByExternalIDs"),
		}
	}

	userIDs := []string{}
	for _, externalUserID := range externalUserIDs {
		userID := ""
		for _, user := range existingUsers {
			if externalUserID == user.ExternalUserID().String() {
				userID = user.UserID().String()
			}
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (port *DomainStudentService) toGradeIDs(ctx context.Context, studentProfiles []StudentProfile) ([]string, error) {
	partnerIDs := []string{}

	for _, s := range studentProfiles {
		partnerIDs = append(partnerIDs, s.Grade.String())
	}

	grades, err := port.DomainStudent.GetGradesByExternalIDs(ctx, partnerIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "s.DomainStudent.GetUsersByExternalIDs"),
		}
	}

	gradeIDs := []string{}
	for _, partnerID := range partnerIDs {
		gradeID := ""
		for _, grade := range grades {
			if partnerID == grade.PartnerInternalID().String() {
				gradeID = grade.GradeID().String()
			}
		}
		gradeIDs = append(gradeIDs, gradeID)
	}

	return gradeIDs, nil
}

func toDomainUserPhoneNumbers(phoneNumber string, phoneNumberType string) entity.DomainUserPhoneNumber {
	return entity.UserPhoneNumberWillBeDelegated{
		UserPhoneNumberAttribute: DomainUserPhoneNumberImpl{
			userPhoneNumberID: field.NewString(idutil.ULIDNow()),
			phoneNumberType:   field.NewString(phoneNumberType),
			phoneNumber:       field.NewString(phoneNumber),
		},
	}
}

func (port *DomainStudentService) toDomainTaggedUsers(ctx context.Context, partnerInternalIDs []field.String) (entity.DomainTaggedUsers, error) {
	domainTaggedUsers := entity.DomainTaggedUsers{}
	externalIDs := []string{}
	for _, partnerInternalID := range partnerInternalIDs {
		externalIDs = append(externalIDs, partnerInternalID.String())
	}

	tags, err := port.DomainStudent.GetTagsByExternalIDs(ctx, externalIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.DomainStudent.GetTagsByExternalIDs"),
		}
	}

	for _, partnerInternalID := range partnerInternalIDs {
		taggedUser := entity.TaggedUserWillBeDelegated{
			HasTagID: entity.EmptyDomainTag{},
		}
		for _, tag := range tags {
			if partnerInternalID.String() == tag.PartnerInternalID().String() {
				taggedUser.HasTagID = tag
				break
			}
		}
		domainTaggedUsers = append(domainTaggedUsers, taggedUser)
	}

	return domainTaggedUsers, nil
}

func (port *DomainStudentService) toDomainSchoolHistories(ctx context.Context, schoolHistories []SchoolHistoryPayload) (entity.DomainSchoolHistories, error) {
	domainSchoolHistories := entity.DomainSchoolHistories{}
	schoolPartnerIDs := []string{}
	schoolCoursePartnerIDs := []string{}
	for _, schoolHistory := range schoolHistories {
		schoolPartnerIDs = append(schoolPartnerIDs, schoolHistory.School.String())
		if schoolHistory.SchoolCourse.String() != "" {
			schoolCoursePartnerIDs = append(schoolCoursePartnerIDs, schoolHistory.SchoolCourse.String())
		}
	}

	schools, err := port.DomainStudent.GetSchoolsByExternalIDs(ctx, schoolPartnerIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.DomainStudent.GetSchoolsByExternalIDs"),
		}
	}

	schoolCourses, err := port.DomainStudent.GetSchoolCoursesByExternalIDs(ctx, schoolCoursePartnerIDs)
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.DomainStudent.GetSchoolCoursesByExternalIDs"),
		}
	}

	for _, schoolCoursePartnerID := range schoolCoursePartnerIDs {
		if !golibs.InArrayString(schoolCoursePartnerID, schoolCourses.PartnerInternalIDs()) {
			return nil, errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: "school_histories.school_course",
			}
		}
	}
	for _, schoolHistory := range schoolHistories {
		domainSchoolHistory := entity.SchoolHistoryWillBeDelegated{
			SchoolHistoryAttribute: DomainSchoolHistoryImpl{SchoolHistoryPayload: schoolHistory},
			HasSchoolInfoID:        entity.DefaultDomainSchoolHistory{},
			HasSchoolCourseID:      entity.DefaultDomainSchoolHistory{},
		}
		for _, school := range schools {
			if schoolHistory.School.String() == school.PartnerInternalID().String() {
				domainSchoolHistory.HasSchoolInfoID = school
				break
			}
		}

		for _, schoolCourse := range schoolCourses {
			if schoolHistory.SchoolCourse.String() == schoolCourse.PartnerInternalID().String() {
				domainSchoolHistory.HasSchoolCourseID = schoolCourse
				break
			}
		}
		domainSchoolHistories = append(domainSchoolHistories, domainSchoolHistory)
	}

	return domainSchoolHistories, nil
}

func (port *DomainStudentService) toDomainUserAccessPaths(ctx context.Context, partnerInternalIDs []field.String) (entity.DomainUserAccessPaths, error) {
	domainUserAccessPaths := entity.DomainUserAccessPaths{}
	externalIDs := []string{}
	for _, partnerInternalID := range partnerInternalIDs {
		externalIDs = append(externalIDs, partnerInternalID.String())
	}

	locations, err := port.DomainStudent.GetLocationsByExternalIDs(ctx, externalIDs)

	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.DomainStudent.GetTagsByExternalIDs"),
		}
	}

	for _, partnerInternalID := range partnerInternalIDs {
		taggedUser := entity.UserAccessPathWillBeDelegated{
			HasLocationID: entity.DefaultUserAccessPath{},
		}
		for _, location := range locations {
			if partnerInternalID.String() == location.PartnerInternalID().String() {
				taggedUser.HasLocationID = location
			}
		}
		domainUserAccessPaths = append(domainUserAccessPaths, taggedUser)
	}

	return domainUserAccessPaths, nil
}

func (port *DomainStudentService) toDomainEnrollmentStatusHistories(ctx context.Context, enrollmentStatusHistories []EnrollmentStatusHistoryPayload, userAccessPaths entity.DomainUserAccessPaths) (entity.DomainEnrollmentStatusHistories, entity.DomainUserAccessPaths, error) {
	domainEnrollmentStatusHistories := make(entity.DomainEnrollmentStatusHistories, 0, len(enrollmentStatusHistories))
	locationPartnerIDs := make([]string, 0, len(enrollmentStatusHistories))

	for _, enrollmentStatusHistory := range enrollmentStatusHistories {
		locationPartnerIDs = append(locationPartnerIDs, enrollmentStatusHistory.Location.String())
	}
	locations, err := port.DomainStudent.GetLocationsByExternalIDs(ctx, locationPartnerIDs)

	if err != nil {
		return nil, nil, errors.Wrap(err, "service.DomainStudent.GetTagsByExternalIDs")
	}

	for _, enrollmentStatusHistory := range enrollmentStatusHistories {
		domainEnrollmentStatusHistory := entity.EnrollmentStatusHistoryWillBeDelegated{
			EnrollmentStatusHistory: DomainEnrollmentStatusHistoryImpl{
				EnrollmentStatusHistoryPayload: enrollmentStatusHistory,
			},
			HasLocationID: entity.DefaultDomainEnrollmentStatusHistory{},
		}

		for _, location := range locations {
			if location.PartnerInternalID().String() == enrollmentStatusHistory.Location.String() {
				domainEnrollmentStatusHistory.HasLocationID = location
			}
		}
		domainEnrollmentStatusHistories = append(domainEnrollmentStatusHistories, domainEnrollmentStatusHistory)
	}

	// Re-set if userAccessPaths is empty
	if len(userAccessPaths) == 0 {
		userAccessPathsPayload := make(entity.DomainUserAccessPaths, 0, len(locations))
		for _, location := range locations {
			userAccessPathsPayload = append(userAccessPathsPayload, entity.UserAccessPathWillBeDelegated{
				HasLocationID: location,
			})
		}
		return domainEnrollmentStatusHistories, userAccessPathsPayload, nil
	}
	return domainEnrollmentStatusHistories, userAccessPaths, nil
}

func (port *DomainStudentService) toUserAddress(ctx context.Context, address AddressPayload) (entity.DomainUserAddress, error) {
	domainUserAddress := entity.UserAddressWillBeDelegated{
		UserAddressAttribute: DomainUserAddressImpl{
			AddressPayload: address,
		},
		HasPrefectureID: entity.DefaultDomainPrefecture{},
	}
	if field.IsPresent(address.Prefecture) {
		domainUserAddress.HasPrefectureID = DomainUserAddressImpl{
			prefectureID: field.NewString(""),
		}
	}
	prefectures, err := port.DomainStudent.GetPrefecturesByCodes(ctx, []string{address.Prefecture.String()})
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  errors.Wrap(err, "service.DomainStudent.GetTagsByExternalIDs"),
		}
	}

	for _, prefecture := range prefectures {
		domainUserAddress.HasPrefectureID = prefecture
	}

	return domainUserAddress, nil
}

func (port *DomainStudentService) fillExistedEmailOfUsers(ctx context.Context, students aggregate.DomainStudents) ([]aggregate.DomainStudent, error) {
	userWithUserIDs, err := port.DomainStudent.GetEmailWithStudentID(ctx, students.StudentIDs())
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  err,
		}
	}

	for idx, student := range students {
		if student.UserID().String() == "" {
			continue
		}

		studentImpl, ok := student.DomainStudent.(DomainStudentImpl)
		if !ok {
			return nil, errcode.Error{
				Index: idx,
				Code:  errcode.InternalError,
				Err:   errors.New("invalid student type"),
			}
		}
		if user, ok := userWithUserIDs[studentImpl.UserID().String()]; ok {
			studentImpl.StudentProfile.Email = user.Email()
			studentImpl.StudentProfile.LoginEmail = user.LoginEmail()
			studentImpl.StudentProfile.UserName = user.Email()
			students[idx].DomainStudent = studentImpl
		}
	}

	return students, nil
}

func (port *DomainStudentService) upsertStudentWithErrorCollections(ctx context.Context, students []StudentProfile, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error) {
	zapLogger := ctxzap.Extract(ctx)

	externalUserIDs := make([]string, 0, len(students))
	for _, v := range students {
		externalUserIDs = append(externalUserIDs, v.ExternalUserID.TrimSpace().String())
	}
	mapExternalUserIDAndUser, err := port.mapExternalUserIDAndUser(ctx, externalUserIDs)
	if err != nil {
		e := InternalError{
			RawErr: err,
		}
		zapLogger.Error(e.Error(),
			zap.Error(e),
		)
		return nil, []error{e}
	}

	errorsCollection := make([]error, 0)

	domainStudents := make(aggregate.DomainStudents, 0)
	for idx, student := range students {
		student.UserID = field.NewNullString()
		user, ok := mapExternalUserIDAndUser[student.ExternalUserID.TrimSpace().String()]

		if ok {
			student.UserID = user.UserID()
		}
		if student.ExternalUserID.TrimSpace().IsEmpty() {
			e := entity.MissingMandatoryFieldError{
				EntityName: entity.StudentEntity,
				Index:      idx,
				FieldName:  string(entity.UserFieldExternalUserID),
			}
			errorsCollection = append(errorsCollection, e)
			continue
		}
		domainStudent := ToDomainStudentV2(DomainStudentImpl{StudentProfile: student}, idx, option.EnableUsername)
		domainStudents = append(domainStudents, domainStudent)
	}

	if !option.EnableUsername {
		domainStudents, err = port.fillExistedEmailOfUsers(ctx, domainStudents)
		if err != nil {
			e := InternalError{
				RawErr: errors.WithStack(errors.Wrap(err, "fillExistedEmailOfUsers")),
			}
			zapLogger.Error(e.Error(),
				zap.Error(e),
			)
			return nil, []error{e}
		}
	}

	studentResp, listOfErrors := port.DomainStudent.UpsertMultipleWithErrorCollection(ctx, domainStudents, option)
	errorsCollection = append(errorsCollection, listOfErrors...)
	return studentResp, errorsCollection
}

func ToDomainStudentV2(student DomainStudentImpl, index int, isEnableUsername bool) aggregate.DomainStudent {
	// trim external_user_id
	student.StudentProfile.ExternalUserID = student.ExternalUserID().TrimSpace()
	student.StudentProfile.FullName = field.NewString(utils.CombineFirstNameAndLastNameToFullName(student.FirstName().String(), student.LastName().String()))
	student.StudentProfile.FullNamePhonetic = field.NewString(utils.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(student.FirstNamePhonetic().String(), student.LastNamePhonetic().String()))

	// Save NULL value if empty string or space
	if student.LastNamePhonetic().TrimSpace().IsEmpty() {
		student.StudentProfile.LastNamePhonetic = field.NewNullString()
	}
	if student.FirstNamePhonetic().TrimSpace().IsEmpty() {
		student.StudentProfile.FirstNamePhonetic = field.NewNullString()
	}
	if student.FullNamePhonetic().TrimSpace().IsEmpty() {
		student.StudentProfile.FullNamePhonetic = field.NewNullString()
	}
	student.StudentProfile.LoginEmail = student.Email()
	if !isEnableUsername {
		student.StudentProfile.UserName = student.Email()
	}

	domainEnrollmentStatusHistories, domainLocations := toEnrollmentStatusHistoriesV2(student.StudentProfile.EnrollmentStatusHistories)
	domainTaggedUsers, domainTags := toTaggedUsersV2(student.Tags, student)
	domainSchoolHistories, domainSchoolsInfo, domainSchoolCourses := toSchoolHistoryV2(student.SchoolHistories)
	domainAddressAttr, domainPrefecture := toUserAddressV2(student)

	return aggregate.DomainStudent{
		DomainStudent:             student,
		Grade:                     toGradeV2(student.Grade),
		UserPhoneNumbers:          toUserPhoneNumberV2(student.UserPhoneNumber),
		UserAccessPaths:           toUserAccessPathsV2(student.Locations, student),
		UserAddress:               domainAddressAttr,
		Tags:                      domainTags,
		TaggedUsers:               domainTaggedUsers,
		SchoolHistories:           domainSchoolHistories,
		SchoolInfos:               domainSchoolsInfo,
		SchoolCourses:             domainSchoolCourses,
		EnrollmentStatusHistories: domainEnrollmentStatusHistories,
		Locations:                 domainLocations,
		Prefecture:                domainPrefecture,
		IndexAttr:                 index,
	}
}

func toGradeV2(obj field.String) entity.DomainGrade {
	if !obj.IsEmpty() {
		return entity.GradeWillBeDelegated{
			HasPartnerInternalID: GradeImpl{
				GradeAttr: obj,
			},
		}
	}
	return entity.NullDomainGrade{}
}

func toUserAccessPathsV2(locations []field.String, student entity.DomainStudent) entity.DomainUserAccessPaths {
	domainUserAccessPaths := make(entity.DomainUserAccessPaths, 0, len(locations))

	for range locations {
		domainUserAccessPaths = append(domainUserAccessPaths, entity.UserAccessPathWillBeDelegated{
			HasUserID: student,
		})
	}
	return domainUserAccessPaths
}

func toUserAddressV2(student DomainStudentImpl) (entity.DomainUserAddress, entity.DomainPrefecture) {
	if student.Address == nil {
		return nil, nil
	}
	userAddressImpl := &UserAddress{
		AddressIDAttr:    field.NewString(idutil.ULIDNow()),
		AddressTypeAttr:  field.NewString(pb.AddressType_HOME_ADDRESS.String()),
		PostalCodeAttr:   student.Address.PostalCode,
		CityAttr:         student.Address.City,
		FirstStreetAttr:  student.Address.FirstStreet,
		SecondStreetAttr: student.Address.SecondStreet,
	}
	return userAddressImpl, PrefectureImpl{PrefectureCodeAttr: student.Address.Prefecture}
}

func toEnrollmentStatusHistoriesV2(enrollmentStatus []EnrollmentStatusHistoryPayload) (entity.DomainEnrollmentStatusHistories, entity.DomainLocations) {
	domainEnrollmentStatusHistories := make(entity.DomainEnrollmentStatusHistories, 0, len(enrollmentStatus))
	domainLocations := make(entity.DomainLocations, 0, len(enrollmentStatus))
	for _, v := range enrollmentStatus {
		// in update student, if all enrollment status fields are empty, we don't need to update enrollment status
		if !field.IsPresent(v.EnrollmentStatus) && v.Location.IsEmpty() && !field.IsPresent(v.StartDate) {
			return entity.DomainEnrollmentStatusHistories{}, nil
		}

		enrollmentStatusStr := field.NewNullString()
		if field.IsPresent(v.EnrollmentStatus) {
			if value, ok := StudentEnrollmentStatusMap[int(v.EnrollmentStatus.Int16())]; ok {
				enrollmentStatusStr = field.NewString(value)
			}
		}
		startDate := v.StartDate.Date()
		if startDate.Format(constant.DateLayout) == time.Now().Format(constant.DateLayout) {
			startDate = v.StartDate.Date().Add(time.Since(v.StartDate.Date()))
		}
		if startDate.IsZero() {
			startDate = time.Now()
		}
		domainEnrollmentStatusHistories = append(domainEnrollmentStatusHistories, EnrollmentStatusHistoryImpl{
			EnrollmentStatusAttr: enrollmentStatusStr,
			StartDateAttr:        field.NewTime(startDate),
		})

		domainLocations = append(domainLocations, entity.LocationWillBeDelegated{
			HasPartnerInternalID: LocationImpl{LocationPartnerInternalAttr: v.Location},
		})
	}
	return domainEnrollmentStatusHistories, domainLocations
}

func toTaggedUsersV2(tagIDs []field.String, user DomainStudentImpl) (entity.DomainTaggedUsers, entity.DomainTags) {
	domainTaggedUsers := make(entity.DomainTaggedUsers, 0, len(tagIDs))
	domainTags := make(entity.DomainTags, 0, len(tagIDs))

	for _, tagID := range tagIDs {
		domainTaggedUsers = append(domainTaggedUsers, entity.TaggedUserWillBeDelegated{
			HasUserID: user,
		})
		domainTags = append(domainTags, entity.TagWillBeDelegated{
			HasPartnerInternalID: TagImpl{
				StudentTagAttr: tagID,
			},
		})
	}

	return domainTaggedUsers, domainTags
}

func toUserPhoneNumberV2(phoneNumber *PhoneNumberPayload) entity.DomainUserPhoneNumbers {
	if phoneNumber == nil {
		return nil
	}
	domainUserPhoneNumbers := make(entity.DomainUserPhoneNumbers, 0, 2)

	domainUserPhoneNumbers = append(domainUserPhoneNumbers, &UserPhoneNumber{
		PhoneIDAttr:     field.NewString(idutil.ULIDNow()),
		PhoneTypeAttr:   field.NewString(entity.StudentPhoneNumber),
		PhoneNumberAttr: phoneNumber.PhoneNumber,
	})

	domainUserPhoneNumbers = append(domainUserPhoneNumbers, &UserPhoneNumber{
		PhoneIDAttr:     field.NewString(idutil.ULIDNow()),
		PhoneTypeAttr:   field.NewString(entity.StudentHomePhoneNumber),
		PhoneNumberAttr: phoneNumber.HomePhoneNumber,
	})

	return domainUserPhoneNumbers
}

func toSchoolHistoryV2(schoolHistories []SchoolHistoryPayload) (entity.DomainSchoolHistories, entity.DomainSchools, entity.DomainSchoolCourses) {
	if len(schoolHistories) == 0 {
		return nil, nil, nil
	}

	domainSchoolHistories := make(entity.DomainSchoolHistories, 0, len(schoolHistories))
	domainSchoolsInfo := make(entity.DomainSchools, 0, len(schoolHistories))
	domainSchoolCourses := make(entity.DomainSchoolCourses, 0, len(schoolHistories))

	for _, schoolHistory := range schoolHistories {
		domainSchoolsInfo = append(domainSchoolsInfo, entity.SchoolWillBeDelegated{
			HasPartnerInternalID: SchoolInfoImpl{
				SchoolPartnerInternalIDAttr: schoolHistory.School,
			},
		})

		domainSchoolCourses = append(domainSchoolCourses, entity.SchoolCourseWillBeDelegated{
			HasPartnerInternalID: SchoolCourseImpl{
				SchoolCoursePartnerInternalIDAttr: schoolHistory.SchoolCourse,
			},
		})

		schoolHistory := &SchoolHistory{
			StartDateAttr: field.Time(schoolHistory.StartDate),
			EndDateAttr:   field.Time(schoolHistory.EndDate),
		}

		domainSchoolHistories = append(domainSchoolHistories, schoolHistory)
	}
	return domainSchoolHistories, domainSchoolsInfo, domainSchoolCourses
}

func (port *DomainStudentService) mapExternalUserIDAndUser(ctx context.Context, externalUserIDs []string) (map[string]entity.User, error) {
	existingUsers, err := port.DomainStudent.GetUsersByExternalIDs(ctx, externalUserIDs)
	if err != nil {
		return nil, err
	}

	mapExternalUserIDAndUser := make(map[string]entity.User, len(existingUsers))
	for _, user := range existingUsers {
		mapExternalUserIDAndUser[user.ExternalUserID().String()] = user
	}
	return mapExternalUserIDAndUser, nil
}

type GradeImpl struct {
	entity.NullDomainGrade

	GradeAttr field.String
}

func (g GradeImpl) PartnerInternalID() field.String {
	return g.GradeAttr
}

type EnrollmentStatusHistoryImpl struct {
	entity.DefaultDomainEnrollmentStatusHistory

	EnrollmentStatusAttr field.String
	LocationIDAttr       field.String
	UserIDAttr           field.String
	ResourcePathAttr     field.String
	StartDateAttr        field.Time
}

func (e EnrollmentStatusHistoryImpl) UserID() field.String {
	return e.UserIDAttr
}

func (e EnrollmentStatusHistoryImpl) EnrollmentStatus() field.String {
	return e.EnrollmentStatusAttr
}

func (e EnrollmentStatusHistoryImpl) StartDate() field.Time {
	switch {
	case field.IsUndefined(e.StartDateAttr):
		return e.DefaultDomainEnrollmentStatusHistory.StartDate()
	default:
		return e.StartDateAttr
	}
}

func (e EnrollmentStatusHistoryImpl) OrganizationID() field.String {
	return e.ResourcePathAttr
}

func (e EnrollmentStatusHistoryImpl) LocationID() field.String {
	return e.LocationIDAttr
}

type LocationImpl struct {
	entity.NullDomainLocation
	LocationAttr                field.String
	LocationPartnerInternalAttr field.String
}

func (location LocationImpl) PartnerInternalID() field.String {
	return location.LocationPartnerInternalAttr
}

type UserAddress struct {
	entity.DefaultDomainUserAddress

	AddressIDAttr    field.String
	AddressTypeAttr  field.String
	PostalCodeAttr   field.String
	PrefectureAttr   field.String
	CityAttr         field.String
	FirstStreetAttr  field.String
	SecondStreetAttr field.String
}

func (u UserAddress) UserAddressID() field.String {
	return u.AddressIDAttr
}

func (u UserAddress) AddressType() field.String {
	return u.AddressTypeAttr
}

func (u UserAddress) PostalCode() field.String {
	return u.PostalCodeAttr
}

func (u UserAddress) City() field.String {
	return u.CityAttr
}

func (u UserAddress) PrefectureID() field.String {
	return u.PrefectureAttr
}

func (u UserAddress) FirstStreet() field.String {
	return u.FirstStreetAttr
}

func (u UserAddress) SecondStreet() field.String {
	return u.SecondStreetAttr
}

type PrefectureImpl struct {
	entity.DefaultDomainPrefecture

	PrefectureCodeAttr field.String
}

func (u PrefectureImpl) PrefectureCode() field.String {
	return u.PrefectureCodeAttr
}

type TagImpl struct {
	entity.EmptyDomainTaggedUser

	StudentTagAttr field.String
}

func (t TagImpl) PartnerInternalID() field.String {
	return t.StudentTagAttr
}

type UserPhoneNumber struct {
	entity.DefaultDomainUserPhoneNumber

	PhoneIDAttr     field.String
	PhoneTypeAttr   field.String
	PhoneNumberAttr field.String
}

func (p UserPhoneNumber) UserPhoneNumberID() field.String {
	return p.PhoneIDAttr
}

func (p UserPhoneNumber) PhoneNumber() field.String {
	return p.PhoneNumberAttr
}

func (p UserPhoneNumber) Type() field.String {
	return p.PhoneTypeAttr
}

type SchoolInfoImpl struct {
	entity.DefaultDomainSchool

	SchoolIDAttr                field.String
	SchoolPartnerInternalIDAttr field.String
}

func (s SchoolInfoImpl) PartnerInternalID() field.String {
	return s.SchoolPartnerInternalIDAttr
}

type SchoolCourseImpl struct {
	entity.DefaultDomainSchoolCourse

	SchoolCourseIDAttr                field.String
	SchoolCoursePartnerInternalIDAttr field.String
}

func (s SchoolCourseImpl) PartnerInternalID() field.String {
	return s.SchoolCoursePartnerInternalIDAttr
}

type SchoolHistory struct {
	entity.DefaultDomainSchoolHistory

	SchoolIDAttr       field.String
	SchoolCourseIDAttr field.String
	StartDateAttr      field.Time
	EndDateAttr        field.Time
}

func (s SchoolHistory) StartDate() field.Time {
	switch {
	case field.IsUndefined(s.StartDateAttr):
		return s.DefaultDomainSchoolHistory.StartDate()
	default:
		return s.StartDateAttr
	}
}

func (s SchoolHistory) EndDate() field.Time {
	switch {
	case field.IsUndefined(s.EndDateAttr):
		return s.DefaultDomainSchoolHistory.EndDate()
	default:
		return s.EndDateAttr
	}
}

func (s SchoolHistory) SchoolID() field.String {
	return s.SchoolIDAttr
}

func (s SchoolHistory) SchoolCourseID() field.String {
	switch {
	case field.IsUndefined(s.SchoolCourseIDAttr):
		return s.DefaultDomainSchoolHistory.SchoolCourseID()
	default:
		return s.SchoolCourseIDAttr
	}
}
