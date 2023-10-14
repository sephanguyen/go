package importstudent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/aggregate"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/port/grpc"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/unleash"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/utils"
	pb "github.com/manabie-com/backend/pkg/manabuf/usermgmt/v2"

	"github.com/gocarina/gocsv"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DomainStudentService struct {
	DomainStudent interface {
		GetUsersByExternalIDs(ctx context.Context, externalUserIDs []string) (entity.Users, error)
		GetGradesByExternalIDs(ctx context.Context, externalIDs []string) ([]entity.DomainGrade, error)
		GetTagsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainTags, error)
		GetLocationsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainLocations, error)
		GetSchoolsByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchools, error)
		GetSchoolCoursesByExternalIDs(ctx context.Context, externalIDs []string) (entity.DomainSchoolCourses, error)
		GetPrefecturesByCodes(ctx context.Context, codes []string) ([]entity.DomainPrefecture, error)
		ValidateUpdateSystemAndExternalUserID(ctx context.Context, studentsToUpdate aggregate.DomainStudents) error
		UpsertMultiple(ctx context.Context, option unleash.DomainStudentFeatureOption, studentsToCreate ...aggregate.DomainStudent) ([]aggregate.DomainStudent, error)
		GetEmailWithStudentID(ctx context.Context, studentIDs []string) (map[string]entity.User, error)
		IsFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled(organization valueobj.HasOrganizationID) bool
		UpsertMultipleWithErrorCollection(ctx context.Context, domainStudents aggregate.DomainStudents, option unleash.DomainStudentFeatureOption) (aggregate.DomainStudents, []error)
		IsFeatureUserNameStudentParentEnabled(organization valueobj.HasOrganizationID) bool
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

func (d *DomainStudentService) ImportStudentV2(ctx context.Context, req *pb.ImportStudentRequest) (*pb.UpsertStudentResponse, error) {
	zapLogger := ctxzap.Extract(ctx)
	organization, err := interceptors.OrganizationFromContext(ctx)
	if err != nil {
		wrapErr := errors.Wrap(err, "failed to get organization from context")
		newErr := grpc.InternalError{
			RawErr: wrapErr,
		}
		zapLogger.Error(newErr.Error(),
			zap.Error(newErr),
		)
		return nil, errorx.GRPCErr(newErr, &pb.ErrorMessages{Messages: []*pb.ErrorMessage{{
			Error: newErr.DomainError(),
			Code:  int32(newErr.DomainCode()),
		}}})
	}
	isEnableIgnoreError := d.DomainStudent.IsFeatureIgnoreInvalidRecordsCSVAndOpenAPIEnabled(organization)

	option := unleash.DomainStudentFeatureOption{
		DomainUserFeatureOption: unleash.DomainUserFeatureOption{
			EnableIgnoreUpdateEmail: true,
		},
	}
	option = d.FeatureManager.FeatureUsernameToStudentFeatureOption(ctx, organization, option)
	option = d.FeatureManager.FeatureDisableAutoDeactivateStudentsToStudentFeatureOption(ctx, organization, option)
	option = d.FeatureManager.FeatureAutoDeactivateAndReactivateStudentsV2ToStudentFeatureOption(ctx, organization, option)
	option = d.FeatureManager.FeatureExperimentalBulkInsertEnrollmentStatusHistoriesToStudentFeatureOption(ctx, organization, option)

	if err := readAndValidatePayload(req.Payload); err != nil {
		if isEnableIgnoreError {
			newErr, _ := err.(errcode.DomainError)
			zapLogger.Error(newErr.Error(),
				zap.Error(newErr),
			)
			return nil, errorx.GRPCErr(err, &pb.ErrorMessages{Messages: []*pb.ErrorMessage{
				{
					Error: newErr.DomainError(),
					Code:  int32(newErr.DomainCode()),
				},
			}})
		}

		oldErr := errcode.Error{
			Code: errcode.InternalError,
			Err:  err,
		}
		zapLogger.Error(oldErr.Error(),
			zap.Error(oldErr.Err),
		)
		return nil, errorx.GRPCErr(oldErr, errorx.PbErrorMessage(oldErr))
	}
	importStudentCSVFields, err := ConvertPayloadToImportStudentData(req.Payload)
	if err != nil {
		if isEnableIgnoreError {
			newErr, _ := err.(errcode.DomainError)
			zapLogger.Error(newErr.Error(),
				zap.Error(newErr),
			)
			return nil, errorx.GRPCErr(newErr, &pb.ErrorMessages{Messages: []*pb.ErrorMessage{
				{
					Error: newErr.DomainError(),
					Code:  int32(newErr.DomainCode()),
				},
			}})
		}

		oldErr := errcode.Error{
			Code: errcode.InternalError,
			Err:  err,
		}
		zapLogger.Error(oldErr.Error(),
			zap.Error(oldErr.Err),
		)
		return nil, errorx.GRPCErr(oldErr, errorx.PbErrorMessage(oldErr))
	}
	if len(importStudentCSVFields) == 0 {
		return &pb.UpsertStudentResponse{}, nil
	}

	if len(importStudentCSVFields) > constant.LimitRowsCSV {
		if isEnableIgnoreError {
			newErr := grpc.MaximumRowsCSVError{
				RequestRows:  len(importStudentCSVFields),
				LimitRowsCSV: constant.LimitRowsCSV,
			}
			zapLogger.Error(newErr.Error(),
				zap.Error(newErr),
			)
			return nil, errorx.GRPCErr(newErr, &pb.ErrorMessages{Messages: []*pb.ErrorMessage{
				{
					Error: newErr.DomainError(),
					Code:  int32(newErr.DomainCode()),
				},
			}})
		}

		oldErr := errcode.Error{
			Code: errcode.InvalidMaximumRows,
		}
		zapLogger.Error(oldErr.Error(),
			zap.Error(oldErr.Err),
		)
		return nil, errorx.GRPCErr(oldErr, errorx.PbErrorMessage(oldErr))
	}

	// Implement ignore error
	if isEnableIgnoreError {
		mapStudentIDAndStudent := make(map[string]entity.User, 0)
		if !option.EnableUsername {
			studentIDs := make([]string, 0, len(importStudentCSVFields))
			for _, student := range importStudentCSVFields {
				studentIDs = append(studentIDs, student.UserID().String())
			}
			mapStudentIDAndStudent, err = d.DomainStudent.GetEmailWithStudentID(ctx, studentIDs)
			if err != nil {
				e := grpc.InternalError{
					RawErr: errors.WithStack(errors.Wrap(err, "failed to get email with student id")),
				}
				zapLogger.Error(e.Error(),
					zap.Error(e),
				)
				return nil, errorx.GRPCErr(err, &pb.ErrorMessage{
					Code:  int32(e.DomainCode()),
					Error: e.DomainError(),
				})
			}
		}

		domainStudents := make(aggregate.DomainStudents, 0)
		collectionErrors := make([]error, 0, len(importStudentCSVFields))
		for idx, student := range importStudentCSVFields {
			if len(mapStudentIDAndStudent) > 0 && !student.UserID().IsEmpty() {
				if user, ok := mapStudentIDAndStudent[student.UserID().String()]; ok {
					student.EmailAttr = user.Email()
					student.LoginEmailAttr = user.LoginEmail()
				}
			}
			domainStudent, err := ToDomainStudentsV2(student, idx, option.EnableUsername)
			if err != nil {
				collectionErrors = append(collectionErrors, err)
				continue
			}
			domainStudents = append(domainStudents, domainStudent)
		}

		respStudents, listOfErrors := d.DomainStudent.UpsertMultipleWithErrorCollection(ctx, domainStudents, option)
		if len(listOfErrors) > 0 {
			collectionErrors = append(collectionErrors, listOfErrors...)
		}

		if len(collectionErrors) > 0 {
			respErrors := make([]*pb.ErrorMessage, 0, len(collectionErrors))
			for _, e := range collectionErrors {
				switch err := e.(type) {
				case errcode.DomainError:
					respErrors = append(respErrors, grpc.ToPbErrorMessageImport(err))
				case errcode.Error:
					respErrors = append(respErrors, errorx.PbErrorMessage(err))
				default:
					respErrors = append(respErrors, &pb.ErrorMessage{
						Error: err.Error(),
						Code:  int32(errcode.InternalError),
					})
				}
			}
			zapLogger.Error("some students were failed when import student csv",
				zap.Errors("list of errors", collectionErrors))
			return &pb.UpsertStudentResponse{Messages: respErrors}, nil
		}
		return &pb.UpsertStudentResponse{StudentProfiles: grpc.UpsertStudentProfiles(respStudents)}, nil
	}

	// Don't implement ignore error
	domainStudents, err := d.ToDomainStudents(ctx, importStudentCSVFields, option.EnableUsername)
	if err != nil {
		switch e := err.(type) {
		case errcode.Error:
			return nil, errorx.GRPCErr(err, errorx.PbErrorMessage(err))
		case errcode.DomainError:
			return nil, errorx.GRPCErr(e, grpc.ToPbErrorMessageImport(e))
		}
	}
	students, err := d.DomainStudent.UpsertMultiple(ctx, option, domainStudents...)
	if err != nil {
		switch e := err.(type) {
		case errcode.Error:
			e.Index += 2
			return nil, errorx.GRPCErr(err, errorx.PbErrorMessage(e))
		case errcode.DomainError:
			return nil, errorx.GRPCErr(e, grpc.ToPbErrorMessageImport(e))
		}
	}
	return &pb.UpsertStudentResponse{StudentProfiles: grpc.UpsertStudentProfiles(students)}, nil
}

