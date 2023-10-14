package timesheet

import (
	"context"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) cancelSubmitThisTimesheet(ctx context.Context) (context.Context, error) {
	// time sleep for lesson sync before submitting
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	req := &pb.CancelSubmissionTimesheetRequest{
		TimesheetId: stepState.CurrentTimesheetID,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewTimesheetStateMachineServiceClient(s.TimesheetConn).CancelSubmissionTimesheet(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetStatusChangedToDraft(ctx context.Context, approvedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch approvedStatus {
	case "successfully":
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if stepState.Response != nil {
			if !stepState.Response.(*pb.CancelSubmissionTimesheetResponse).Success {
				return ctx, fmt.Errorf("error cannot cancel submit timesheet record")
			}

			err := s.checkTimesheetRecord(ctx, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	case "unsuccessfully":
		if stepState.ResponseErr == nil {
			return ctx, fmt.Errorf("error timesheet record should not be draft")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
