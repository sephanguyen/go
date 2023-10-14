package domain

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
)

type AcademicYear struct {
	AcademicYearID pgtype.Text
	Name           pgtype.Text
	StartDate      pgtype.Date
	EndDate        pgtype.Date
}

func (ay *AcademicYear) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"academic_year_id", "name", "start_date", "end_date"}
	values = []interface{}{&ay.AcademicYearID, &ay.Name, &ay.StartDate, &ay.EndDate}
	return
}

func (ay *AcademicYear) TableName() string {
	return "academic_year"
}

type AcademicYearRepository interface {
	GetCurrentAcademicYear(ctx context.Context, db database.Ext) (*domain.AcademicYear, error)
}
