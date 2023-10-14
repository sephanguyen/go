package mastermgmt

import (
	"context"
	"fmt"
	"strings"

	mpb "github.com/manabie-com/backend/pkg/manabuf/mastermgmt/v1"
)

func (s *suite) exportAcademicCalendar(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &mpb.ExportAcademicCalendarRequest{
		AcademicYearId: stepState.AcademicYearIDs[0],
		LocationId:     stepState.CenterIDs[0],
	}
	stepState.Response, stepState.ResponseErr = mpb.NewAcademicYearServiceClient(s.MasterMgmtConn).
		ExportAcademicCalendar(contextWithToken(s, ctx), req)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) returnsAcademicCalendarInCsv(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.ResponseErr != nil {
		return ctx, fmt.Errorf("can not export academic calendar: %s", stepState.ResponseErr.Error())
	}
	resp := stepState.Response.(*mpb.ExportAcademicCalendarResponse)
	expectedRows := stepState.ExpectedCSV
	respLines := strings.Split(string(resp.Data), "\n")

	totalRespRowValues := len(respLines) - 1
	totalExpectRowValues := len(expectedRows) - 1
	// exclude header

	if expectedRows[0] != respLines[0] {
		return ctx, fmt.Errorf("academic calendar csv header is not valid.\nexpected: %s\ngot: %s", expectedRows[0], respLines[0])
	}
	for index, v := range expectedRows[1:] {
		i := 0
		for float64(i) < float64((totalRespRowValues-index-1))/float64(totalExpectRowValues) {
			// loop to check value on all child locations
			if !strings.Contains(respLines[i*totalExpectRowValues+index+1], v) {
				return ctx, fmt.Errorf("academic calendar csv is not valid.\nexpected:%s\ngot %s", v, respLines[i*totalExpectRowValues+index+1])
			}
			i++
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
