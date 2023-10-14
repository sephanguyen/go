package timesheet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pbb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func removeFromSlicesChangesOrder(req []*pb.OtherWorkingHoursRequest, elementIndex int) []*pb.OtherWorkingHoursRequest {
	if elementIndex > len(req) {
		return nil
	}
	req[elementIndex] = req[len(req)-1]
	req[len(req)-1] = (*pb.OtherWorkingHoursRequest)(nil)
	req = req[:len(req)-1]
	return req
}

func (s *Suite) buildDataForUpdateTimesheetWithOWHsRequest(ctx context.Context, numOfOWHs int, status string) (*pb.UpdateTimesheetRequest, error) {
	var (
		timesheetID string
		err         error
		stepState   = StepStateFromContext(ctx)
	)

	// make timesheet
	genDate := generateRandomDate()
	timesheetID, err = initTimesheet(ctx, stepState.CurrentUserID, locationIDs[0], status, genDate, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
	if err != nil {
		return nil, err
	}

	// make list OWHs
	listUpdateOWHs := make([]*pb.OtherWorkingHoursRequest, 0, numOfOWHs)
	for i := 0; i < numOfOWHs; i++ {
		genTime := time.Date(genDate.Year(), genDate.Month(), genDate.Day(), i+8 /*start from 8h AM*/, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))
		owhsE, err := initOtherWorkingHours(ctx, timesheetID, initTimesheetConfigID1, genTime, strconv.Itoa(constants.ManabieSchool))
		if err != nil {
			return nil, err
		}

		owhsTemp := &pb.OtherWorkingHoursRequest{
			OtherWorkingHoursId: owhsE.ID.String,
			TimesheetConfigId:   owhsE.TimesheetConfigID.String,
			StartTime:           timestamppb.New(owhsE.StartTime.Time),
			EndTime:             timestamppb.New(owhsE.EndTime.Time),
			Remarks:             owhsE.Remarks.String,
		}

		listUpdateOWHs = append(listUpdateOWHs, owhsTemp)
	}

	request := &pb.UpdateTimesheetRequest{
		TimesheetId:           timesheetID,
		ListOtherWorkingHours: listUpdateOWHs,
	}

	return request, nil
}

func (s *Suite) newUpdateTimesheetWithOWHsDataForCurrentStaff(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		err     error
		request *pb.UpdateTimesheetRequest
	)

	now := time.Now()
	owhs := &pb.OtherWorkingHoursRequest{
		TimesheetConfigId: initTimesheetConfigID1,
		StartTime:         timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
		EndTime:           timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
		Remarks:           randStringBytes(10),
	}
	owhs2 := &pb.OtherWorkingHoursRequest{
		TimesheetConfigId: initTimesheetConfigID1,
		StartTime:         timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
		EndTime:           timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, 30, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
		Remarks:           randStringBytes(10),
	}
	switch action {
	case "insert":
		currentOWHsListLen := 0
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListOtherWorkingHours = append(request.ListOtherWorkingHours, owhs)

	case "update":
		currentOWHsListLen := 1
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListOtherWorkingHours[0].StartTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].EndTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].Remarks = randStringBytes(15)

	case "delete":
		currentOWHsListLen := 2
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.ListOtherWorkingHours = removeFromSlicesChangesOrder(request.ListOtherWorkingHours, 0)

	case "insert,delete":
		currentOWHsListLen := 5
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.ListOtherWorkingHours = removeFromSlicesChangesOrder(request.ListOtherWorkingHours, 0)

		request.ListOtherWorkingHours = append(request.ListOtherWorkingHours, owhs)

	case "insert,update":
		currentOWHsListLen := 1
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListOtherWorkingHours[0].StartTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].EndTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].Remarks = randStringBytes(15)

		request.ListOtherWorkingHours = append(request.ListOtherWorkingHours, owhs)

	case "update,delete":
		currentOWHsListLen := 2
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListOtherWorkingHours[0].StartTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].EndTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].Remarks = randStringBytes(15)

		request.ListOtherWorkingHours = removeFromSlicesChangesOrder(request.ListOtherWorkingHours, 1)

	case "insert,update,delete":
		currentOWHsListLen := 5
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListOtherWorkingHours[0].StartTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].EndTime = timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 45, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].Remarks = randStringBytes(15)

		request.ListOtherWorkingHours = removeFromSlicesChangesOrder(request.ListOtherWorkingHours, 1)

		request.ListOtherWorkingHours = append(request.ListOtherWorkingHours, owhs)
	case "have-5,insert-2,delete-1":
		currentOWHsListLen := 5
		request, err = s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.ListOtherWorkingHours = removeFromSlicesChangesOrder(request.ListOtherWorkingHours, 0)

		request.ListOtherWorkingHours = append(request.ListOtherWorkingHours, owhs, owhs2)

	default:
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("invalid action: %v", action)
	}

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) userUpdateInvalidArgsForTimesheet(ctx context.Context, invalidArgCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	currentOWHsListLen := 5
	request, err := s.buildDataForUpdateTimesheetWithOWHsRequest(ctx, currentOWHsListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	now := time.Now()
	year, month, day := now.Date()

	owhs := &pb.OtherWorkingHoursRequest{
		TimesheetConfigId: initTimesheetConfigID1,
		StartTime:         timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
		EndTime:           timestamppb.New(time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 30, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
		Remarks:           randStringBytes(10),
	}

	switch invalidArgCase {
	case "empty timesheet id":
		request.TimesheetId = ""

	case "remark > 500 characters":
		request.Remark = randStringBytes(remarksLimit + 1) // +1 over limit

	case "other working hours list over 5":
		request.ListOtherWorkingHours = removeFromSlicesChangesOrder(request.ListOtherWorkingHours, 0)
		request.ListOtherWorkingHours = append(request.ListOtherWorkingHours, owhs)

	case "other working hours working type empty":
		request.ListOtherWorkingHours[0].TimesheetConfigId = ""

	case "other working hours start time null":
		request.ListOtherWorkingHours[0].StartTime = (*timestamppb.Timestamp)(nil)

	case "other working hours end time null":
		request.ListOtherWorkingHours[0].EndTime = (*timestamppb.Timestamp)(nil)

	case "other working hours end time before start time":
		request.ListOtherWorkingHours[0].StartTime = timestamppb.New(time.Date(year, month, day, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].EndTime = timestamppb.New(time.Date(year, month, day, 10, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))

	case "other working hours end time == start time":
		request.ListOtherWorkingHours[0].StartTime = timestamppb.New(time.Date(year, month, day, 10, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].EndTime = request.ListOtherWorkingHours[0].StartTime

	case "other working hours start time != end time date":
		request.ListOtherWorkingHours[0].StartTime = timestamppb.New(time.Date(year, month, 10, 10, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))
		request.ListOtherWorkingHours[0].EndTime = timestamppb.New(time.Date(year, month, 11, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP)))

	case "other working hours remarks > 100 character":
		request.ListOtherWorkingHours[0].Remarks = randStringBytes(otherWorkingHoursRemarksLimit + 1)

	case "other working hours working type invalid":
		request.ListOtherWorkingHours[0].TimesheetConfigId = invalidTimesheetConfigID3

	default:
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("invalid example argument: %v", invalidArgCase)
	}

	stepState.Request = request
	return StepStateToContext(ctx, stepState), nil
}
