package timesheet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) approvesThisTimesheet(ctx context.Context) (context.Context, error) {
	// time sleep for lesson sync before approving
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	req := &pb.ApproveTimesheetRequest{
		TimesheetIds: stepState.CurrentTimesheetIDs,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewTimesheetStateMachineServiceClient(s.TimesheetConn).ApproveTimesheet(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) staffApprovesThisTimesheet(ctx context.Context, otherStaffGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.SignedAsAccount(ctx, otherStaffGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	ctx, err = s.approvesThisTimesheet(ctx)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetStatusChangedToApprove(ctx context.Context, approvedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch approvedStatus {
	case "successfully":
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if stepState.Response != nil {
			if !stepState.Response.(*pb.ApproveTimesheetResponse).Success {
				return ctx, fmt.Errorf("error cannot approved timesheet record")
			}

			err := s.checkTimesheetRecord(ctx, pb.TimesheetStatus_TIMESHEET_STATUS_APPROVED.String())
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	case "unsuccessfully":
		if stepState.ResponseErr == nil {
			return ctx, fmt.Errorf("error timesheet record should not be approved")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) staffHasanExistingSubmittedTimesheet(ctx context.Context, numOfTimesheet string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	numOfTimesheetInt, err := strconv.Atoi(numOfTimesheet)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	for i := 0; i < numOfTimesheetInt; i++ {
		err = s.buildCreateTimesheeBasedOnStatus(ctx, "SUBMITTED", "TODAY")
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) eachTimesheetsHasLessonRecordsWith(ctx context.Context, lessonStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.createLessonRecords(ctx, lessonStatus)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	return StepStateToContext(ctx, stepState), nil
}
