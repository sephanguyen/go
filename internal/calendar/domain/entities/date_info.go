package entities

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type DateInfo struct {
	Date        time.Time
	LocationID  string
	DateTypeID  constants.DateTypeID
	OpeningTime string
	Status      constants.DateInfoStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
	TimeZone    string

	DB           database.QueryExecer
	DateInfoRepo infrastructure.DateInfoPort
	LocationRepo infrastructure.LocationPort
}

func NewDateInfo(date time.Time,
	locationID, dateType, openingTime, status, timezone string,
	db database.QueryExecer,
	dateInfoRepo infrastructure.DateInfoPort,
	locationRepo infrastructure.LocationPort) (*DateInfo, error) {
	now := time.Now()

	dateInfo := &DateInfo{
		Date:         date,
		LocationID:   locationID,
		OpeningTime:  openingTime,
		TimeZone:     timezone,
		CreatedAt:    now,
		UpdatedAt:    now,
		DB:           db,
		DateInfoRepo: dateInfoRepo,
		LocationRepo: locationRepo,
	}

	if len(dateType) > 0 {
		dateTypeID, err := GetDateTypeID(dateType)
		if err != nil {
			return nil, err
		}
		dateInfo.DateTypeID = dateTypeID
	}

	if len(status) > 0 {
		dateInfoStatus, err := GetDateInfoStatus(status)
		if err != nil {
			return nil, err
		}
		dateInfo.Status = dateInfoStatus
	}

	return dateInfo, nil
}

func GetDateInfoStatus(status string) (constants.DateInfoStatus, error) {
	switch strings.ToLower(status) {
	case "none":
		return constants.None, nil
	case "draft":
		return constants.Draft, nil
	case "published":
		return constants.Published, nil
	}

	return "", errors.New("unsupported date info status")
}

func (d *DateInfo) Validate(ctx context.Context) error {
	if d.Date.IsZero() {
		return fmt.Errorf("date cannot be empty")
	}

	if len(d.LocationID) == 0 {
		return fmt.Errorf("location id cannot be empty")
	}

	if d.DateTypeID == "closed" && len(d.OpeningTime) != 0 {
		return fmt.Errorf("when date type is 'closed', opening time must be empty")
	}

	if _, err := d.LocationRepo.GetLocationByID(ctx, d.DB, d.LocationID); err != nil {
		return fmt.Errorf("failed to get location id %s in database: %w", d.LocationID, err)
	}

	return nil
}

func (d *DateInfo) Upsert(ctx context.Context) error {
	if err := d.Validate(ctx); err != nil {
		return err
	}

	return d.DateInfoRepo.UpsertDateInfo(ctx, d.DB, &dto.UpsertDateInfoParams{
		DateInfo: &dto.DateInfo{
			Date:        d.Date,
			LocationID:  d.LocationID,
			DateTypeID:  string(d.DateTypeID),
			OpeningTime: d.OpeningTime,
			Status:      string(d.Status),
			TimeZone:    d.TimeZone,
		},
	})
}

func (d *DateInfo) Duplicate(ctx context.Context, dates []time.Time) error {
	if err := d.Validate(ctx); err != nil {
		return err
	}

	storedInfo, err := d.DateInfoRepo.GetDateInfoByDateAndLocationID(ctx, d.DB, d.Date, d.LocationID)
	if err != nil {
		return err
	}

	return d.DateInfoRepo.DuplicateDateInfo(ctx, d.DB, &dto.DuplicateDateInfoParams{
		DateInfo: &dto.DateInfo{
			Date:        storedInfo.Date,
			LocationID:  storedInfo.LocationID,
			DateTypeID:  storedInfo.DateTypeID,
			OpeningTime: storedInfo.OpeningTime,
			Status:      storedInfo.Status,
			TimeZone:    storedInfo.TimeZone,
		},
		Dates: dates,
	})
}
