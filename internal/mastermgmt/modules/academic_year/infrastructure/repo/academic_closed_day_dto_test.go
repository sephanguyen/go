package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewAcademicClosedDayFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		academicClosedDayEntity := &domain.AcademicClosedDay{
			AcademicClosedDayID: "academic_closed_day_id",
			Date:                now,
			AcademicYearID:      "academic_year_id",
			AcademicWeekID:      "academic_week_id",
			LocationID:          "location_id",
			UpdatedAt:           now,
			CreatedAt:           now,
		}
		expectedAcademicClosedDay := &AcademicClosedDay{
			AcademicClosedDayID: database.Text("academic_closed_day_id"),
			Date:                pgtype.Date{Time: now, Status: pgtype.Present},
			AcademicYearID:      database.Text("academic_year_id"),
			AcademicWeekID:      database.Text("academic_week_id"),
			LocationID:          database.Text("location_id"),
			CreatedAt:           database.Timestamptz(now),
			UpdatedAt:           database.Timestamptz(now),
			DeletedAt:           pgtype.Timestamptz{Status: pgtype.Null},
		}
		gotAcademicClosedDay, err := NewAcademicClosedDayFromEntity(academicClosedDayEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedAcademicClosedDay, gotAcademicClosedDay)
	})

}
