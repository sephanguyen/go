package mastermgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) validAcademicCalendarCsvPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationID := stepState.CenterIDs[0]
	academicYearID := stepState.AcademicYearIDs[0]
	week1 := "Week " + idutil.ULIDNow()
	week2 := "Week " + idutil.ULIDNow()
	week3 := "Week " + idutil.ULIDNow()
	const format = "%s,%s,%s,%s,%s,%s"
	r1 := fmt.Sprintf(format, "1", week1, "2023-04-01", "2023-04-07", "Term 1", "2023-04-02")
	r2 := fmt.Sprintf(format, "2", week2, "2023-04-11", "2023-04-14", "Term 1", "")
	r3 := fmt.Sprintf(format, "3", week3, "2023-04-15", "2023-04-21", "Term 1", "2023-04-16;2023-04-17")

	csv := fmt.Sprintf(`order,academic_week,start_date,end_date,period,academic_closed_day
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportAcademicCalendarRequest{
		Payload:            []byte(csv),
		LocationId:         locationID,
		AcademicYearId:     academicYearID,
		AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	expectedRows := [][]string{
		{
			"academic_week_id", "order", "name", "start_date", "end_date", "period", "academic_closed_day", "academic_year", "location",
		},
		{
			"1", week1, "2023-04-01", "2023-04-07", "Term 1", "2023-04-02",
		},
		{
			"2", week2, "2023-04-11", "2023-04-14", "Term 1", "",
		},
		{
			"3", week3, "2023-04-15", "2023-04-21", "Term 1", "2023-04-16;2023-04-17",
		},
	}

	stepState.ExpectedCSV = s.getQuotedCSVRows(expectedRows)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) importAcademicCalendar(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = mpb.NewAcademicYearServiceClient(s.MasterMgmtConn).
		ImportAcademicCalendar(contextWithToken(s, ctx), stepState.Request.(*mpb.ImportAcademicCalendarRequest))
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) addAcademicYear(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	newAcademicYearID := idutil.ULIDNow()
	newAcademicYearName := "Year" + idutil.ULIDNow()

	query := `INSERT INTO academic_year (academic_year_id, name, start_date, end_date, created_at, updated_at, deleted_at, resource_path)
		values ($1, $2, now(), now(), now(), now(), null, '-2147483648')
		ON CONFLICT ON CONSTRAINT pk__academic_year DO UPDATE SET deleted_at = null`
	_, err := s.MasterMgmtDBTrace.Exec(ctx, query, newAcademicYearID, newAcademicYearName)

	if err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("cannot insert academic year with `id:%s`, %v", newAcademicYearID, err)
	}
	stepState.AcademicYearIDs = append(stepState.AcademicYearIDs, newAcademicYearID)

	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) invalidAcademicCalendarCsvPayload(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	locationID := stepState.CenterIDs[0]
	format := "%s,%s,%s,%s,%s"
	r1 := fmt.Sprintf(format, "Week "+idutil.ULIDNow(), "2023-04-01", "invalid-date", "Term 1", "2023-04-02")
	r2 := fmt.Sprintf(format, "Week "+idutil.ULIDNow(), "2023-04-11", "2023-04-14", "Term 1", "")
	r3 := fmt.Sprintf(format, "Week "+idutil.ULIDNow(), "2023-04-15", "2023-04-21", "Term 1", "2023-04-16;2023-04-17")

	csv := fmt.Sprintf(`academic_week,start_date,end_date,period,academic_closed_day
	%s
	%s
	%s`, r1, r2, r3)
	stepState.Request = &mpb.ImportAcademicCalendarRequest{
		Payload:            []byte(csv),
		LocationId:         locationID,
		AcademicYearId:     stepState.AcademicYearIDs[0],
		AcademicClosedDays: []string{"2023-04-08", "2023-04-09", "2023-04-10"},
	}
	stepState.ValidCsvRows = []string{r1, r2, r3}
	return StepStateToContext(ctx, stepState), nil
}
