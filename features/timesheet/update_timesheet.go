package timesheet

import (
	"context"
	"strconv"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) buildUpdateTimesheetRequest(ctx context.Context, isForCurrentUserID bool, timesheetStatus string) (context.Context, error) {
	var (
		timesheetID string
		err         error
		stepState   = StepStateFromContext(ctx)
	)

	timesheetID, err = getOneTimesheetIDInDB(ctx, stepState.CurrentUserID, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10), isForCurrentUserID)
	if timesheetID == "" {
		timesheetID, err = initTimesheet(ctx, stepState.CurrentUserID, locationIDs[0], timesheetStatus, generateRandomDate(), strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	owhs := &pb.OtherWorkingHoursRequest{
		TimesheetConfigId: initTimesheetConfigID1,
		StartTime:         timestamppb.Now(),
		EndTime:           timestamppb.Now(),
	}

	stepState.Request = &pb.UpdateTimesheetRequest{
		TimesheetId: timesheetID,
		ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
			owhs,
		},
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateTimesheet(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Request == nil {
		stepState.Request = &pb.UpdateTimesheetRequest{}
	}
	stepState.Response, stepState.ResponseErr =
		pb.NewTimesheetServiceClient(s.TimesheetConn).UpdateTimesheet(contextWithToken(ctx), stepState.Request.(*pb.UpdateTimesheetRequest))

	if stepState.ResponseErr != nil {
		stepState.CurrentTimesheetID = stepState.
			Request.(*pb.UpdateTimesheetRequest).TimesheetId
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newUpdatedTimesheetDataWithStatusForCurrentStaff(ctx context.Context, timesheetStatusFormat string) (context.Context, error) {
	timesheetStatus := convertTimesheetStatusFormat(timesheetStatusFormat)

	stepState := StepStateFromContext(ctx)
	ctx, err := s.buildUpdateTimesheetRequest(ctx, true /*isForCurrentUserID*/, timesheetStatus)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newUpdatedTimesheetDataWithStatusForOtherStaff(ctx context.Context, timesheetStatusFormat, otherStaffGroup string) (context.Context, error) {
	timesheetStatus := convertTimesheetStatusFormat(timesheetStatusFormat)

	stepState := StepStateFromContext(ctx)

	ctx, err := s.SignedAsAccount(ctx, otherStaffGroup)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	// create timesheet record of the other staff signed in and build request
	ctx, err = s.buildUpdateTimesheetRequest(ctx, true /*isForCurrentUserID*/, timesheetStatus)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) adminChangeUserTimesheetStatus(ctx context.Context, timesheetStatusFormat string) (context.Context, error) {
	timesheetStatus := convertTimesheetStatusFormat(timesheetStatusFormat)

	stepState := StepStateFromContext(ctx)

	ctx, err := s.SignedAsAccount(ctx, "staff granted role school admin")
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	err = updateTimesheetStatusInDB(ctx, stepState.CurrentTimesheetIDs[0], timesheetStatus)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}
