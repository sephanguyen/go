package domain

import "time"

type AcademicYear struct {
	AcademicYearID string
	Name           string
	StartDate      time.Time
	EndDate        time.Time
	UpdatedAt      time.Time
	CreatedAt      time.Time
	DeletedAt      *time.Time
	ResourcePath   string
	Repo           AcademicYearRepo
}

type AcademicYears []*AcademicYear
