package calendar

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"go.uber.org/multierr"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) userChooseTheDate(ctx context.Context, date, locationID string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	Date, err := time.Parse("2006-01-02", date)

	if err != nil {
		return ctx, fmt.Errorf("parse datetime error: %w", err)
	}

	dinfo := &dto.DateInfo{
		Date:       Date,
		LocationID: locationID,
	}

	stepState.DateInfos = append(stepState.DateInfos, dinfo)
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) duplicateDateInfo(ctx context.Context, condition, sDate, eDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	dateInfos := stepState.DateInfos
	startDate, err1 := time.Parse("2006-01-02", sDate)
	endDate, err2 := time.Parse("2006-01-02", eDate)

	if err := multierr.Combine(err1, err2); err != nil {
		return ctx, fmt.Errorf("parse datetime error: %w", err)
	}

	for _, dateInfo := range dateInfos {
		req := &cpb.DuplicateDateInfoRequest{}
		repeatInfopb := &cpb.RepeatInfo{}
		repeatInfopb.Condition = condition
		repeatInfopb.StartDate = timestamppb.New(startDate)
		repeatInfopb.EndDate = timestamppb.New(endDate)

		dateInfopb := &cpb.DateInfo{}
		dateInfopb.Date = timestamppb.New(dateInfo.Date)
		dateInfopb.LocationId = dateInfo.LocationID

		if !startDate.IsZero() && !endDate.IsZero() {
			req = &cpb.DuplicateDateInfoRequest{
				DateInfo:   dateInfopb,
				RepeatInfo: repeatInfopb,
			}
		}

		stepState.Response, stepState.ResponseErr = cpb.NewDateInfoModifierServiceClient(s.CalendarConn).DuplicateDateInfo(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)

		if stepState.ResponseErr != nil {
			return StepStateToContext(ctx, stepState), stepState.ResponseErr
		}
	}
	return StepStateToContext(ctx, stepState), nil
}
