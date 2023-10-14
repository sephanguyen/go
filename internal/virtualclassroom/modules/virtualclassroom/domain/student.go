package domain

import (
	"errors"
	"time"
)

var (
	ErrStudentNotFound = errors.New("student not found")
)

type Student struct {
	StudentID         string
	CurrentGrade      int
	GradeID           string
	StudentExternalID string
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         *time.Time
}