func (d *DomainStudentService) ToDomainStudents(ctx context.Context, studentCSV []*StudentCSV, isEnableUsername bool) ([]aggregate.DomainStudent, error) {
	domainStudents := make([]aggregate.DomainStudent, 0, len(studentCSV))
	for i, student := range studentCSV {
		// `csvRowIndex` is only for error response, do not use this for other purposes
		csvRowIndex := i + 2
		if field.IsUndefined(student.UserID()) {
			err := errcode.Error{
				Code:      errcode.MissingField,
				FieldName: string(entity.UserFieldUserID),
				Err:       fmt.Errorf("user_id does not exist"),
				Index:     csvRowIndex,
			}
			return nil, errorx.GRPCErr(err)
		}
		if !isEnableUsername {
			student.UserNameAttr = student.EmailAttr
		}

		// trim external_user_id
		student.ExternalUserIDAttr = toExternalUserID(student)

		if err := d.validateUserID(ctx, student, csvRowIndex); err != nil {
			return nil, err
		}

		fullName, fullNamePhonetic := toFullNameAndFullNamePhonetic(student)
		student.FullNameAttr = fullName
		student.FullNamePhoneticAttr = fullNamePhonetic
		student.LoginEmailAttr = student.EmailAttr
		birthday, err := toBirthDay(student.BirthdayAttr, csvRowIndex)
		if err != nil {
			return nil, err
		}
		student.birthday = birthday

		gender, err := ToGender(student.GenderAttr, csvRowIndex)
		if err != nil {
			return nil, err
		}
		student.GenderAttr = gender

		gradePartnerID, err := d.toInternalGrade(ctx, student, csvRowIndex)
		if err != nil {
			return nil, err
		}
		student.GradeAttr = gradePartnerID

		addressAttr, err := d.toUserAddress(ctx, student, csvRowIndex)
		if err != nil {
			return nil, err
		}

		contactPreference, err := d.toContactPreference(student.StudentContactPreferenceAttr, csvRowIndex)
		if err != nil {
			return nil, err
		}
		student.StudentContactPreferenceAttr = contactPreference

		domainUserAccessPaths, err := d.toUserAccessPaths(ctx, student.LocationAttr, student, csvRowIndex)
		if err != nil {
			return nil, err
		}

		domainTaggedUsers, err := d.toTaggedUsers(ctx, student.StudentTagAttr, student, csvRowIndex)
		if err != nil {
			return nil, err
		}

		domainSchoolHistories, err := d.toSchoolHistory(ctx, student, csvRowIndex)
		if err != nil {
			return nil, err
		}

		enrollmentStatusHistories, err := d.toEnrollmentStatusHistories(ctx, student, csvRowIndex)
		if err != nil {
			return nil, err
		}
		if len(enrollmentStatusHistories) > 0 {
			enrollmentStatusHistory := enrollmentStatusHistories[0]
			student.EnrollmentStatusAttr = enrollmentStatusHistory.EnrollmentStatus()
		}
		domainStudents = append(domainStudents, aggregate.DomainStudent{
			DomainStudent:             student,
			UserAddress:               addressAttr,
			UserPhoneNumbers:          toUserPhoneNumbers(student),
			UserAccessPaths:           domainUserAccessPaths,
			TaggedUsers:               domainTaggedUsers,
			SchoolHistories:           domainSchoolHistories,
			EnrollmentStatusHistories: enrollmentStatusHistories,
			IndexAttr:                 i,
		})
	}
	if !isEnableUsername {
		var err error
		domainStudents, err = d.fillExistedEmailOfUsers(ctx, domainStudents)
		if err != nil {
			return nil, err
		}
	}

	return domainStudents, nil
}

