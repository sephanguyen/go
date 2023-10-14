package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/timesheet/domain/common"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) submitsThisTimesheet(ctx context.Context) (context.Context, error) {
	// time sleep for lesson sync before submitting
	time.Sleep(3 * time.Second)
	stepState := StepStateFromContext(ctx)
	req := &pb.SubmitTimesheetRequest{
		TimesheetId: stepState.CurrentTimesheetID,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewTimesheetStateMachineServiceClient(s.TimesheetConn).SubmitTimesheet(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetStatusChangedToSubmitted(ctx context.Context, approvedStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	switch approvedStatus {
	case "successfully":
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if stepState.Response != nil {
			if !stepState.Response.(*pb.SubmitTimesheetResponse).Success {
				return ctx, fmt.Errorf("error cannot submit timesheet record")
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

func (s *Suite) checkTimesheetRecord(ctx context.Context, timesheetStatus string) error {
	stepState := StepStateFromContext(ctx)
	var count int

	stmt := fmt.Sprintf(`
		SELECT
			count(timesheet_id)
		FROM
			timesheet
		WHERE
			timesheet_id IN (%s)
		AND
			deleted_at IS NULL
		AND
			timesheet_status = $1;
		`, common.ConcatQueryValue(stepState.CurrentTimesheetIDs...))
	err := s.TimesheetDB.QueryRow(ctx, stmt, timesheetStatus).Scan(&count)
	if err != nil {
		return err
	}

	if count != len(stepState.CurrentTimesheetIDs) {
		return fmt.Errorf("unexpected %d timesheet record should be %d count affected", count, len(stepState.CurrentTimesheetIDs))
	}

	return nil
}

func (s *Suite) anExistingTimesheetWithDateForCurrentStaff(ctx context.Context, timesheetStatus, timesheetDate string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.buildCreateTimesheeBasedOnStatus(ctx, timesheetStatus, timesheetDate)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
