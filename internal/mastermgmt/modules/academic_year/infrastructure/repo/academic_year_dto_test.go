package repo

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewAcademicYearFromEntity(t *testing.T) {
	now := time.Now()
	t.Run("success", func(t *testing.T) {
		academicYearEntity := &domain.AcademicYear{
			AcademicYearID: "academic_year_id",
			Name:           "2023",
			StartDate:      now,
			EndDate:        now.Add(24 * 7 * time.Hour),
			UpdatedAt:      now,
			CreatedAt:      now,
		}
		expectedAcademicYear := &AcademicYear{
			AcademicYearID: database.Text("academic_year_id"),
			Name:           database.Text("2023"),
			StartDate:      pgtype.Date{Time: now, Status: pgtype.Present},
			EndDate:        pgtype.Date{Time: now.Add(24 * 7 * time.Hour), Status: pgtype.Present},
			CreatedAt:      database.Timestamptz(now),
			UpdatedAt:      database.Timestamptz(now),
			DeletedAt:      pgtype.Timestamptz{Status: pgtype.Null},
		}
		gotAcademicYear, err := NewAcademicYearFromEntity(academicYearEntity)
		assert.NoError(t, err)
		assert.EqualValues(t, expectedAcademicYear, gotAcademicYear)
	})

}
