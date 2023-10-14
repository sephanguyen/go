package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type AcademicClosedDay struct {
	AcademicClosedDayID string
	Date                time.Time
	AcademicWeekID      string
	AcademicYearID      string
	LocationID          string
	UpdatedAt           time.Time
	CreatedAt           time.Time
	DeletedAt           *time.Time
	ResourcePath        string
	Repo                AcademicClosedDayRepo
}

type AcademicClosedDays []*AcademicClosedDay

type AcademicClosedDayBuilder struct {
	academicClosedDay *AcademicClosedDay
}

func NewAcademicClosedDayBuilder() *AcademicClosedDayBuilder {
	return &AcademicClosedDayBuilder{
		academicClosedDay: &AcademicClosedDay{},
	}
}

func (acd *AcademicClosedDayBuilder) WithAcademicClosedDayRepo(repo AcademicClosedDayRepo) *AcademicClosedDayBuilder {
	acd.academicClosedDay.Repo = repo
	return acd
}

func (acd *AcademicClosedDay) IsValid() error {
	if acd.Date.IsZero() {
		return fmt.Errorf("AcademicClosedDay.Date cannot be empty")
	}
	if len(acd.AcademicYearID) == 0 {
		return fmt.Errorf("AcademicClosedDay.AcademicYearID cannot be empty")
	}
	if acd.Date.IsZero() {
		return fmt.Errorf("AcademicClosedDay.EndDate cannot be empty")
	}
	if acd.CreatedAt.IsZero() {
		return fmt.Errorf("AcademicClosedDay.CreatedAt cannot be empty")
	}
	if acd.UpdatedAt.IsZero() {
		return fmt.Errorf("AcademicClosedDay.UpdatedAt cannot be empty")
	}
	if acd.UpdatedAt.Before(acd.CreatedAt) {
		return fmt.Errorf("AcademicClosedDay.UpdatedAt cannot before AcademicClosedDay.CreatedAt")
	}
	return nil
}

func (acd *AcademicClosedDayBuilder) Build(isRequiredWeek bool) (*AcademicClosedDay, error) {
	if err := acd.academicClosedDay.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid closed day: %w", err)
	}

	if isRequiredWeek && acd.academicClosedDay.AcademicWeekID == "" {
		return nil, fmt.Errorf("missing academic_week_id")
	}

	if acd.academicClosedDay.AcademicClosedDayID == "" {
		acd.academicClosedDay.AcademicClosedDayID = idutil.ULIDNow()
	}
	return acd.academicClosedDay, nil
}

func (acd *AcademicClosedDayBuilder) WithAcademicClosedDayID(id string) *AcademicClosedDayBuilder {
	acd.academicClosedDay.AcademicClosedDayID = id
	if id == "" {
		acd.academicClosedDay.AcademicClosedDayID = idutil.ULIDNow()
	}
	return acd
}

func (acd *AcademicClosedDayBuilder) WithDate(date time.Time) *AcademicClosedDayBuilder {
	acd.academicClosedDay.Date = date
	return acd
}

func (acd *AcademicClosedDayBuilder) WithAcademicYearID(id string) *AcademicClosedDayBuilder {
	acd.academicClosedDay.AcademicYearID = id
	return acd
}

func (acd *AcademicClosedDayBuilder) WithAcademicWeekID(id string) *AcademicClosedDayBuilder {
	acd.academicClosedDay.AcademicWeekID = id
	return acd
}

func (acd *AcademicClosedDayBuilder) WithLocationID(locationID string) *AcademicClosedDayBuilder {
	acd.academicClosedDay.LocationID = locationID
	return acd
}

func (acd *AcademicClosedDayBuilder) WithModificationTime(createdAt, updatedAt time.Time) *AcademicClosedDayBuilder {
	acd.academicClosedDay.CreatedAt = createdAt
	acd.academicClosedDay.UpdatedAt = updatedAt
	return acd
}

func (acd *AcademicClosedDayBuilder) WithDeletedTime(deletedAt *time.Time) *AcademicClosedDayBuilder {
	acd.academicClosedDay.DeletedAt = deletedAt
	return acd
}

func (acd *AcademicClosedDayBuilder) GetAcademicClosedDay() *AcademicClosedDay {
	return acd.academicClosedDay
}
