package calendar

import (
	"context"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) createScheduler(ctx context.Context, startDate, endDate time.Time, freq string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &cpb.CreateSchedulerRequest{
		StartDate: timestamppb.New(startDate),
		EndDate:   timestamppb.New(endDate),
		Frequency: cpb.Frequency(cpb.Frequency_value[strings.ToUpper(freq)]),
	}
	res, err := cpb.NewSchedulerModifierServiceClient(s.CalendarConn).CreateScheduler(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err == nil {
		stepState.SchedulerID = res.SchedulerId
	}
	stepState.Request = req
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) createSchedulerFromScenario(ctx context.Context, _startDate, _endDate, freq string) (context.Context, error) {
	startDate, _ := time.Parse(time.RFC3339, _startDate)
	endDate, _ := time.Parse(time.RFC3339, _endDate)
	return s.createScheduler(ctx, startDate, endDate, freq)
}

func (s *suite) randomScheduler(ctx context.Context) (context.Context, error) {
	startDate := time.Date(2022, 9, 20, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 30)
	freq := "WEEKLY"
	return s.createScheduler(ctx, startDate, endDate, freq)
}

func (s *suite) existedScheduler(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	repo := repositories.SchedulerRepo{}
	scheduler, err := repo.GetByID(ctx, s.CalendarDBTrace, stepState.SchedulerID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*cpb.CreateSchedulerRequest)

	if scheduler.SchedulerID != stepState.SchedulerID {
		return StepStateToContext(ctx, stepState), NotMatchError("scheduler_id", stepState.SchedulerID, scheduler.SchedulerID)
	}

	if !scheduler.StartDate.Equal(req.StartDate.AsTime()) {
		return StepStateToContext(ctx, stepState), NotMatchError("start_date", req.StartDate.AsTime(), scheduler.StartDate)
	}

	if !scheduler.EndDate.Equal(req.EndDate.AsTime()) {
		return StepStateToContext(ctx, stepState), NotMatchError("end_date", req.EndDate.AsTime(), scheduler.EndDate)
	}

	if scheduler.Frequency != strings.ToLower(req.Frequency.String()) {
		return StepStateToContext(ctx, stepState), NotMatchError("end_date", req.Frequency.String(), scheduler.Frequency)
	}
	return StepStateToContext(ctx, stepState), nil
}
