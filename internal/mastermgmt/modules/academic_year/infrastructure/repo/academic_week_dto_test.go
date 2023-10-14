package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewAcademicWeekFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		academicWeekEntity := &domain.AcademicWeek{
			AcademicWeekID: "academic_week_id",
			WeekOrder:      1,
			Name:           "Week 1",
			StartDate:      now,
			EndDate:        now.Add(24 * 7 * time.Hour),
			Period:         "Term 1",
			AcademicYearID: "academic_year_id",
			LocationID:     "location_id",
			UpdatedAt:      now,
			CreatedAt:      now,
		}
		expectedAcademicWeek := &AcademicWeek{
			AcademicWeekID: database.Text("academic_week_id"),
			WeekOrder:      database.Int2(1),
			Name:           database.Text("Week 1"),
			StartDate:      pgtype.Date{Time: now, Status: pgtype.Present},
			EndDate:        pgtype.Date{Time: now.Add(24 * 7 * time.Hour), Status: pgtype.Present},
			Period:         database.Text("Term 1"),
			AcademicYearID: database.Text("academic_year_id"),
			LocationID:     database.Text("location_id"),
			CreatedAt:      database.Timestamptz(now),
			UpdatedAt:      database.Timestamptz(now),
			DeletedAt:      pgtype.Timestamptz{Status: pgtype.Null},
		}
		gotAcademicWeek, err := NewAcademicWeekFromEntity(academicWeekEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedAcademicWeek, gotAcademicWeek)
	})

}
