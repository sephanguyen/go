package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/time_slot/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewTimeSlotFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		timeSlotEntity := &domain.TimeSlot{
			TimeSlotID:         "time_slot_01",
			TimeSlotInternalID: "1",
			StartTime:          "11:00",
			EndTime:            "13:00",
			LocationID:         "location_id",
			UpdatedAt:          now,
			CreatedAt:          now,
		}
		expectedTimeSlot := &TimeSlot{
			TimeSlotID:         database.Text("time_slot_01"),
			TimeSlotInternalID: database.Text("1"),
			StartTime:          database.Text("11:00"),
			EndTime:            database.Text("13:00"),
			LocationID:         database.Text("location_id"),
			CreatedAt:          database.Timestamptz(now),
			UpdatedAt:          database.Timestamptz(now),
			DeletedAt:          pgtype.Timestamptz{Status: pgtype.Null},
		}
		gotTimeSlot, err := NewTimeSlotFromEntity(timeSlotEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedTimeSlot, gotTimeSlot)
	})

}
