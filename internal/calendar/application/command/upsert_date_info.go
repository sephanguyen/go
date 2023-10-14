package command

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/entities"
	"github.com/manabie-com/backend/internal/calendar/domain/valueobj"
	"github.com/manabie-com/backend/internal/calendar/infrastructure"
	"github.com/manabie-com/backend/internal/golibs/database"
)

type UpsertDateInfoCommand struct {
	DB           database.QueryExecer
	DateInfoRepo infrastructure.DateInfoPort
	LocationRepo infrastructure.LocationPort
}

type UpsertDateInfoRequest struct {
	Date        time.Time
	LocationID  string
	DateTypeID  string
	OpeningTime string
	Status      string
	Timezone    string
}

type DuplicateDateInfoRequest struct {
	Date        time.Time
	LocationID  string
	DateTypeID  string
	OpeningTime string
	Status      string
	Timezone    string
	StartDate   time.Time
	EndDate     time.Time
	Frequency   string
}

func (c *UpsertDateInfoCommand) UpsertDateInfo(ctx context.Context, req *UpsertDateInfoRequest) error {
	dateInfo, err := entities.NewDateInfo(req.Date,
		req.LocationID,
		req.DateTypeID,
		req.OpeningTime,
		req.Status,
		req.Timezone,
		c.DB,
		c.DateInfoRepo,
		c.LocationRepo,
	)
	if err != nil {
		return err
	}

	if err := dateInfo.Upsert(ctx); err != nil {
		return err
	}

	return nil
}

func (c *UpsertDateInfoCommand) DuplicateDateInfo(ctx context.Context, req *DuplicateDateInfoRequest) error {
	dateInfo, err := entities.NewDateInfo(req.Date,
		req.LocationID,
		req.DateTypeID,
		req.OpeningTime,
		req.Status,
		req.Timezone,
		c.DB,
		c.DateInfoRepo,
		c.LocationRepo,
	)
	if err != nil {
		return err
	}

	duplicationInfo, err := valueobj.NewDuplicationInfo(
		req.StartDate,
		req.EndDate,
		req.Frequency,
	)
	if err != nil {
		return err
	}

	if err := duplicationInfo.Validate(); err != nil {
		return err
	}

	if err := dateInfo.Duplicate(ctx, duplicationInfo.RetrieveDateOccurrences()); err != nil {
		return err
	}

	return nil
}