func ToDomainStudentsV2(student *StudentCSV, index int, isEnableUsername bool) (aggregate.DomainStudent, error) {
	if field.IsUndefined(student.UserID()) {
		err := entity.MissingMandatoryFieldError{
			FieldName:  string(entity.UserFieldUserID),
			EntityName: entity.StudentEntity,
			Index:      index,
		}
		return aggregate.DomainStudent{}, err
	}
	if !isEnableUsername {
		student.UserNameAttr = student.EmailAttr
	}
	// trim external_user_id
	student.ExternalUserIDAttr = toExternalUserID(student)

	fullName, fullNamePhonetic := toFullNameAndFullNamePhonetic(student)
	student.FullNameAttr = fullName
	student.FullNamePhoneticAttr = fullNamePhonetic
	if student.LoginEmailAttr.String() == "" {
		student.LoginEmailAttr = student.EmailAttr
	}

	birthday, err := toBirthDayV2(student.BirthdayAttr, index)
	if err != nil {
		return aggregate.DomainStudent{}, err
	}
	student.birthday = birthday

	gender, err := ToGenderV2(student.GenderAttr, index)
	if err != nil {
		return aggregate.DomainStudent{}, err
	}
	student.GenderAttr = gender

	addressAttr, prefecture := toUserAddressV2(student)

	contactPreference, err := toContactPreferenceV2(student.StudentContactPreferenceAttr, index)
	if err != nil {
		return aggregate.DomainStudent{}, err
	}

	student.StudentContactPreferenceAttr = contactPreference

	gradeDomain := toInternalGradeV2(student)

	domainUserAccessPaths := toUserAccessPathsV2(student.LocationAttr, student)

	domainTaggedUsers, domainTags := toTaggedUsersV2(student.StudentTagAttr, student)

	domainSchoolHistories, domainSchoolsInfo, domainSchoolCourses, err := toSchoolHistoryV2(student, index)
	if err != nil {
		return aggregate.DomainStudent{}, err
	}

	enrollmentStatusHistories, domainLocations, err := toEnrollmentStatusHistoriesV2(student, index)
	if err != nil {
		return aggregate.DomainStudent{}, err
	}
	if len(enrollmentStatusHistories) > 0 {
		enrollmentStatusHistory := enrollmentStatusHistories[0]
		student.EnrollmentStatusAttr = enrollmentStatusHistory.EnrollmentStatus()
	}

	return aggregate.DomainStudent{
		DomainStudent:             student,
		UserAddress:               addressAttr,
		UserPhoneNumbers:          toUserPhoneNumbers(student),
		UserAccessPaths:           domainUserAccessPaths,
		TaggedUsers:               domainTaggedUsers,
		Tags:                      domainTags,
		SchoolHistories:           domainSchoolHistories,
		SchoolInfos:               domainSchoolsInfo,
		SchoolCourses:             domainSchoolCourses,
		EnrollmentStatusHistories: enrollmentStatusHistories,
		Grade:                     gradeDomain,
		Locations:                 domainLocations,
		Prefecture:                prefecture,
		IndexAttr:                 index,
	}, nil
}

func (d *DomainStudentService) toUserAddress(ctx context.Context, student *StudentCSV, csvIndex int) (*UserAddress, error) {
	userAddressImpl := &UserAddress{
		AddressIDAttr:    field.NewString(idutil.ULIDNow()),
		AddressTypeAttr:  field.NewString(pb.AddressType_HOME_ADDRESS.String()),
		PostalCodeAttr:   student.PostalCodeAttr,
		CityAttr:         student.CityAttr,
		FirstStreetAttr:  student.FirstStreetAttr,
		SecondStreetAttr: student.SecondStreetAttr,
	}

	if field.IsPresent(student.PrefectureAttr) {
		prefectureCodes, err := d.DomainStudent.GetPrefecturesByCodes(ctx, []string{student.PrefectureAttr.String()})
		if err != nil {
			return nil, errcode.Error{
				Code:      errcode.InternalError,
				Err:       err,
				FieldName: entity.StudentUserAddressPrefectureField,
				Index:     csvIndex,
			}
		}
		if len(prefectureCodes) == 0 {
			return nil, errcode.Error{
				Code:      errcode.InvalidData,
				Err:       err,
				FieldName: entity.StudentUserAddressPrefectureField,
				Index:     csvIndex,
			}
		}
		userAddressImpl.PrefectureAttr = prefectureCodes[0].PrefectureID()
	}

	return userAddressImpl, nil
}

func (d *DomainStudentService) toContactPreference(reference field.String, csvIndex int) (field.String, error) {
	contactPreference := field.NewNullString()
	if field.IsPresent(reference) {
		contactPreferenceValue, ok := mapStudentContactPreference[reference.String()]
		if !ok {
			return contactPreference, errcode.Error{
				Code:      errcode.InvalidData,
				FieldName: entity.StudentFieldContactPreference,
				Index:     csvIndex,
			}
		}
		contactPreference = field.NewString(contactPreferenceValue)
	}
	return contactPreference, nil
}

func toContactPreferenceV2(reference field.String, index int) (field.String, error) {
	contactPreference := field.NewNullString()
	if field.IsPresent(reference) {
		contactPreferenceValue, ok := mapStudentContactPreference[reference.String()]
		if !ok {
			return contactPreference, entity.InvalidFieldError{
				FieldName:  entity.StudentFieldContactPreference,
				EntityName: entity.StudentEntity,
				Index:      index,
				Reason:     entity.NotMatchingEnum,
			}
		}
		contactPreference = field.NewString(contactPreferenceValue)
	}
	return contactPreference, nil
}

func (d *DomainStudentService) toUserAccessPaths(ctx context.Context, location field.String, student *StudentCSV, csvIndex int) (entity.DomainUserAccessPaths, error) {
	locationIDs := strings.Split(location.String(), constant.ArraySeparatorCSV)
	domainUserAccessPaths := make(entity.DomainUserAccessPaths, 0, len(locationIDs))

	if field.IsPresent(location) {
		locations, err := d.DomainStudent.GetLocationsByExternalIDs(ctx, locationIDs)
		if err != nil {
			return nil, errcode.Error{
				Code:      errcode.InternalError,
				Err:       err,
				FieldName: entity.StudentLocationsField,
				Index:     csvIndex,
			}
		}
		domainUserAccessPaths = locations.ToUserAccessPath(student)
	}

	return domainUserAccessPaths, nil
}

