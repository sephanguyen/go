package calendar

import (
	"context"
)

func (s *suite) anExistingDateInfoForDateAndLocation(ctx context.Context, date, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	stepState.DateTypeID = "spare"

	return s.userCreatesADateInfoForDateAndLocation(StepStateToContext(ctx, stepState), date, location)
}

func (s *suite) userUpdatesDateInfoForDateAndLocation(ctx context.Context, date, location string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	dateTypeID := stepState.DateTypeID
	openingTime := "10:00"
	if dateTypeID == "closed" {
		openingTime = ""
	}

	return s.prepareUpsertDateInfoRequest(StepStateToContext(ctx, stepState), date, location, dateTypeID, openingTime, "published", "")
}
