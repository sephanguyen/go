package timesheet

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pbb "github.com/manabie-com/backend/pkg/genproto/bob"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) newTimesheetDataWithOtherWorkingHours(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	stepState.Request, err = buildCreateTimesheetWithOtherWorkingHoursRequest(ctx, stepState.CurrentUserID, true /*isForCurrentUserID*/)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newTimesheetDataWithInvalidRequest(ctx context.Context, invalidArgCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timesheetDate := generateRandomDate()
	year, month, day := timesheetDate.Date()

	request := &pb.CreateTimesheetRequest{
		StaffId:       stepState.CurrentUserID,
		TimesheetDate: timestamppb.New(timesheetDate),
		LocationId:    locationIDs[0],
		Remark:        timesheetRemark,
	}

	switch invalidArgCase {
	case "empty StaffId":
		request.StaffId = ""
	case "empty LocationId":
		request.LocationId = ""
	case "null Date":
		request.TimesheetDate = (*timestamppb.Timestamp)(nil)
	case "remark > 500 characters":
		request.Remark = randStringBytes(remarksLimit + 1) // +1 over limit
	case "other working hours list over 5":
		listOWHsTemp := make([]*pb.OtherWorkingHoursRequest, 0, listOtherWorkingHoursLimit+1)
		for i := 0; i <= (listOtherWorkingHoursLimit + 1); i++ { // +1 over limit
			owhs := &pb.OtherWorkingHoursRequest{
				TimesheetConfigId: initTimesheetConfigID1,
				StartTime:         timestamppb.New(time.Date(year, month, day, 10, i*45, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				EndTime:           timestamppb.New(time.Date(year, month, day, 11, i*45, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
			}

			listOWHsTemp = append(listOWHsTemp, owhs)
		}
		request.ListOtherWorkingHours = listOWHsTemp
	case "other working hours working type empty":
		listOWHsTemp := []*pb.OtherWorkingHoursRequest{
			{
				TimesheetConfigId: "",
				StartTime:         timestamppb.New(time.Date(year, month, day, 10, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				EndTime:           timestamppb.New(time.Date(year, month, day, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
			},
		}
		request.ListOtherWorkingHours = listOWHsTemp
	case "other working hours start time null":
		listOWHsTemp := []*pb.OtherWorkingHoursRequest{
			{
				TimesheetConfigId: initTimesheetConfigID1,
				StartTime:         (*timestamppb.Timestamp)(nil),
				EndTime:           timestamppb.New(time.Date(year, month, day, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
			},
		}
		request.ListOtherWorkingHours = listOWHsTemp
	case "other working hours end time null":
		listOWHsTemp := []*pb.OtherWorkingHoursRequest{
			{
				TimesheetConfigId: initTimesheetConfigID1,
				StartTime:         timestamppb.New(time.Date(year, month, day, 10, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				EndTime:           (*timestamppb.Timestamp)(nil),
			},
		}
		request.ListOtherWorkingHours = listOWHsTemp
	case "other working hours end time before start time":
		listOWHsTemp := []*pb.OtherWorkingHoursRequest{
			{
				TimesheetConfigId: initTimesheetConfigID1,
				StartTime:         timestamppb.New(time.Date(year, month, day, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				EndTime:           timestamppb.New(time.Date(year, month, day, 10, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
			},
		}
		request.ListOtherWorkingHours = listOWHsTemp
	case "other working hours start time != end time date":
		listOWHsTemp := []*pb.OtherWorkingHoursRequest{
			{
				TimesheetConfigId: initTimesheetConfigID1,
				StartTime:         timestamppb.New(time.Date(year, month, 10, 10, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				EndTime:           timestamppb.New(time.Date(year, month, 11, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
			},
		}
		request.ListOtherWorkingHours = listOWHsTemp
	case "other working hours remarks > 100 character":
		listOWHsTemp := []*pb.OtherWorkingHoursRequest{
			{
				TimesheetConfigId: initTimesheetConfigID1,
				StartTime:         timestamppb.New(time.Date(year, month, day, 10, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				EndTime:           timestamppb.New(time.Date(year, month, day, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				Remarks:           randStringBytes(otherWorkingHoursRemarksLimit + 1),
			},
		}
		request.ListOtherWorkingHours = listOWHsTemp
	case "other working hours working type invalid":
		listOWHsTemp := []*pb.OtherWorkingHoursRequest{
			{
				TimesheetConfigId: invalidTimesheetConfigID3,
				StartTime:         timestamppb.New(time.Date(year, month, day, 10, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
				EndTime:           timestamppb.New(time.Date(year, month, day, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
			},
		}
		request.ListOtherWorkingHours = listOWHsTemp
	default:
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("invalid example argument: %v", invalidArgCase)
	}

	stepState.Request = request
	return StepStateToContext(ctx, stepState), nil
}

func buildCreateTimesheetWithOtherWorkingHoursRequest(ctx context.Context, staffID string, isForCurrentUserID bool) (*pb.CreateTimesheetRequest, error) {
	var err error
	timesheetDate := generateRandomDate()
	year, month, day := timesheetDate.Date()

	listOWHs := []*pb.OtherWorkingHoursRequest{
		{
			TimesheetConfigId: initTimesheetConfigID1,
			StartTime:         timestamppb.New(time.Date(year, month, day, 10, 0, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
			EndTime:           timestamppb.New(time.Date(year, month, day, 11, 20, 0, 0, timeutil.Timezone(pbb.COUNTRY_JP))),
		},
	}

	request := &pb.CreateTimesheetRequest{
		StaffId:               staffID,
		TimesheetDate:         timestamppb.New(timesheetDate),
		LocationId:            locationIDs[0],
		Remark:                timesheetRemark,
		ListOtherWorkingHours: listOWHs,
	}

	if !isForCurrentUserID {
		request.StaffId, err = getStaffIDDifferenceCurrentUserID(ctx, staffID)
		if err != nil {
			return nil, err
		}
	}

	return request, nil
}