func (d *DomainStudentService) toInternalGrade(ctx context.Context, importStudentCSVField *StudentCSV, csvIndex int) (field.String, error) {
	if !field.IsPresent(importStudentCSVField.GradeAttr) {
		return field.NewNullString(), nil
	}

	grades, err := d.DomainStudent.GetGradesByExternalIDs(ctx, []string{importStudentCSVField.GradeAttr.String()})
	if err != nil {
		return field.NewNullString(), errcode.Error{
			Code:      errcode.InternalError,
			FieldName: entity.StudentGradeField,
			Index:     csvIndex,
		}
	}

	if len(grades) == 0 {
		return field.NewNullString(), errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: entity.StudentGradeField,
			Err:       fmt.Errorf("grade not found"),
			Index:     csvIndex,
		}
	}

	return grades[0].GradeID(), nil
}

func toInternalGradeV2(importStudentCSVField *StudentCSV) entity.DomainGrade {
	if field.IsPresent(importStudentCSVField.GradeAttr) {
		return entity.GradeWillBeDelegated{
			HasPartnerInternalID: GradeImpl{
				GradeAttr: field.NewString(importStudentCSVField.GradeAttr.String()),
			},
		}
	}
	return entity.NullDomainGrade{}
}

func toUserAddressV2(student *StudentCSV) (entity.DomainUserAddress, entity.DomainPrefecture) {
	userAddressImpl := &UserAddress{
		AddressIDAttr:    field.NewString(idutil.ULIDNow()),
		AddressTypeAttr:  field.NewString(pb.AddressType_HOME_ADDRESS.String()),
		PostalCodeAttr:   student.PostalCodeAttr,
		CityAttr:         student.CityAttr,
		FirstStreetAttr:  student.FirstStreetAttr,
		SecondStreetAttr: student.SecondStreetAttr,
	}
	return userAddressImpl, PrefectureImpl{PrefectureCodeAttr: student.PrefectureAttr}
}

func toUserAccessPathsV2(locationCSV field.String, student *StudentCSV) entity.DomainUserAccessPaths {
	locationPartnerIDs := strings.Split(locationCSV.String(), constant.ArraySeparatorCSV)
	domainUserAccessPaths := make(entity.DomainUserAccessPaths, 0, len(locationPartnerIDs))

	for range locationPartnerIDs {
		domainUserAccessPaths = append(domainUserAccessPaths, entity.UserAccessPathWillBeDelegated{
			HasUserID: student,
		})
	}
	return domainUserAccessPaths
}

func (d *DomainStudentService) toTaggedUsers(ctx context.Context, tagField field.String, user *StudentCSV, csvIndex int) (entity.DomainTaggedUsers, error) {
	tagIDs := strings.Split(tagField.String(), constant.ArraySeparatorCSV)
	domainTaggedUsers := make(entity.DomainTaggedUsers, 0, len(tagIDs))

	if field.IsPresent(tagField) {
		if isDuplicatedIDs(tagIDs) {
			return nil, errcode.Error{
				Code:      errcode.DuplicatedData,
				FieldName: entity.StudentTagsField,
				Index:     csvIndex,
			}
		}

		domainTags, err := d.DomainStudent.GetTagsByExternalIDs(ctx, tagIDs)
		if err != nil {
			// TODO: if repo find records, we should use zap logger
			return nil, errcode.Error{
				Code:      errcode.InternalError,
				FieldName: entity.StudentTagsField,
				Index:     csvIndex,
			}
		}

		if err := importCsvValidateTag(constant.RoleStudent, tagIDs, domainTags); err != nil {
			return nil, errcode.Error{
				Code:      errcode.InvalidData,
				Err:       err,
				FieldName: entity.StudentTagsField,
				Index:     csvIndex,
			}
		}

		domainTaggedUsers = domainTags.ToTaggedUser(user)
	}
	return domainTaggedUsers, nil
}

func (d *DomainStudentService) toEnrollmentStatusHistories(ctx context.Context, importStudentCSVField *StudentCSV, csvIndex int) (entity.DomainEnrollmentStatusHistories, error) {
	// in update student, if all enrollment status fields are empty, we don't need to update enrollment status

	if !field.IsPresent(importStudentCSVField.EnrollmentStatus()) &&
		!field.IsPresent(importStudentCSVField.LocationAttr) &&
		!field.IsPresent(importStudentCSVField.StatusStartDateAttr) {
		return entity.DomainEnrollmentStatusHistories{}, nil
	}

	location := importStudentCSVField.LocationAttr
	enrollmentStatus := importStudentCSVField.EnrollmentStatus()
	statusStartDate := importStudentCSVField.StatusStartDateAttr

	if err := validateCSVEnrollmentStatusFields(location, enrollmentStatus, statusStartDate, csvIndex); err != nil {
		return nil, err
	}

	locationPartnerIDs := utils.SplitWithCapacity(location.String(), constant.ArraySeparatorCSV, 0)
	enrollmentStatuses := utils.SplitWithCapacity(enrollmentStatus.String(), constant.ArraySeparatorCSV, len(locationPartnerIDs))
	statusStartDates := utils.SplitWithCapacity(statusStartDate.String(), constant.ArraySeparatorCSV, len(locationPartnerIDs))

	domainLocations, err := d.DomainStudent.GetLocationsByExternalIDs(ctx, locationPartnerIDs)
	if err != nil {
		return nil, errcode.Error{
			Code:      errcode.InternalError,
			Err:       err,
			FieldName: entity.StudentLocationsField,
			Index:     csvIndex,
		}
	}
	if len(domainLocations) != len(locationPartnerIDs) {
		return nil, errcode.Error{
			Code:      errcode.InvalidData,
			Err:       errcode.ErrUserLocationsAreInvalid,
			FieldName: entity.StudentLocationsField,
			Index:     csvIndex,
		}
	}

	enrollmentStatusHistories := make(entity.DomainEnrollmentStatusHistories, 0, len(domainLocations))

	locationByPartnerID := mapLocationByPartnerID(domainLocations)
	for idx, locationPartnerID := range locationPartnerIDs {
		domainLocation := locationByPartnerID[locationPartnerID]

		startDate := field.NewNullTime()
		if err := startDate.UnmarshalCSV(statusStartDates[idx]); err != nil {
			return nil, errcode.Error{
				Code:      errcode.InvalidData,
				Err:       err,
				FieldName: entity.StudentFieldEnrollmentStatusStartDate,
				Index:     csvIndex,
			}
		}

		enrollmentStatusInt := field.NewNullInt16()
		if err := enrollmentStatusInt.UnmarshalCSV(enrollmentStatuses[idx]); err != nil {
			return nil, errcode.Error{
				Code:      errcode.InvalidData,
				Err:       err,
				FieldName: entity.StudentFieldEnrollmentStatus,
				Index:     csvIndex,
			}
		}

		enrollmentStatusStr := field.NewNullString()
		if field.IsPresent(enrollmentStatusInt) {
			enrollmentStatusStr = field.NewString(studentEnrollmentStatusMap[enrollmentStatusInt.Int16()])
		}

		enrollmentStatusHistories = append(enrollmentStatusHistories, EnrollmentStatusHistory{
			EnrollmentStatusAttr: enrollmentStatusStr,
			LocationIDAttr:       domainLocation.LocationID(),
			StartDateAttr:        startDate,
		})
	}

	return enrollmentStatusHistories, nil
}

