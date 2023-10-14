package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type WorkingHours struct {
	WorkingHoursID pgtype.Text
	Day            pgtype.Text
	OpeningTime    pgtype.Text
	ClosingTime    pgtype.Text
	LocationID     pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (wh *WorkingHours) FieldMap() ([]string, []interface{}) {
	return []string{
			"working_hour_id",
			"day",
			"opening_time",
			"closing_time",
			"location_id",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&wh.WorkingHoursID,
			&wh.Day,
			&wh.OpeningTime,
			&wh.ClosingTime,
			&wh.LocationID,
			&wh.UpdatedAt,
			&wh.CreatedAt,
			&wh.DeletedAt,
		}
}

func (wh *WorkingHours) TableName() string {
	return "working_hour"
}

func (wh *WorkingHours) ToWorkingHoursDomain() *domain.WorkingHours {
	return &domain.WorkingHours{
		Day:         wh.Day.String,
		OpeningTime: wh.OpeningTime.String,
		ClosingTime: wh.ClosingTime.String,
		LocationID:  wh.LocationID.String,
		CreatedAt:   wh.CreatedAt.Time,
		UpdatedAt:   wh.UpdatedAt.Time,
		DeletedAt:   &wh.DeletedAt.Time,
	}
}

func NewWorkingHoursFromEntity(wh *domain.WorkingHours) (*WorkingHours, error) {
	workingHoursDTO := &WorkingHours{}
	database.AllNullEntity(workingHoursDTO)
	if err := multierr.Combine(
		workingHoursDTO.WorkingHoursID.Set(wh.WorkingHoursID),
		workingHoursDTO.Day.Set(wh.Day),
		workingHoursDTO.OpeningTime.Set(wh.OpeningTime),
		workingHoursDTO.ClosingTime.Set(wh.ClosingTime),
		workingHoursDTO.LocationID.Set(wh.LocationID),
		workingHoursDTO.CreatedAt.Set(wh.CreatedAt),
		workingHoursDTO.UpdatedAt.Set(wh.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from working hours entity to working hours dto: %w", err)
	}
	if wh.DeletedAt != nil {
		if err := workingHoursDTO.DeletedAt.Set(wh.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return workingHoursDTO, nil
}
