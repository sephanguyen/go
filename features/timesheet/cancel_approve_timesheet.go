package timesheet

import (
	"context"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) cancelApproveThisTimesheet(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.CancelApproveTimesheetRequest{
		TimesheetId: stepState.CurrentTimesheetID,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewTimesheetStateMachineServiceClient(s.TimesheetConn).CancelApproveTimesheet(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetStatusApproveChangedToSubmitted(ctx context.Context, approvedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch approvedStatus {
	case "successfully":
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if stepState.Response != nil {
			if !stepState.Response.(*pb.CancelApproveTimesheetResponse).Success {
				return ctx, fmt.Errorf("error cannot cancel approve timesheet record")
			}

			err := s.checkTimesheetRecord(ctx, pb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	case "unsuccessfully":
		if stepState.ResponseErr == nil {
			return ctx, fmt.Errorf("error timesheet record should not be submitted")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