func toTaggedUsersV2(tagField field.String, user *StudentCSV) (entity.DomainTaggedUsers, entity.DomainTags) {
	tagIDs := strings.Split(tagField.String(), constant.ArraySeparatorCSV)
	domainTaggedUsers := make(entity.DomainTaggedUsers, 0, len(tagIDs))
	domainTags := make(entity.DomainTags, 0, len(tagIDs))

	if field.IsPresent(tagField) {
		for _, tagID := range tagIDs {
			domainTaggedUsers = append(domainTaggedUsers, entity.TaggedUserWillBeDelegated{
				HasUserID: user,
			})
			domainTags = append(domainTags, entity.TagWillBeDelegated{
				HasPartnerInternalID: TagImpl{
					StudentTagAttr: field.NewString(tagID),
				},
			})
		}
	}
	return domainTaggedUsers, domainTags
}

func toEnrollmentStatusHistoriesV2(importStudentCSVField *StudentCSV, index int) (entity.DomainEnrollmentStatusHistories, entity.DomainLocations, error) {
	// in update student, if all enrollment status fields are empty, we don't need to update enrollment status
	if !field.IsPresent(importStudentCSVField.EnrollmentStatusAttr) &&
		!field.IsPresent(importStudentCSVField.LocationAttr) &&
		!field.IsPresent(importStudentCSVField.StatusStartDateAttr) {
		return entity.DomainEnrollmentStatusHistories{}, nil, nil
	}

	location := importStudentCSVField.LocationAttr
	enrollmentStatus := importStudentCSVField.EnrollmentStatusAttr
	statusStartDate := importStudentCSVField.StatusStartDateAttr

	locationPartnerIDs := utils.SplitWithCapacity(location.String(), constant.ArraySeparatorCSV, 0)
	enrollmentStatuses := utils.SplitWithCapacity(enrollmentStatus.String(), constant.ArraySeparatorCSV, len(locationPartnerIDs))
	statusStartDates := utils.SplitWithCapacity(statusStartDate.String(), constant.ArraySeparatorCSV, len(locationPartnerIDs))
	enrollmentStatusHistories := make(entity.DomainEnrollmentStatusHistories, 0, len(locationPartnerIDs))

	maxLength := len(locationPartnerIDs)

	if len(enrollmentStatuses) > maxLength {
		maxLength = len(enrollmentStatuses)
	}

	domainLocations := make(entity.DomainLocations, 0, len(locationPartnerIDs))

	for idx := 0; idx < maxLength; idx++ {
		startDate := field.NewTime(time.Now())
		if len(statusStartDates) > idx {
			if statusStartDates[idx] != "" {
				err := startDate.UnmarshalCSV(statusStartDates[idx])
				if err != nil {
					return nil, nil, entity.InvalidFieldError{
						FieldName:  entity.StudentFieldEnrollmentStatusStartDate,
						Index:      index,
						EntityName: entity.StudentEntity,
						Reason:     entity.FailedUnmarshal,
					}
				}
				if startDate.Time().Format(constant.DateLayout) == time.Now().Format(constant.DateLayout) {
					startDateAddedNow := startDate.Time().Add(time.Since(startDate.Time()))
					startDate = field.NewTime(startDateAddedNow)
				}
			}
		}

		enrollmentStatusInt := field.NewNullInt16()
		if len(enrollmentStatuses) > idx {
			if err := enrollmentStatusInt.UnmarshalCSV(enrollmentStatuses[idx]); err != nil {
				return nil, nil, entity.InvalidFieldError{
					FieldName:  entity.StudentFieldEnrollmentStatus,
					Index:      index,
					EntityName: entity.StudentEntity,
					Reason:     entity.FailedUnmarshal,
				}
			}
		}

		enrollmentStatusStr := field.NewNullString()
		if field.IsPresent(enrollmentStatusInt) {
			if value, ok := studentEnrollmentStatusMap[enrollmentStatusInt.Int16()]; ok {
				enrollmentStatusStr = field.NewString(value)
			}
		}

		enrollmentStatusHistories = append(enrollmentStatusHistories, EnrollmentStatusHistory{
			EnrollmentStatusAttr: enrollmentStatusStr,
			StartDateAttr:        startDate,
		})

		locationPartnerID := field.NewNullString()
		if len(locationPartnerIDs) > idx {
			locationPartnerID = field.NewString(locationPartnerIDs[idx])
		}

		domainLocations = append(domainLocations, entity.LocationWillBeDelegated{
			HasPartnerInternalID: LocationImpl{LocationPartnerInternalAttr: locationPartnerID},
		})
	}

	return enrollmentStatusHistories, domainLocations, nil
}

func mapTwoArraysWithKeyValue(arrKey []string, arrValue []string) (map[string]string, error) {
	if len(arrKey) != len(arrValue) {
		return nil, fmt.Errorf("invalid data")
	}

	mapData := map[string]string{}

	for idx, key := range arrKey {
		if arrValue[idx] != "" {
			mapData[key] = arrValue[idx]
		}
	}

	return mapData, nil
}

