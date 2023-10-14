package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type AcademicWeek struct {
	AcademicWeekID pgtype.Text
	WeekOrder      pgtype.Int2
	Name           pgtype.Text
	StartDate      pgtype.Date
	EndDate        pgtype.Date
	Period         pgtype.Text
	AcademicYearID pgtype.Text
	LocationID     pgtype.Text
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (w *AcademicWeek) FieldMap() ([]string, []interface{}) {
	return []string{
			"academic_week_id",
			"week_order",
			"name",
			"start_date",
			"end_date",
			"period",
			"academic_year_id",
			"location_id",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&w.AcademicWeekID,
			&w.WeekOrder,
			&w.Name,
			&w.StartDate,
			&w.EndDate,
			&w.Period,
			&w.AcademicYearID,
			&w.LocationID,
			&w.UpdatedAt,
			&w.CreatedAt,
			&w.DeletedAt,
		}
}

func (w *AcademicWeek) TableName() string {
	return "academic_week"
}

func (w *AcademicWeek) ToAcademicWeekDomain() *domain.AcademicWeek {
	return &domain.AcademicWeek{
		AcademicWeekID: w.AcademicWeekID.String,
		WeekOrder:      int(w.WeekOrder.Int),
		Name:           w.Name.String,
		StartDate:      w.StartDate.Time,
		EndDate:        w.EndDate.Time,
		Period:         w.Period.String,
		AcademicYearID: w.AcademicYearID.String,
		CreatedAt:      w.CreatedAt.Time,
		UpdatedAt:      w.UpdatedAt.Time,
		DeletedAt:      &w.DeletedAt.Time,
	}
}

func NewAcademicWeekFromEntity(w *domain.AcademicWeek) (*AcademicWeek, error) {
	academicWeekDTO := &AcademicWeek{}
	database.AllNullEntity(academicWeekDTO)
	if err := multierr.Combine(
		academicWeekDTO.AcademicWeekID.Set(w.AcademicWeekID),
		academicWeekDTO.WeekOrder.Set(w.WeekOrder),
		academicWeekDTO.Name.Set(w.Name),
		academicWeekDTO.StartDate.Set(w.StartDate),
		academicWeekDTO.EndDate.Set(w.EndDate),
		academicWeekDTO.Period.Set(w.Period),
		academicWeekDTO.AcademicYearID.Set(w.AcademicYearID),
		academicWeekDTO.LocationID.Set(w.LocationID),
		academicWeekDTO.CreatedAt.Set(w.CreatedAt),
		academicWeekDTO.UpdatedAt.Set(w.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from academic week entity to academic week dto: %w", err)
	}
	if w.DeletedAt != nil {
		if err := academicWeekDTO.DeletedAt.Set(w.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return academicWeekDTO, nil
}
