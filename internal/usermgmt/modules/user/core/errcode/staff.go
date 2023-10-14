package errcode

import "github.com/pkg/errors"

var (
	ErrStaffStartDateIsInvalid         = errors.New("start date is invalid")
	ErrStaffEndDateIsInvalid           = errors.New("end date is invalid")
	ErrStaffStartDateIsLessThanEndDate = errors.New("start date cannot be less than end date")
	ErrStaffWorkingStatusIsEmpty       = errors.New("working status cannot be empty")
)
