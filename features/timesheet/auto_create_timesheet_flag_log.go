package timesheet

import (
	"context"
	"fmt"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) countLogRecordOfUser(ctx context.Context, flagOnStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	var count int32

	stmt := `
		SELECT
			count(staff_id)
		FROM
			auto_create_flag_activity_log
		WHERE
			staff_id = $1
		AND
			flag_on = $2
		AND
			deleted_at IS NULL
		`
	err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentStaffID, flagOnStatus).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if err != nil && count > 1 {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.NumberLogRecordsOfCurrentStaffID = count

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkOnelogRecordIsInserted(ctx context.Context, flagOnStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if stepState.Response != nil {
		if !stepState.Response.(*pb.UpdateAutoCreateTimesheetFlagResponse).Successful {
			return ctx, fmt.Errorf("error cannot upsert auto create timesheet flag record")
		}

		var count int

		stmt := `
		SELECT
			count(staff_id)
		FROM
			auto_create_flag_activity_log
		WHERE
			staff_id = $1
		AND
			flag_on  = $2
		AND
			deleted_at IS NULL
		`
		err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentStaffID, flagOnStatus).Scan(&count)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		if err != nil && int(stepState.NumberLogRecordsOfCurrentStaffID) != count+1 {
			return StepStateToContext(ctx, stepState), err
		}
	}

	return StepStateToContext(ctx, stepState), nil
}
