package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type TimeSlot struct {
	TimeSlotID         pgtype.Text
	TimeSlotInternalID pgtype.Text
	StartTime          pgtype.Text
	EndTime            pgtype.Text
	LocationID         pgtype.Text
	UpdatedAt          pgtype.Timestamptz
	CreatedAt          pgtype.Timestamptz
	DeletedAt          pgtype.Timestamptz
}

func (ts *TimeSlot) FieldMap() ([]string, []interface{}) {
	return []string{
			"time_slot_id",
			"time_slot_internal_id",
			"start_time",
			"end_time",
			"location_id",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&ts.TimeSlotID,
			&ts.TimeSlotInternalID,
			&ts.StartTime,
			&ts.EndTime,
			&ts.LocationID,
			&ts.UpdatedAt,
			&ts.CreatedAt,
			&ts.DeletedAt,
		}
}

func (ts *TimeSlot) TableName() string {
	return "time_slot"
}

func (ts *TimeSlot) ToTimeSlotDomain() *domain.TimeSlot {
	return &domain.TimeSlot{
		TimeSlotID:         ts.TimeSlotID.String,
		TimeSlotInternalID: ts.TimeSlotInternalID.String,
		StartTime:          ts.StartTime.String,
		EndTime:            ts.EndTime.String,
		LocationID:         ts.LocationID.String,
		CreatedAt:          ts.CreatedAt.Time,
		UpdatedAt:          ts.UpdatedAt.Time,
		DeletedAt:          &ts.DeletedAt.Time,
	}
}

func NewTimeSlotFromEntity(ts *domain.TimeSlot) (*TimeSlot, error) {
	TimeSlotDTO := &TimeSlot{}
	database.AllNullEntity(TimeSlotDTO)
	if err := multierr.Combine(
		TimeSlotDTO.TimeSlotID.Set(ts.TimeSlotID),
		TimeSlotDTO.TimeSlotInternalID.Set(ts.TimeSlotInternalID),
		TimeSlotDTO.StartTime.Set(ts.StartTime),
		TimeSlotDTO.EndTime.Set(ts.EndTime),
		TimeSlotDTO.LocationID.Set(ts.LocationID),
		TimeSlotDTO.CreatedAt.Set(ts.CreatedAt),
		TimeSlotDTO.UpdatedAt.Set(ts.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from time slot entity to time slot dto: %w", err)
	}
	if ts.DeletedAt != nil {
		if err := TimeSlotDTO.DeletedAt.Set(ts.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return TimeSlotDTO, nil
}