func (d *DomainStudentService) toSchoolHistory(ctx context.Context, importStudentCSVField *StudentCSV, csvIndex int) (entity.DomainSchoolHistories, error) {
	// init err validation base
	errValidation := errcode.Error{
		Code:  errcode.InvalidData,
		Index: csvIndex,
	}

	if !field.IsPresent(importStudentCSVField.SchoolAttr) &&
		!field.IsPresent(importStudentCSVField.SchoolCourseAttr) &&
		!field.IsPresent(importStudentCSVField.StartDateAttr) &&
		!field.IsPresent(importStudentCSVField.EndDateAttr) {
		return nil, nil
	}

	if !field.IsPresent(importStudentCSVField.SchoolAttr) {
		if field.IsPresent(importStudentCSVField.SchoolCourseAttr) {
			errValidation.FieldName = entity.StudentSchoolCourseField
			return nil, errValidation
		}
		if field.IsPresent(importStudentCSVField.StartDateAttr) {
			errValidation.FieldName = entity.StudentSchoolHistoryStartDateField
			return nil, errValidation
		}
		if field.IsPresent(importStudentCSVField.EndDateAttr) {
			errValidation.FieldName = entity.StudentSchoolHistoryEndDateField
			return nil, errValidation
		}
	}

	// require
	schoolCSV := importStudentCSVField.SchoolAttr.String()
	schoolPartnerIDs := strings.Split(schoolCSV, constant.ArraySeparatorCSV)
	for _, schoolPartnerID := range schoolPartnerIDs {
		if schoolPartnerID == "" {
			errValidation.FieldName = entity.StudentSchoolField
			return nil, errValidation
		}
	}

	// optional course
	schoolCourseCSV := importStudentCSVField.SchoolCourseAttr.String()
	schoolCoursePartnerIDs := strings.Split(schoolCourseCSV, constant.ArraySeparatorCSV)

	schoolCoursePartnerIDBySchoolPartnerID, err := mapTwoArraysWithKeyValue(schoolPartnerIDs, schoolCoursePartnerIDs)
	if err != nil {
		errValidation.FieldName = entity.StudentSchoolCourseField
		return nil, errValidation
	}

	// optional start date
	startDateCSV := importStudentCSVField.StartDateAttr.String()
	startDateArr := strings.Split(startDateCSV, constant.ArraySeparatorCSV)

	startDates, err := mapTwoArraysWithKeyValue(schoolPartnerIDs, startDateArr)
	if err != nil {
		errValidation.FieldName = entity.StudentSchoolHistoryStartDateField
		return nil, errValidation
	}

	// optional end date
	endDateCSV := importStudentCSVField.EndDateAttr.String()
	endDateArr := strings.Split(endDateCSV, constant.ArraySeparatorCSV)

	endDates, err := mapTwoArraysWithKeyValue(schoolPartnerIDs, endDateArr)
	if err != nil {
		errValidation.FieldName = entity.StudentSchoolHistoryEndDateField
		return nil, errValidation
	}

	schoolInfos, err := d.DomainStudent.GetSchoolsByExternalIDs(ctx, schoolPartnerIDs)
	if err != nil {
		return nil, errcode.Error{
			Code:      errcode.InternalError,
			Err:       err,
			FieldName: entity.StudentSchoolField,
			Index:     csvIndex,
		}
	}
	if len(schoolInfos) != len(schoolPartnerIDs) {
		errValidation.FieldName = entity.StudentSchoolField
		return nil, errValidation
	}

	schoolCourseIDBySchoolPartnerID := map[string]string{}
	if len(schoolCoursePartnerIDBySchoolPartnerID) != 0 {
		schoolCoursePartnerIds := make([]string, 0, len(schoolCoursePartnerIDBySchoolPartnerID))
		for _, id := range schoolCoursePartnerIDBySchoolPartnerID {
			schoolCoursePartnerIds = append(schoolCoursePartnerIds, id)
		}
		schoolCoursesEntities, err := d.DomainStudent.GetSchoolCoursesByExternalIDs(ctx, schoolCoursePartnerIds)
		if err != nil {
			return nil, errcode.Error{
				Code:      errcode.InternalError,
				Err:       err,
				FieldName: entity.StudentSchoolCourseField,
				Index:     csvIndex,
			}
		}
		if len(schoolCoursesEntities) != len(schoolCoursePartnerIDBySchoolPartnerID) {
			errValidation.FieldName = entity.StudentSchoolCourseField
			return nil, errValidation
		}
		// map schoolCoursePartnerId to schoolCourseId
		for _, course := range schoolCoursesEntities {
			for key, schoolCoursePartnerID := range schoolCoursePartnerIDBySchoolPartnerID {
				if schoolCoursePartnerID == course.PartnerInternalID().String() {
					schoolCourseIDBySchoolPartnerID[key] = course.SchoolCourseID().String()
				}
			}
		}
	}

	schoolHistories := make(entity.DomainSchoolHistories, 0, len(schoolPartnerIDs))
	for _, schoolInfo := range schoolInfos {
		schoolHistory := &SchoolHistory{}

		schoolHistory.SchoolIDAttr = schoolInfo.SchoolID()
		schoolPartnerID := schoolInfo.PartnerInternalID().String()
		if _, ok := schoolCourseIDBySchoolPartnerID[schoolPartnerID]; ok {
			schoolHistory.SchoolCourseIDAttr = field.NewString(schoolCourseIDBySchoolPartnerID[schoolPartnerID])
		}

		if _, ok := startDates[schoolPartnerID]; ok {
			startDate := field.NewNullTime()
			if err := startDate.UnmarshalCSV(startDates[schoolPartnerID]); err != nil {
				errValidation.FieldName = entity.StudentSchoolHistoryStartDateField
				return nil, errValidation
			}
			schoolHistory.StartDateAttr = startDate
		}

		if _, ok := endDates[schoolPartnerID]; ok {
			endDate := field.NewNullTime()
			if err := endDate.UnmarshalCSV(endDates[schoolPartnerID]); err != nil {
				errValidation.FieldName = entity.StudentSchoolHistoryEndDateField
				return nil, errValidation
			}
			schoolHistory.EndDateAttr = endDate
		}

		schoolHistories = append(schoolHistories, schoolHistory)
	}
	return schoolHistories, nil
}

