package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

// ValidStudent func following business logic to validate a school admin
// Ref for product specs: <https://product-specs.com>
// Returns domain error if there are any violations
func ValidStudent(student DomainStudent) error {
	return errorx.ReturnFirstErr(
		validStudentGrade(student),
		validStudentEnrollmentStatus(student),
		ValidateStudentContactPreference(student),
	)
}

func validStudentGrade(student DomainStudent) error {
	userErr := errcode.Error{
		FieldName: StudentFieldCurrentGrade,
	}
	if student.CurrentGrade().Int16() < 0 || student.CurrentGrade().Int16() > 16 {
		userErr.Code = errcode.InvalidData
		return userErr
	}
	return nil
}

func validStudentEnrollmentStatus(student DomainStudent) error {
	userErr := errcode.Error{
		FieldName: StudentFieldEnrollmentStatus,
	}
	switch student.EnrollmentStatus().Status() {
	case field.StatusUndefined, field.StatusNull:
		// Skip when updating student with empty EnrollmentStatus
		if student.UserID().String() != "" {
			return nil
		}
		userErr.Code = errcode.MissingMandatory
		return userErr
	}
	switch student.EnrollmentStatus().String() {
	case StudentEnrollmentStatusPotential,
		StudentEnrollmentStatusEnrolled,
		StudentEnrollmentStatusWithdrawn,
		StudentEnrollmentStatusGraduated,
		StudentEnrollmentStatusTemporary,
		StudentEnrollmentStatusNonPotential,
		StudentEnrollmentStatusLOA:
		return nil
	default:
		return errcode.Error{
			Code:      errcode.InvalidData,
			FieldName: StudentFieldEnrollmentStatus,
		}
	}
}

func ValidateStudentContactPreference(student DomainStudent) error {
	index := GetIndex(student)
	if !field.IsPresent(student.ContactPreference()) {
		return nil
	}

	switch student.ContactPreference().String() {
	case UserPhoneNumberTypeStudentPhoneNumber,
		UserPhoneNumberTypeStudentHomePhoneNumber,
		UserPhoneNumberTypeParentPrimaryPhoneNumber,
		UserPhoneNumberTypeParentSecondaryPhoneNumber:
		return nil
	default:
		return InvalidFieldError{
			FieldName:  StudentFieldContactPreference,
			EntityName: UserEntity,
			Index:      index,
			Reason:     NotMatchingEnum,
		}
	}
}
