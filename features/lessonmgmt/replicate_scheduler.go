package lessonmgmt

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/helper"
	cpb "github.com/manabie-com/backend/pkg/manabuf/calendar/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) createSchedulerToCalendarDB(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	now := time.Now()
	req := &cpb.CreateSchedulerRequest{
		StartDate: timestamppb.New(now),
		EndDate:   timestamppb.New(now.AddDate(0, 0, 1)),
		Frequency: cpb.Frequency_WEEKLY,
	}
	res, err := cpb.NewSchedulerModifierServiceClient(s.CalendarConn).CreateScheduler(helper.GRPCContext(ctx, "token", stepState.AuthToken), req)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.SchedulerID = res.SchedulerId
	stepState.Request = req
	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) schedulerSynced(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// wait for sync process done
	time.Sleep(5 * time.Second)
	var schedulerID string
	query := "SELECT scheduler_id from scheduler where scheduler_id = $1"
	err := s.BobDBTrace.QueryRow(ctx, query, stepState.SchedulerID).Scan(&schedulerID)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if len(schedulerID) == 0 {
		return StepStateToContext(ctx, stepState), fmt.Errorf("not found scheduler id `%s`", stepState.SchedulerID)
	}
	return StepStateToContext(ctx, stepState), nil
}
