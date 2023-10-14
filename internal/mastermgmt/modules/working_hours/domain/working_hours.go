package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type WorkingHours struct {
	WorkingHoursID string
	Day            string
	OpeningTime    string
	ClosingTime    string
	LocationID     string
	UpdatedAt      time.Time
	CreatedAt      time.Time
	DeletedAt      *time.Time
	ResourcePath   string

	Repo WorkingHoursRepo
}

type WorkingHoursList []*WorkingHours

type WorkingHoursBuilder struct {
	workingHours *WorkingHours
}

func NewWorkingHoursBuilder() *WorkingHoursBuilder {
	return &WorkingHoursBuilder{
		workingHours: &WorkingHours{},
	}
}

func (wh *WorkingHoursBuilder) WithWorkingHoursRepo(repo WorkingHoursRepo) *WorkingHoursBuilder {
	wh.workingHours.Repo = repo
	return wh
}

func (wh *WorkingHoursBuilder) WithWorkingHoursID(id string) *WorkingHoursBuilder {
	wh.workingHours.WorkingHoursID = id
	if id == "" {
		wh.workingHours.WorkingHoursID = idutil.ULIDNow()
	}
	return wh
}

func (wh *WorkingHoursBuilder) WithDay(day string) *WorkingHoursBuilder {
	wh.workingHours.Day = day
	return wh
}

func (wh *WorkingHoursBuilder) WithModificationTime(createdAt, updatedAt time.Time) *WorkingHoursBuilder {
	wh.workingHours.CreatedAt = createdAt
	wh.workingHours.UpdatedAt = updatedAt
	return wh
}

func (wh *WorkingHoursBuilder) WithOpeningTime(openingTime string) *WorkingHoursBuilder {
	wh.workingHours.OpeningTime = openingTime
	return wh
}

func (wh *WorkingHoursBuilder) WithClosingTime(closingTime string) *WorkingHoursBuilder {
	wh.workingHours.ClosingTime = closingTime
	return wh
}

func (wh *WorkingHoursBuilder) WithLocationID(locationID string) *WorkingHoursBuilder {
	wh.workingHours.LocationID = locationID
	return wh
}

func isTimeStringFormatValid(timeString string) bool {
	if len(timeString) != 5 {
		return false
	}

	split := strings.Split(timeString, ":")

	if len(split) != 2 {
		return false
	}

	h, errH := strconv.Atoi(split[0])
	m, errM := strconv.Atoi(split[1])

	if errH != nil || errM != nil || h < 0 || h > 23 || m < 0 || m > 59 {
		return false
	}

	return true
}

func (wh *WorkingHours) IsValid() error {
	if len(wh.Day) == 0 {
		return fmt.Errorf("WorkingHours.Day cannot be empty")
	}
	if len(wh.OpeningTime) == 0 {
		return fmt.Errorf("WorkingHours.OpeningTime cannot be empty")
	}
	if !isTimeStringFormatValid(wh.OpeningTime) {
		return fmt.Errorf("WorkingHours.OpeningTime is not valid time format")
	}
	if len(wh.ClosingTime) == 0 {
		return fmt.Errorf("WorkingHours.ClosingTime cannot be empty")
	}
	if !isTimeStringFormatValid(wh.ClosingTime) {
		return fmt.Errorf("WorkingHours.ClosingTime is not valid time format")
	}
	if !utf8.ValidString(wh.Day) {
		return fmt.Errorf("WorkingHours.Day is not valid UTF8 format")
	}
	if wh.CreatedAt.IsZero() {
		return fmt.Errorf("WorkingHours.CreatedAt cannot be empty")
	}
	if wh.UpdatedAt.IsZero() {
		return fmt.Errorf("WorkingHours.UpdatedAt cannot be empty")
	}
	if wh.UpdatedAt.Before(wh.CreatedAt) {
		return fmt.Errorf("WorkingHours.UpdatedAt cannot before WorkingHours.CreatedAt")
	}

	return nil
}

func (wh *WorkingHoursBuilder) BuildWithoutPKCheck() (*WorkingHours, error) {
	if err := wh.workingHours.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid working hours: %w", err)
	}

	return wh.workingHours, nil
}
