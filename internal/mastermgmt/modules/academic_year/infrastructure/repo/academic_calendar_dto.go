package repo

import "github.com/jackc/pgtype"

type AcademicCalendar struct {
	AcademicWeekID     pgtype.Text
	WeekOrder          pgtype.Text
	Name               pgtype.Text
	StartDate          pgtype.Date
	EndDate            pgtype.Date
	Period             pgtype.Text
	AcademicClosedDays pgtype.Text
	AcademicYear       pgtype.Text
	Location           pgtype.Text
}

func (ac *AcademicCalendar) FieldMap() ([]string, []interface{}) {
	return []string{
			"academic_week_id",
			"week_order",
			"name",
			"start_date",
			"end_date",
			"period",
			"academic_closed_day",
			"academic_year",
			"location",
		}, []interface{}{
			&ac.AcademicWeekID,
			&ac.WeekOrder,
			&ac.Name,
			&ac.StartDate,
			&ac.EndDate,
			&ac.Period,
			&ac.AcademicClosedDays,
			&ac.AcademicYear,
			&ac.Location,
		}
}

func (ac *AcademicCalendar) TableName() string {
	return "academic_calendar"
}
