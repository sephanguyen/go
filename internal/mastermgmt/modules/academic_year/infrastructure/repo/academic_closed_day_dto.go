package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type AcademicClosedDay struct {
	AcademicClosedDayID pgtype.Text
	Date                pgtype.Date
	AcademicWeekID      pgtype.Text
	AcademicYearID      pgtype.Text
	LocationID          pgtype.Text
	UpdatedAt           pgtype.Timestamptz
	CreatedAt           pgtype.Timestamptz
	DeletedAt           pgtype.Timestamptz
}

func (acd *AcademicClosedDay) FieldMap() ([]string, []interface{}) {
	return []string{
			"academic_closed_day_id",
			"date",
			"academic_week_id",
			"academic_year_id",
			"location_id",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&acd.AcademicClosedDayID,
			&acd.Date,
			&acd.AcademicWeekID,
			&acd.AcademicYearID,
			&acd.LocationID,
			&acd.UpdatedAt,
			&acd.CreatedAt,
			&acd.DeletedAt,
		}
}

func (acd *AcademicClosedDay) TableName() string {
	return "academic_closed_day"
}

func (acd *AcademicClosedDay) ToAcademicClosedDayDomain() *domain.AcademicClosedDay {
	return &domain.AcademicClosedDay{
		AcademicClosedDayID: acd.AcademicClosedDayID.String,
		Date:                acd.Date.Time,
		AcademicWeekID:      acd.AcademicWeekID.String,
		AcademicYearID:      acd.AcademicYearID.String,
		CreatedAt:           acd.CreatedAt.Time,
		UpdatedAt:           acd.UpdatedAt.Time,
		DeletedAt:           &acd.DeletedAt.Time,
	}
}

func NewAcademicClosedDayFromEntity(acd *domain.AcademicClosedDay) (*AcademicClosedDay, error) {
	academicClosedDayDTO := &AcademicClosedDay{}
	database.AllNullEntity(academicClosedDayDTO)
	if err := multierr.Combine(
		academicClosedDayDTO.AcademicClosedDayID.Set(acd.AcademicClosedDayID),
		academicClosedDayDTO.Date.Set(acd.Date),
		academicClosedDayDTO.AcademicWeekID.Set(acd.AcademicWeekID),
		academicClosedDayDTO.AcademicYearID.Set(acd.AcademicYearID),
		academicClosedDayDTO.LocationID.Set(acd.LocationID),
		academicClosedDayDTO.CreatedAt.Set(acd.CreatedAt),
		academicClosedDayDTO.UpdatedAt.Set(acd.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from academic year entity to academic year dto: %w", err)
	}
	if acd.DeletedAt != nil {
		if err := academicClosedDayDTO.DeletedAt.Set(acd.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return academicClosedDayDTO, nil
}
