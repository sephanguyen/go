package entity

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

type enrollmentStatusHistoryWithInvalidEnrollmentStatus struct {
	DefaultDomainEnrollmentStatusHistory
}

func (enrollmentStatusHistoryWithInvalidEnrollmentStatus) EnrollmentStatus() field.String {
	return field.NewString("invalid status")
}

func TestEnrollmentStatusHistory_validStudentEnrollmentStatus(t *testing.T) {
	t.Run("enrollment_status is empty", func(t *testing.T) {
		t.Parallel()

		err := validEnrollmentStatus(DefaultDomainEnrollmentStatusHistory{})
		assert.Equal(t, errcode.Error{
			FieldName: StudentFieldEnrollmentStatus,
			Code:      errcode.InvalidData,
		}, err)
	})

	t.Run("enrollment_status is invalid", func(t *testing.T) {
		t.Parallel()

		err := validEnrollmentStatus(enrollmentStatusHistoryWithInvalidEnrollmentStatus{})
		assert.Equal(t, errcode.Error{
			FieldName: StudentFieldEnrollmentStatus,
			Code:      errcode.InvalidData,
		}, err)
	})
}

type enrollmentStatusHistoryWithInvalidStartDateAndEndDate struct {
	DefaultDomainEnrollmentStatusHistory
}

func (enrollmentStatusHistoryWithInvalidStartDateAndEndDate) StartDate() field.Time {
	return field.NewTime(time.Now().Add(48 * time.Hour))
}

func (enrollmentStatusHistoryWithInvalidStartDateAndEndDate) EndDate() field.Time {
	return field.NewTime(time.Now())
}

func TestEnrollmentStatusHistory_validStartDateAndEndDate(t *testing.T) {
	t.Run("start date is invalid", func(t *testing.T) {
		t.Parallel()

		err := validStartDateAndEndDate(enrollmentStatusHistoryWithInvalidStartDateAndEndDate{})
		assert.Equal(t, errcode.Error{
			FieldName: StartDateFieldEnrollmentStatusHistory,
			Code:      errcode.InvalidData,
		}, err)
	})
}
