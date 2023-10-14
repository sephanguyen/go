package repository

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/master_data/domain"
)

type AcademicWeekRepository struct{}

func (l *AcademicWeekRepository) GetByDateRange(ctx context.Context, db database.Ext, locationID string, academicWeeks []string, startDate time.Time, endDate time.Time) ([]*domain.AcademicWeek, error) {
	ctx, span := interceptors.StartSpan(ctx, "AcademicWeekRepository.GetByDateRange")
	defer span.End()
	query := `select aw.academic_week_id ,aw.week_order, aw."name", aw.start_date ,aw.end_date,aw.location_id  
	   from academic_week aw  
	   where aw.location_id = $1 and aw.week_order::text = any($2) and aw.end_date >= ($3 at time zone 'Asia/Ho_Chi_Minh')::date and aw.start_date <= ($4 at time zone 'Asia/Ho_Chi_Minh')::date
	   order by aw.week_order`

	row, err := db.Query(ctx, query, locationID, academicWeeks, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	res := []*domain.AcademicWeek{}
	for row.Next() {
		academicWeek := &domain.AcademicWeek{}
		_, value := academicWeek.FieldMap()
		if err = row.Scan(value...); err != nil {
			return nil, err
		}
		res = append(res, academicWeek)
	}
	if err = row.Err(); err != nil {
		return nil, err
	}
	return res, nil
}
