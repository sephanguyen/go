package domain

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/academic_year/domain"

	"github.com/jackc/pgtype"
)

type AcademicWeek struct {
	AcademicWeekID pgtype.Text
	WeekOrder      pgtype.Int2
	Name           pgtype.Text
	StartDate      pgtype.Date
	EndDate        pgtype.Date
	LocationID     pgtype.Text
}

func (aw *AcademicWeek) FieldMap() (fields []string, values []interface{}) {
	fields = []string{"academic_week_id", "week_order", "name", "start_date", "end_date", "location_id"}
	values = []interface{}{&aw.AcademicWeekID, &aw.WeekOrder, &aw.Name, &aw.StartDate, &aw.EndDate, &aw.LocationID}
	return
}

func (aw *AcademicWeek) TableName() string {
	return "academic_week"
}

type AcademicWeekRepository interface {
	GetByDateRange(ctx context.Context, db database.Ext, locationID string, academicWeeks []string, startDate time.Time, endDate time.Time) ([]*domain.AcademicWeek, error)
}