func toSchoolHistoryV2(importStudentCSVField *StudentCSV, index int) (entity.DomainSchoolHistories, entity.DomainSchools, entity.DomainSchoolCourses, error) {
	if !field.IsPresent(importStudentCSVField.SchoolAttr) &&
		!field.IsPresent(importStudentCSVField.SchoolCourseAttr) &&
		!field.IsPresent(importStudentCSVField.StartDateAttr) &&
		!field.IsPresent(importStudentCSVField.EndDateAttr) {
		return nil, nil, nil, nil
	}

	if !field.IsPresent(importStudentCSVField.SchoolAttr) {
		if field.IsPresent(importStudentCSVField.SchoolCourseAttr) {
			return nil, nil, nil, entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentSchoolCourseField,
				Index:      index,
				Reason:     entity.NotPresentField,
			}
		}
		if field.IsPresent(importStudentCSVField.StartDateAttr) {
			return nil, nil, nil, entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentSchoolHistoryStartDateField,
				Index:      index,
				Reason:     entity.NotPresentField,
			}
		}
		if field.IsPresent(importStudentCSVField.EndDateAttr) {
			return nil, nil, nil, entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentSchoolHistoryEndDateField,
				Index:      index,
				Reason:     entity.NotPresentField,
			}
		}
	}

	// require
	schoolCSV := importStudentCSVField.SchoolAttr.String()
	schoolPartnerIDs := strings.Split(schoolCSV, constant.ArraySeparatorCSV)
	for _, schoolPartnerID := range schoolPartnerIDs {
		if schoolPartnerID == "" {
			return nil, nil, nil, entity.InvalidFieldError{
				EntityName: entity.StudentEntity,
				FieldName:  entity.StudentSchoolField,
				Index:      index,
				Reason:     entity.Empty,
			}
		}
	}

	// optional course
	schoolCourseCSV := importStudentCSVField.SchoolCourseAttr.String()
	schoolCoursePartnerIDs := strings.Split(schoolCourseCSV, constant.ArraySeparatorCSV)

	// optional start date
	startDateCSV := importStudentCSVField.StartDateAttr.String()
	startDateArr := strings.Split(startDateCSV, constant.ArraySeparatorCSV)

	// optional end date
	endDateCSV := importStudentCSVField.EndDateAttr.String()
	endDateArr := strings.Split(endDateCSV, constant.ArraySeparatorCSV)

	maxLength := len(schoolPartnerIDs)

	if len(schoolCoursePartnerIDs) > maxLength {
		maxLength = len(schoolCoursePartnerIDs)
	}

	if len(startDateArr) > maxLength {
		maxLength = len(startDateArr)
	}

	if len(endDateArr) > maxLength {
		maxLength = len(endDateArr)
	}

	domainSchoolHistories := make(entity.DomainSchoolHistories, maxLength)
	domainSchoolsInfo := make(entity.DomainSchools, maxLength)
	domainSchoolCourses := make(entity.DomainSchoolCourses, maxLength)

	for idx := 0; idx < maxLength; idx++ {
		schoolPartnerID := ""
		if idx < len(schoolPartnerIDs) {
			schoolPartnerID = schoolPartnerIDs[idx]
		}
		domainSchoolsInfo[idx] = entity.SchoolWillBeDelegated{
			HasPartnerInternalID: SchoolInfoImpl{
				SchoolPartnerInternalIDAttr: field.NewString(schoolPartnerID),
			},
		}

		schoolCoursePartnerID := ""
		if idx < len(schoolCoursePartnerIDs) {
			schoolCoursePartnerID = schoolCoursePartnerIDs[idx]
		}
		domainSchoolCourses[idx] = entity.SchoolCourseWillBeDelegated{
			HasPartnerInternalID: SchoolCourseImpl{
				SchoolCoursePartnerInternalIDAttr: field.NewString(schoolCoursePartnerID),
			},
		}

		schoolHistory := &SchoolHistory{}
		startDate := field.NewNullTime()
		if idx < len(startDateArr) {
			if err := startDate.UnmarshalCSV(startDateArr[idx]); err != nil {
				return nil, nil, nil, entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolHistoryStartDateField,
					Index:      index,
					Reason:     entity.FailedUnmarshal,
				}
			}
		}
		schoolHistory.StartDateAttr = startDate

		endDate := field.NewNullTime()
		if idx < len(endDateArr) {
			if err := endDate.UnmarshalCSV(endDateArr[idx]); err != nil {
				return nil, nil, nil, entity.InvalidFieldError{
					EntityName: entity.StudentEntity,
					FieldName:  entity.StudentSchoolHistoryEndDateField,
					Index:      index,
					Reason:     entity.FailedUnmarshal,
				}
			}
		}
		schoolHistory.EndDateAttr = endDate

		domainSchoolHistories[idx] = schoolHistory
	}
	return domainSchoolHistories, domainSchoolsInfo, domainSchoolCourses, nil
}

func (d *DomainStudentService) validateUserID(ctx context.Context, student *StudentCSV, csvIndex int) error {
	zapLogger := ctxzap.Extract(ctx)

	if !field.IsPresent(student.UserID()) {
		return nil
	}
	domainStudents := aggregate.DomainStudents{{DomainStudent: student}}
	if err := d.DomainStudent.ValidateUpdateSystemAndExternalUserID(ctx, domainStudents); err != nil {
		zapLogger.Error(
			"ValidateUpdateSystemAndExternalUserID",
			zap.Error(err),
			zap.String("validateUserID", "DomainStudent.ValidateUpdateSystemAndExternalUserID"),
		)
		e, ok := err.(errcode.Error)
		if !ok {
			e.Code = errcode.InternalError
			e.Err = err
		}

		e.Index = csvIndex
		return e
	}
	return nil
}

func (d *DomainStudentService) fillExistedEmailOfUsers(ctx context.Context, students aggregate.DomainStudents) ([]aggregate.DomainStudent, error) {
	userByUserID, err := d.DomainStudent.GetEmailWithStudentID(ctx, students.StudentIDs())
	if err != nil {
		return nil, errcode.Error{
			Code: errcode.InternalError,
			Err:  fmt.Errorf("failed to get existed emails with user ids: %w", err),
		}
	}

	for idx, student := range students {
		if student.UserID().String() == "" {
			continue
		}

		studentCSV, ok := student.DomainStudent.(*StudentCSV)
		if !ok {
			return nil, errcode.Error{
				Index: idx,
				Code:  errcode.InternalError,
				Err:   errors.New("invalid student type"),
			}
		}

		studentCSV.EmailAttr = userByUserID[studentCSV.UserID().String()].Email()
		studentCSV.LoginEmailAttr = userByUserID[studentCSV.UserID().String()].LoginEmail()
		studentCSV.UserNameAttr = studentCSV.EmailAttr
		students[idx].DomainStudent = studentCSV
	}

	return students, nil
}

func toBirthDay(birthDayStr field.String, csvIndex int) (field.Date, error) {
	birthDay := field.NewNullDate()
	if err := birthDay.UnmarshalCSV(birthDayStr.String()); err != nil {
		return field.NewNullDate(), errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: entity.StudentBirthdayField,
			Index:     csvIndex,
		}
	}
	return birthDay, nil
}

func toBirthDayV2(birthDayStr field.String, index int) (field.Date, error) {
	birthDay := field.NewNullDate()
	if err := birthDay.UnmarshalCSV(birthDayStr.String()); err != nil {
		return field.NewNullDate(), entity.InvalidFieldError{
			FieldName:  entity.StudentBirthdayField,
			EntityName: entity.StudentEntity,
			Index:      index,
			Reason:     entity.FailedUnmarshal,
		}
	}
	return birthDay, nil
}

