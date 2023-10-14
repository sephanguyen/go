package repo

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
	"go.uber.org/multierr"
)

type AcademicYear struct {
	AcademicYearID pgtype.Text
	Name           pgtype.Text
	StartDate      pgtype.Date
	EndDate        pgtype.Date
	UpdatedAt      pgtype.Timestamptz
	CreatedAt      pgtype.Timestamptz
	DeletedAt      pgtype.Timestamptz
}

func (ay *AcademicYear) FieldMap() ([]string, []interface{}) {
	return []string{
			"academic_year_id",
			"name",
			"start_date",
			"end_date",
			"updated_at",
			"created_at",
			"deleted_at",
		}, []interface{}{
			&ay.AcademicYearID,
			&ay.Name,
			&ay.StartDate,
			&ay.EndDate,
			&ay.UpdatedAt,
			&ay.CreatedAt,
			&ay.DeletedAt,
		}
}

func (ay *AcademicYear) TableName() string {
	return "academic_year"
}

func (ay *AcademicYear) ToAcademicYearDomain() *domain.AcademicYear {
	return &domain.AcademicYear{
		AcademicYearID: ay.AcademicYearID.String,
		Name:           ay.Name.String,
		StartDate:      ay.StartDate.Time,
		EndDate:        ay.EndDate.Time,
		CreatedAt:      ay.CreatedAt.Time,
		UpdatedAt:      ay.UpdatedAt.Time,
		DeletedAt:      &ay.DeletedAt.Time,
	}
}

func NewAcademicYearFromEntity(ay *domain.AcademicYear) (*AcademicYear, error) {
	academicYearDTO := &AcademicYear{}
	database.AllNullEntity(academicYearDTO)
	if err := multierr.Combine(
		academicYearDTO.AcademicYearID.Set(ay.AcademicYearID),
		academicYearDTO.Name.Set(ay.Name),
		academicYearDTO.StartDate.Set(ay.StartDate),
		academicYearDTO.EndDate.Set(ay.EndDate),
		academicYearDTO.CreatedAt.Set(ay.CreatedAt),
		academicYearDTO.UpdatedAt.Set(ay.UpdatedAt),
	); err != nil {
		return nil, fmt.Errorf("could not mapping from academic year entity to academic year dto: %w", err)
	}
	if ay.DeletedAt != nil {
		if err := academicYearDTO.DeletedAt.Set(ay.DeletedAt); err != nil {
			return nil, fmt.Errorf("could not set deleted_at: %w", err)
		}
	}
	return academicYearDTO, nil
}
