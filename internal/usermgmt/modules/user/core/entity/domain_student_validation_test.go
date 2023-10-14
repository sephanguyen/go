package entity

import (
	"testing"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

type studentWithInvalidCurrentGrade struct {
	NullDomainStudent
}

func (studentWithInvalidCurrentGrade) CurrentGrade() field.Int16 {
	return field.NewInt16(-1)
}

type studentWithInvalidEnrollmentStatus struct {
	NullDomainStudent
}

func (studentWithInvalidEnrollmentStatus) EnrollmentStatus() field.String {
	return field.NewString("invalid status")
}

type studentWithInvalidContactPreference struct {
	NullDomainStudent
}

func (studentWithInvalidContactPreference) ContactPreference() field.String {
	return field.NewString("invalid contact")
}

type studentWithValidEnrollment struct {
	NullDomainStudent
}

func (studentWithValidEnrollment) EnrollmentStatus() field.String {
	return field.NewString(StudentEnrollmentStatusPotential)

}

type studentWithNullEnrollmentStatus struct {
	NullDomainStudent
}

func (studentWithNullEnrollmentStatus) UserID() field.String {
	return field.NewString("user-id")

}

func TestDomainStudent_validStudentGrade(t *testing.T) {
	// t.Run("current_grade is empty", func(t *testing.T) {
	// 	t.Parallel()

	// 	err := validStudentGrade(NullDomainStudent{})
	// 	assert.Equal(t, errcode.Error{
	// 		FieldName: StudentFieldCurrentGrade,
	// 		Code:      errcode.MissingMandatory,
	// 	}, err)
	// })

	t.Run("current_grade is invalid", func(t *testing.T) {
		t.Parallel()

		err := validStudentGrade(studentWithInvalidCurrentGrade{})
		assert.Equal(t, errcode.Error{
			FieldName: StudentFieldCurrentGrade,
			Code:      errcode.InvalidData,
		}, err)
	})
}

func TestDomainStudent_validStudentEnrollmentStatus(t *testing.T) {
	t.Run("happy case: valid enrollment when creating", func(t *testing.T) {
		t.Parallel()
		err := validStudentEnrollmentStatus(studentWithValidEnrollment{})
		assert.Nil(t, err)
	})

	t.Run("happy case: null enrollment when updating", func(t *testing.T) {
		t.Parallel()
		err := validStudentEnrollmentStatus(studentWithNullEnrollmentStatus{})
		assert.Nil(t, err)
	})

	t.Run("enrollment_status is empty", func(t *testing.T) {
		t.Parallel()

		err := validStudentEnrollmentStatus(NullDomainStudent{})
		assert.Equal(t, errcode.Error{
			FieldName: StudentFieldEnrollmentStatus,
			Code:      errcode.MissingMandatory,
		}, err)
	})

	t.Run("enrollment_status is invalid", func(t *testing.T) {
		t.Parallel()

		err := validStudentEnrollmentStatus(studentWithInvalidEnrollmentStatus{})
		assert.Equal(t, errcode.Error{
			FieldName: StudentFieldEnrollmentStatus,
			Code:      errcode.InvalidData,
		}, err)
	})
}

func TestDomainStudent_ValidateStudentContactPreference(t *testing.T) {
	t.Run("contact_preference is invalid", func(t *testing.T) {
		t.Parallel()

		err := ValidateStudentContactPreference(studentWithInvalidContactPreference{})
		assert.Equal(t, InvalidFieldError{
			FieldName:  StudentFieldContactPreference,
			EntityName: UserEntity,
			Index:      -1,
			Reason:     NotMatchingEnum,
		}, err)
	})
}
