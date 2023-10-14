package domain

import (
	"time"
)

type CourseType struct {
	CourseTypeID string
	Name         string
	UpdatedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
}
