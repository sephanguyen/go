package domain

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type AcademicWeek struct {
	AcademicWeekID string
	WeekOrder      int
	Name           string
	StartDate      time.Time
	EndDate        time.Time
	AcademicYearID string
	Period         string
	LocationID     string
	UpdatedAt      time.Time
	CreatedAt      time.Time
	DeletedAt      *time.Time
	ResourcePath   string
	Repo           AcademicWeekRepo
}

type AcademicWeeks []*AcademicWeek

type AcademicWeekBuilder struct {
	academicWeek *AcademicWeek
}

func NewAcademicWeekBuilder() *AcademicWeekBuilder {
	return &AcademicWeekBuilder{
		academicWeek: &AcademicWeek{},
	}
}

func (aw *AcademicWeek) IsValid() error {
	if len(aw.Name) == 0 {
		return fmt.Errorf("AcademicWeek.Name cannot be empty")
	}

	if len(aw.Period) == 0 {
		return fmt.Errorf("AcademicWeek.Period cannot be empty")
	}

	if !utf8.ValidString(aw.Name) {
		return fmt.Errorf("AcademicWeek.Name is not valid UTF8 format")
	}

	if len(aw.AcademicYearID) == 0 {
		return fmt.Errorf("AcademicWeek.AcademicYearID cannot be empty")
	}
	if aw.StartDate.IsZero() {
		return fmt.Errorf("AcademicWeek.StartDate cannot be empty")
	}
	if aw.EndDate.IsZero() {
		return fmt.Errorf("AcademicWeek.EndDate cannot be empty")
	}
	if aw.CreatedAt.IsZero() {
		return fmt.Errorf("AcademicWeek.CreatedAt cannot be empty")
	}
	if aw.UpdatedAt.IsZero() {
		return fmt.Errorf("AcademicWeek.UpdatedAt cannot be empty")
	}
	if aw.EndDate.Before(aw.StartDate) {
		return fmt.Errorf("AcademicWeek.EndDate cannot before AcademicWeek.StartDate")
	}
	if aw.UpdatedAt.Before(aw.CreatedAt) {
		return fmt.Errorf("AcademicWeek.UpdatedAt cannot before AcademicWeek.CreatedAt")
	}

	return nil
}

func (aw *AcademicWeekBuilder) Build() (*AcademicWeek, error) {
	if err := aw.academicWeek.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid academic week: %w", err)
	}

	if aw.academicWeek.AcademicWeekID == "" {
		aw.academicWeek.AcademicWeekID = idutil.ULIDNow()
	}

	return aw.academicWeek, nil
}

func (aw *AcademicWeekBuilder) WithAcademicWeekRepo(repo AcademicWeekRepo) *AcademicWeekBuilder {
	aw.academicWeek.Repo = repo
	return aw
}

func (aw *AcademicWeekBuilder) WithAcademicWeekID(id string) *AcademicWeekBuilder {
	aw.academicWeek.AcademicWeekID = id
	if id == "" {
		aw.academicWeek.AcademicWeekID = idutil.ULIDNow()
	}
	return aw
}

func (aw *AcademicWeekBuilder) WithAcademicYearID(id string) *AcademicWeekBuilder {
	aw.academicWeek.AcademicYearID = id
	return aw
}

func (aw *AcademicWeekBuilder) WithWeekOrder(weekOrder int) *AcademicWeekBuilder {
	aw.academicWeek.WeekOrder = weekOrder
	return aw
}

func (aw *AcademicWeekBuilder) WithName(week string) *AcademicWeekBuilder {
	aw.academicWeek.Name = week
	return aw
}

func (aw *AcademicWeekBuilder) WithStartDate(startDate time.Time) *AcademicWeekBuilder {
	aw.academicWeek.StartDate = startDate
	return aw
}

func (aw *AcademicWeekBuilder) WithEndDate(endDate time.Time) *AcademicWeekBuilder {
	aw.academicWeek.EndDate = endDate
	return aw
}

func (aw *AcademicWeekBuilder) WithPeriod(period string) *AcademicWeekBuilder {
	aw.academicWeek.Period = period
	return aw
}

func (aw *AcademicWeekBuilder) WithLocationID(locationID string) *AcademicWeekBuilder {
	aw.academicWeek.LocationID = locationID
	return aw
}

func (aw *AcademicWeekBuilder) WithModificationTime(createdAt, updatedAt time.Time) *AcademicWeekBuilder {
	aw.academicWeek.CreatedAt = createdAt
	aw.academicWeek.UpdatedAt = updatedAt
	return aw
}

func (aw *AcademicWeekBuilder) WithDeletedTime(deletedAt *time.Time) *AcademicWeekBuilder {
	aw.academicWeek.DeletedAt = deletedAt
	return aw
}

func (aw *AcademicWeekBuilder) GetAcademicWeek() *AcademicWeek {
	return aw.academicWeek
}
