package domain

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
)

type TimeSlot struct {
	TimeSlotID         string
	TimeSlotInternalID string
	StartTime          string
	EndTime            string
	LocationID         string
	UpdatedAt          time.Time
	CreatedAt          time.Time
	DeletedAt          *time.Time
	ResourcePath       string
	Repo               TimeSlotRepo
}

type TimeSlots []*TimeSlot

type TimeSlotBuilder struct {
	timeSlot *TimeSlot
}

func NewTimeSlotBuilder() *TimeSlotBuilder {
	return &TimeSlotBuilder{
		timeSlot: &TimeSlot{},
	}
}

func (ts *TimeSlot) IsValid() error {
	if len(ts.TimeSlotInternalID) == 0 {
		return fmt.Errorf("TimeSlot.TimeSlotInternalID cannot be empty")
	}

	if len(ts.StartTime) == 0 {
		return fmt.Errorf("TimeSlot.StartTime cannot be empty")
	}

	if !isTimeStringFormatValid(ts.StartTime) {
		return fmt.Errorf("TimeSlot.StartTime is not valid time format")
	}

	if len(ts.EndTime) == 0 {
		return fmt.Errorf("TimeSlot.EndTime cannot be empty")
	}

	if !isTimeStringFormatValid(ts.EndTime) {
		return fmt.Errorf("TimeSlot.EndTime is not valid time format")
	}

	if ts.CreatedAt.IsZero() {
		return fmt.Errorf("TimeSlot.CreatedAt cannot be empty")
	}

	if ts.UpdatedAt.IsZero() {
		return fmt.Errorf("TimeSlot.UpdatedAt cannot be empty")
	}

	if ts.UpdatedAt.Before(ts.CreatedAt) {
		return fmt.Errorf("TimeSlot.UpdatedAt cannot before TimeSlot.CreatedAt")
	}

	return nil
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

func (ts *TimeSlotBuilder) BuildWithoutPKCheck() (*TimeSlot, error) {
	if err := ts.timeSlot.IsValid(); err != nil {
		return nil, fmt.Errorf("invalid time slot: %w", err)
	}

	return ts.timeSlot, nil
}

func (ts *TimeSlotBuilder) WithTimeSlotRepo(repo TimeSlotRepo) *TimeSlotBuilder {
	ts.timeSlot.Repo = repo
	return ts
}

func (ts *TimeSlotBuilder) WithTimeSlotID(id string) *TimeSlotBuilder {
	ts.timeSlot.TimeSlotID = id
	if id == "" {
		ts.timeSlot.TimeSlotID = idutil.ULIDNow()
	}
	return ts
}

func (ts *TimeSlotBuilder) WithTimeSlotInternalID(id string) *TimeSlotBuilder {
	ts.timeSlot.TimeSlotInternalID = id
	return ts
}

func (ts *TimeSlotBuilder) WithStartTime(startTime string) *TimeSlotBuilder {
	ts.timeSlot.StartTime = startTime
	return ts
}

func (ts *TimeSlotBuilder) WithEndTime(endTime string) *TimeSlotBuilder {
	ts.timeSlot.EndTime = endTime
	return ts
}

func (ts *TimeSlotBuilder) WithLocationID(locationID string) *TimeSlotBuilder {
	ts.timeSlot.LocationID = locationID
	return ts
}

func (ts *TimeSlotBuilder) WithModificationTime(createdAt, updatedAt time.Time) *TimeSlotBuilder {
	ts.timeSlot.CreatedAt = createdAt
	ts.timeSlot.UpdatedAt = updatedAt
	return ts
}

func (ts *TimeSlotBuilder) WithDeletedTime(deletedAt *time.Time) *TimeSlotBuilder {
	ts.timeSlot.DeletedAt = deletedAt
	return ts
}

func (ts *TimeSlotBuilder) GetTimeSlot() *TimeSlot {
	return ts.timeSlot
}
