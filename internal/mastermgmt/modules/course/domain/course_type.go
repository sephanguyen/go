package domain

import (
	"time"
)

type CourseType struct {
	CourseTypeID string
	Name         string
	IsArchived   bool
	Remarks      string
	UpdatedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
}
