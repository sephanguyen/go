package calendar

import (
	"context"

	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/calendar/infrastructure/repositories"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *suite) updateScheduler(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	createReq := stepState.Request.(*cpb.CreateSchedulerRequest)
	endDate := createReq.EndDate.AsTime().AddDate(0, 0, 1)
	req := &cpb.UpdateSchedulerRequest{
		SchedulerId: stepState.SchedulerID,
		EndDate:     timestamppb.New(endDate),
	}
	_, err := cpb.NewSchedulerModifierServiceClient(s.CalendarConn).UpdateScheduler(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	stepState.Request = req
	stepState.ResponseErr = err
	return StepStateToContext(ctx, stepState), nil
}

func (s *suite) updatedScheduler(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	repo := repositories.SchedulerRepo{}
	scheduler, err := repo.GetByID(ctx, s.CalendarDBTrace, stepState.SchedulerID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	req := stepState.Request.(*cpb.UpdateSchedulerRequest)

	if !scheduler.EndDate.Equal(req.EndDate.AsTime()) {
		return StepStateToContext(ctx, stepState), NotMatchError("end_date", req.EndDate.AsTime(), scheduler.EndDate)
	}
	return StepStateToContext(ctx, stepState), nil
}
