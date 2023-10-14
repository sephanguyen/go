package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/working_hours/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewWorkingHoursFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		workingHoursEntity := &domain.WorkingHours{
			WorkingHoursID: "working_hour_id",
			Day:            "Monday",
			OpeningTime:    "08:00",
			ClosingTime:    "17:00",
			LocationID:     "location_id",
			UpdatedAt:      now,
			CreatedAt:      now,
		}
		expectedWorkingHours := &WorkingHours{
			WorkingHoursID: database.Text("working_hour_id"),
			Day:            database.Text("Monday"),
			OpeningTime:    database.Text("08:00"),
			ClosingTime:    database.Text("17:00"),
			LocationID:     database.Text("location_id"),
			CreatedAt:      database.Timestamptz(now),
			UpdatedAt:      database.Timestamptz(now),
			DeletedAt:      pgtype.Timestamptz{Status: pgtype.Null},
		}
		gotWorkingHours, err := NewWorkingHoursFromEntity(workingHoursEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedWorkingHours, gotWorkingHours)
	})

}
