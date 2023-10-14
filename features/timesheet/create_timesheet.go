package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	"github.com/manabie-com/backend/internal/timesheet/infrastructure/repository"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgx/v4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) newTimesheetDataForCurrentStaff(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	stepState.Request, err = buildCreateTimesheetRequest(ctx, stepState.CurrentUserID, true /*isForCurrentUserID*/)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newTimesheetDataForOtherStaff(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	stepState.Request, err = buildCreateTimesheetRequest(ctx, stepState.CurrentUserID, false /*isForCurrentUserID*/)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newTimesheetForExistingTimesheet(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	request := &pb.CreateTimesheetRequest{
		StaffId:       initStaffID,
		TimesheetDate: timestamppb.New(initTimesheetDate),
		LocationId:    initLocationTimesheet,
	}
	request.ListOtherWorkingHours = []*pb.OtherWorkingHoursRequest{
		{
			TimesheetConfigId: initTimesheetConfigID1,
			StartTime:         timestamppb.New(time.Now()),
			EndTime:           timestamppb.New(time.Now().Add(time.Hour)),
		},
	}

	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func buildCreateTimesheetRequest(ctx context.Context, staffID string, isForCurrentUserID bool) (*pb.CreateTimesheetRequest, error) {
	var (
		err     error
		request = &pb.CreateTimesheetRequest{
			StaffId:       staffID,
			TimesheetDate: timestamppb.New(generateRandomDate()),
			LocationId:    locationIDs[0],
		}
	)

	request.ListOtherWorkingHours = []*pb.OtherWorkingHoursRequest{
		{
			TimesheetConfigId: initTimesheetConfigID1,
			StartTime:         timestamppb.New(time.Now()),
			EndTime:           timestamppb.New(time.Now().Add(time.Hour)),
		},
	}

	if !isForCurrentUserID {
		request.StaffId, err = getStaffIDDifferenceCurrentUserID(ctx, staffID)
		if err != nil {
			return nil, err
		}
	}

	return request, nil
}

func (s *Suite) userCreateANewTimesheet(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Request == nil {
		stepState.Request = &pb.CreateTimesheetRequest{}
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr =
		pb.NewTimesheetServiceClient(s.TimesheetConn).CreateTimesheet(contextWithToken(ctx), stepState.Request.(*pb.CreateTimesheetRequest))

	if stepState.ResponseErr == nil {
		stepState.CurrentTimesheetID = stepState.
			Response.(*pb.CreateTimesheetResponse).TimesheetId
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) theTimesheetIsCreated(ctx context.Context, expectTimesheetCreated string) (context.Context, error) {
	var (
		stepState     = StepStateFromContext(ctx)
		timesheetRepo = repository.TimesheetRepoImpl{}
	)

	timesheet, err := timesheetRepo.FindTimesheetByTimesheetID(ctx, s.CommonSuite.TimesheetDB, database.Text(stepState.CurrentTimesheetID))

	if err != nil {
		if expectTimesheetCreated == "false" && err.Error() == pgx.ErrNoRows.Error() {
			return StepStateToContext(ctx, stepState), nil
		}

		return StepStateToContext(ctx, stepState), fmt.Errorf("failed to query Timesheet: %s", err)
	}

	createdRequest, ok := stepState.Request.(*pb.CreateTimesheetRequest)
	if !ok {
		return StepStateToContext(ctx, stepState), fmt.Errorf("expected stepState.Request to be *bpb.CreateTimesheetRequest, got %T", createdRequest)
	}
	ctx, err = s.ValidateTimesheetForCreatedRequest(ctx, timesheet, createdRequest)

	if expectTimesheetCreated == "true" && err != nil {
		return StepStateToContext(ctx, stepState), fmt.Errorf("validation failed for create Timesheet: %s", err)
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) ValidateTimesheetForCreatedRequest(ctx context.Context, timesheet *entity.Timesheet, req *pb.CreateTimesheetRequest) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if timesheet.StaffID.String != req.StaffId {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("expected %s for StaffIDs, got %s", req.StaffId, timesheet.StaffID.String)
	}

	if timesheet.LocationID.String != req.LocationId {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("expected %s for StaffIDs, got %s", req.StaffId, timesheet.StaffID.String)
	}

	if pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String() != timesheet.TimesheetStatus.String {
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("expected %s for TimesheetStatus, got %s", pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(), timesheet.TimesheetStatus.String)
	}

	return StepStateToContext(ctx, stepState), nil
}
