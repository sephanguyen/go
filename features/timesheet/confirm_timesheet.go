package timesheet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) staffHasAnExistingApproveTimesheet(ctx context.Context, numOfTimesheet string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOfTimesheetInt, err := strconv.Atoi(numOfTimesheet)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for i := 0; i < numOfTimesheetInt; i++ {
		err = s.buildCreateTimesheeBasedOnStatus(ctx, "APPROVED", "TODAY")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) staffConfirmsThisTimesheet(ctx context.Context, userGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.SignedAsAccount(ctx, userGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.confirmsThisTimesheet(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) confirmsThisTimesheet(ctx context.Context) (context.Context, error) {
	// time sleep for lesson sync before approving
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	req := &pb.ConfirmTimesheetRequest{
		TimesheetIds: stepState.CurrentTimesheetIDs,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewTimesheetStateMachineServiceClient(s.TimesheetConn).ConfirmTimesheet(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetStatusesChangedToConfirm(ctx context.Context, confirmStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch confirmStatus {
	case "successfully":
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if stepState.Response != nil {
			if !stepState.Response.(*pb.ConfirmTimesheetResponse).Success {
				return ctx, fmt.Errorf("error cannot confirm timesheet record")
			}

			err := s.checkTimesheetRecord(ctx, pb.TimesheetStatus_TIMESHEET_STATUS_CONFIRMED.String())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	case "unsuccessfully":
		if stepState.ResponseErr == nil {
			return ctx, fmt.Errorf("error timesheet record should not be confirmed")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