func ToGender(genderStr field.String, csvIndex int) (field.String, error) {
	if genderStr.String() == "" {
		return field.NewNullString(), nil
	}

	enumGender := field.NewNullInt16()
	if err := enumGender.UnmarshalCSV(genderStr.String()); err != nil {
		return field.NewNullString(), errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: entity.StudentGenderField,
			Err:       err,
			Index:     csvIndex,
		}
	}

	gender, ok := constant.UserGenderMap[int(enumGender.Int16())]
	if !ok {
		return field.NewNullString(), errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: entity.StudentGenderField,
			Index:     csvIndex,
		}
	}

	return field.NewString(gender), nil
}

func ToGenderV2(genderStr field.String, index int) (field.String, error) {
	if genderStr.IsEmpty() {
		return field.NewNullString(), nil
	}

	enumGender := field.NewNullInt16()
	if err := enumGender.UnmarshalCSV(genderStr.String()); err != nil {
		return field.NewNullString(), entity.InvalidFieldError{
			FieldName:  entity.StudentGenderField,
			EntityName: entity.StudentEntity,
			Index:      index,
			Reason:     entity.FailedUnmarshal,
		}
	}

	gender, ok := constant.UserGenderMap[int(enumGender.Int16())]
	if !ok {
		return field.NewNullString(), entity.InvalidFieldError{
			FieldName:  entity.StudentGenderField,
			EntityName: entity.StudentEntity,
			Index:      index,
			Reason:     entity.NotMatchingEnum,
		}
	}

	return field.NewString(gender), nil
}

func toUserPhoneNumbers(student *StudentCSV) entity.DomainUserPhoneNumbers {
	domainUserPhoneNumbers := make(entity.DomainUserPhoneNumbers, 0, 2)

	domainUserPhoneNumbers = append(domainUserPhoneNumbers, &UserPhoneNumber{
		PhoneIDAttr:     field.NewString(idutil.ULIDNow()),
		PhoneTypeAttr:   field.NewString(constant.StudentPhoneNumber),
		PhoneNumberAttr: student.StudentPhoneNumberAttr,
	})

	domainUserPhoneNumbers = append(domainUserPhoneNumbers, &UserPhoneNumber{
		PhoneIDAttr:     field.NewString(idutil.ULIDNow()),
		PhoneTypeAttr:   field.NewString(constant.StudentHomePhoneNumber),
		PhoneNumberAttr: student.StudentHomePhoneNumberAttr,
	})

	return domainUserPhoneNumbers
}

func readAndValidatePayload(payload []byte) error {
	sizeInMB := len(payload) / (1024 * 1024)
	if sizeInMB > 5 {
		return grpc.InvalidPayloadSizeCSVError{
			RequestSize: sizeInMB,
		}
	}
	if len(payload) == 0 {
		return grpc.InvalidPayloadSizeCSVError{
			RequestSize: sizeInMB,
		}
	}
	return nil
}

func ConvertPayloadToImportStudentData(payload []byte) ([]*StudentCSV, error) {
	var studentImportData []*StudentCSV
	// todo: check gocsv.UnmarshalBytes ignore header
	if err := gocsv.UnmarshalBytes(payload, &studentImportData); err != nil {
		return nil, grpc.InternalError{
			RawErr: errors.Wrap(err, "wrong format csv"),
		}
	}
	return studentImportData, nil
}

func isDuplicatedIDs(ids []string) bool {
	mapIDs := map[string]struct{}{}
	for _, ids := range ids {
		mapIDs[ids] = struct{}{}
	}

	return len(ids) != len(mapIDs)
}

func importCsvValidateTag(role string, partnerInternalIDs []string, tags entity.DomainTags) error {
	handleError := func(role string) error {
		err := fmt.Errorf("tag is only for student")
		if role == constant.RoleParent {
			return fmt.Errorf("tag is only for parent")
		}
		return err
	}

	if ok := tags.ContainPartnerInternalIDs(partnerInternalIDs...); !ok {
		return handleError(role)
	}

	for _, tag := range tags {
		if tag.IsArchived().Boolean() {
			return handleError(role)
		}

		if role == constant.RoleParent && !entity.IsParentTag(tag) {
			return handleError(role)
		}

		if role == constant.RoleStudent && !entity.IsStudentTag(tag) {
			return handleError(role)
		}
	}

	return nil
}

func validateCSVEnrollmentStatusFields(location field.String, enrollmentStatus field.String, statusStartDate field.String, csvIndex int) error {
	lenArrayLocations := len(strings.Split(location.String(), constant.ArraySeparatorCSV))

	if field.IsPresent(enrollmentStatus) {
		lenArrayEnrollmentStatuses := len(strings.Split(enrollmentStatus.String(), constant.ArraySeparatorCSV))
		if lenArrayLocations != lenArrayEnrollmentStatuses {
			return errcode.Error{
				Code:      errcode.InvalidData,
				Err:       fmt.Errorf("length of location and enrollment status are not equal"),
				FieldName: entity.StudentFieldEnrollmentStatus,
				Index:     csvIndex,
			}
		}
	}

	if field.IsPresent(statusStartDate) {
		lenArrayStatusStartDates := len(strings.Split(statusStartDate.String(), constant.ArraySeparatorCSV))
		if lenArrayLocations != lenArrayStatusStartDates {
			return errcode.Error{
				Code:      errcode.InvalidData,
				Err:       fmt.Errorf("length of location and status start date are not equal"),
				FieldName: entity.StudentFieldEnrollmentStatusStartDate,
				Index:     csvIndex,
			}
		}
	}

	return nil
}

func mapLocationByPartnerID(locations entity.DomainLocations) map[string]entity.DomainLocation {
	locationByPartnerID := make(map[string]entity.DomainLocation)
	for _, location := range locations {
		locationByPartnerID[location.PartnerInternalID().String()] = location
	}
	return locationByPartnerID
}

func toFullNameAndFullNamePhonetic(student *StudentCSV) (fullName, fullNamePhonetic field.String) {
	firstName := student.FirstNameAttr.String()
	lastName := student.LastNameAttr.String()
	fullName = field.NewString(utils.CombineFirstNameAndLastNameToFullName(firstName, lastName))

	firstNamePhonetic := student.FirstNamePhoneticAttr.String()
	lastNamePhonetic := student.LastNamePhoneticAttr.String()
	fullNamePhonetic = field.NewString(utils.CombineFirstNamePhoneticAndLastNamePhoneticToFullName(firstNamePhonetic, lastNamePhonetic))
	if fullNamePhonetic.String() == "" {
		fullNamePhonetic = field.NewNullString()
	}
	return
}

func toExternalUserID(student *StudentCSV) field.String {
	trimmedExternalUserID := strings.TrimSpace(student.ExternalUserIDAttr.String())

	if trimmedExternalUserID != "" {
		return field.NewString(trimmedExternalUserID)
	}

	return field.NewNullString()
}
