package timesheet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) deletesThisTimesheet(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	req := &pb.DeleteTimesheetRequest{
		TimesheetId: stepState.CurrentTimesheetID,
	}

	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr = pb.NewTimesheetStateMachineServiceClient(s.TimesheetConn).DeleteTimesheet(contextWithToken(ctx), req)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) anExistingTimesheetForCurrentStaff(ctx context.Context, timesheetStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	err := s.buildCreateTimesheeBasedOnStatus(ctx, timesheetStatus, "TODAY")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) anExistingTimesheetForOtherStaff(ctx context.Context, timesheetStatus, otherStaffGroup string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	ctx, err := s.SignedAsAccount(ctx, otherStaffGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = s.buildCreateTimesheeBasedOnStatus(ctx, timesheetStatus, "TODAY")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) buildCreateTimesheeBasedOnStatus(ctx context.Context, timesheetStatusFormat, timesheetDate string) error {
	stepState := StepStateFromContext(ctx)
	var getTimesheetDateTime time.Time

	timesheetStatus := convertTimesheetStatusFormat(timesheetStatusFormat)

	switch timesheetDate {
	case "YESTERDAY":
		getTimesheetDateTime = time.Now().AddDate(0, 0, -1)
	case "TOMORROW":
		getTimesheetDateTime = time.Now().AddDate(0, 0, 1)
	case "5DAYS FROM TODAY":
		getTimesheetDateTime = time.Now().AddDate(0, 0, 5)
	case "2MONTHS FROM TODAY":
		getTimesheetDateTime = time.Now().AddDate(0, 2, 0)
	default:
		getTimesheetDateTime = time.Now()
	}

	timesheetID, err := initTimesheet(ctx, stepState.CurrentUserID, locationIDs[0], timesheetStatus, getTimesheetDateTime, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
	if err != nil {
		return err
	}
	stepState.CurrentTimesheetIDs = append(stepState.CurrentTimesheetIDs, timesheetID)
	stepState.CurrentTimesheetID = timesheetID
	return nil
}

func (s *Suite) timesheetIsDeleted(ctx context.Context, deleteStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	switch deleteStatus {
	case "successfully":
		if stepState.ResponseErr != nil {
			return ctx, stepState.ResponseErr
		}
		if stepState.Response != nil {
			if !stepState.Response.(*pb.DeleteTimesheetResponse).Success {
				return ctx, fmt.Errorf("error cannot delete timesheet record")
			}

			err := s.checkTimesheetDeleted(ctx)
			if err != nil {
				return StepStateToContext(ctx, stepState), err
			}
		}
	case "unsuccessfully":
		if stepState.ResponseErr == nil {
			return ctx, fmt.Errorf("error timesheet record should not be deleted")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetHasLessonRecords(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	// create lesson records
	ctx, err := s.createLessonRecords(ctx, "PUBLISHED")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) checkTimesheetDeleted(ctx context.Context) error {
	stepState := StepStateFromContext(ctx)
	var count int
	stmt := `
		SELECT
			count(timesheet_id)
		FROM
			timesheet
		WHERE
			timesheet_id = $1
		AND
			deleted_at IS NOT NULL
		`
	err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentTimesheetID).Scan(&count)
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("unexpected %d timesheet record", count)
	}

	return nil
}

func (s *Suite) timesheetHasOtherWorkingHoursRecords(ctx context.Context, otherWorkingHoursRecords string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	numOfOtherWorkingHoursRecord, err := strconv.Atoi(otherWorkingHoursRecords)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for i := 0; i < numOfOtherWorkingHoursRecord; i++ {
		_, err := initOtherWorkingHours(ctx, stepState.CurrentTimesheetID, initTimesheetConfigID1, time.Now(), strconv.Itoa(constants.ManabieSchool))
		if err != nil {
			return nil, err
		}
	}

	stepState.NumberOfOtherWorkingHours = int32(numOfOtherWorkingHoursRecord)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetOtherWorkingHoursRecordsIsDeleted(ctx context.Context, deleteStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count int32
	stmt := `
		SELECT
			count(other_working_hours_id)
		FROM
			other_working_hours
		WHERE
			timesheet_id = $1
		AND
			deleted_at IS NOT NULL
		`
	if deleteStatus == "unsuccessfully" {
		stmt = `
		SELECT
			count(other_working_hours_id)
		FROM
			other_working_hours
		WHERE
			timesheet_id = $1
		AND
			deleted_at IS NULL
		`
	}

	err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentTimesheetID).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != stepState.NumberOfOtherWorkingHours {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected %d other working hours record", count)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetHasTransportExpensesRecords(ctx context.Context, transportExpensesRecords string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	numOfTransportExpensesRecords, err := strconv.Atoi(transportExpensesRecords)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	for i := 0; i < numOfTransportExpensesRecords; i++ {
		_, err := initTransportExpenses(ctx, stepState.CurrentTimesheetID, strconv.Itoa(constants.ManabieSchool))

		if err != nil {
			return nil, err
		}
	}

	stepState.NumberOfTransportExpensesRecords = int32(numOfTransportExpensesRecords)

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) timesheetTransportExpensesIsDeleted(ctx context.Context, deleteStatus string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var count int32
	stmt := `
		SELECT
			count(transportation_expense_id)
		FROM
			transportation_expense
		WHERE
			timesheet_id = $1
		AND
			deleted_at IS NOT NULL
		`
	if deleteStatus == "unsuccessfully" {
		stmt = `
		SELECT
			count(transportation_expense_id)
		FROM
			transportation_expense
		WHERE
			timesheet_id = $1
		AND
			deleted_at IS NULL
		`
	}

	err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.CurrentTimesheetID).Scan(&count)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	if count != stepState.NumberOfTransportExpensesRecords {
		return StepStateToContext(ctx, stepState), fmt.Errorf("unexpected %d transportation expenses record", count)
	}

	return StepStateToContext(ctx, stepState), nil
}
