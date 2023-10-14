package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/errcode"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/errorx"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

func ValidStudentEnrollmentStatusHistory(studentEnrollmentStatusHistory DomainEnrollmentStatusHistory) error {
	return errorx.ReturnFirstErr(
		validEnrollmentStatus(studentEnrollmentStatusHistory),
		validStartDateAndEndDate(studentEnrollmentStatusHistory),
	)
}

func validEnrollmentStatus(studentEnrollmentStatus DomainEnrollmentStatusHistory) error {
	userErr := errcode.Error{
		FieldName: StudentFieldEnrollmentStatus,
		Code:      errcode.InvalidData,
	}
	switch studentEnrollmentStatus.EnrollmentStatus().Status() {
	case field.StatusUndefined, field.StatusNull:
		userErr.Code = errcode.InvalidData
	}
	switch studentEnrollmentStatus.EnrollmentStatus().String() {
	case StudentEnrollmentStatusEnrolled,
		StudentEnrollmentStatusWithdrawn,
		StudentEnrollmentStatusGraduated,
		StudentEnrollmentStatusTemporary,
		StudentEnrollmentStatusLOA:
		return nil
	}
	return userErr
}

func validStartDateAndEndDate(studentEnrollmentStatus DomainEnrollmentStatusHistory) error {
	startDate := studentEnrollmentStatus.StartDate().Ptr()
	endDate := studentEnrollmentStatus.EndDate().Ptr()
	hadDate := !startDate.IsZero() && !endDate.IsZero()

	if hadDate && startDate.After(endDate) {
		return errcode.Error{
			FieldName: StartDateFieldEnrollmentStatusHistory,
			Code:      errcode.InvalidData,
		}
	}
	return nil
}
