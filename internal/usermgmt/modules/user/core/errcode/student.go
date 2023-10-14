package errcode

import (
	"github.com/pkg/errors"
)

var (
	ErrStudentGradeIsEmpty              = errors.New("student grade is empty")
	ErrStudentGradeIsInvalid            = errors.New("student grade is invalid")
	ErrStudentEnrollmentStatusIsEmpty   = errors.New("student enrollment status is empty")
	ErrStudentEnrollmentStatusIsInvalid = errors.New("student enrollment status is invalid")
	ErrStudentOnlyLocationCenter        = errors.New("student locations must be the lowest locations")
)
