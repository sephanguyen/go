package timesheet

import (
	"context"
	"fmt"
	"time"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Suite) newTimesheetWithTranportExpensesData(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var err error

	stepState.Request, err = buildCreateTimesheetWithTransportExpensesRequest(ctx, stepState.CurrentUserID, true /*isForCurrentUserID*/)
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newInvalidTransportExpenseDataRequest(ctx context.Context, invalidArgCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	timesheetDate := generateRandomDate()

	request := &pb.CreateTimesheetRequest{
		StaffId:       stepState.CurrentUserID,
		TimesheetDate: timestamppb.New(timesheetDate),
		LocationId:    locationIDs[0],
		Remark:        timesheetRemark,
	}

	switch invalidArgCase {
	case "transportation expenses list over 10":

		listTransportExpenseTemp := make([]*pb.TransportationExpensesRequest, 0, listTransportExpensesLimit+1)
		for i := 0; i <= (listTransportExpensesLimit + 1); i++ { // +1 over limit
			transportExpenses := &pb.TransportationExpensesRequest{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "Ha Noi",
				To:        "TP HCM",
				Amount:    12,
				RoundTrip: true,
			}

			listTransportExpenseTemp = append(listTransportExpenseTemp, transportExpenses)
		}
		request.ListTransportationExpenses = listTransportExpenseTemp
	case "transportation type invalid":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_INVALID,
				From:      "Ha Noi",
				To:        "TP HCM",
				Amount:    20,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "transportation expenses from null":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "",
				To:        "TP HCM",
				Amount:    12,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp
	case "transportation expenses from > 100 character":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      randStringBytes(transportExpenseFromToLimit + 1),
				To:        "TP HCM",
				Amount:    15,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "transportation expenses to null":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "",
				To:        "TP HCM",
				Amount:    15,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp
	case "transportation expenses to > 100 character":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HA NOI",
				To:        randStringBytes(transportExpenseFromToLimit + 1),
				Amount:    12,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "cost amount is empty":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HA NOI",
				To:        "TP HCM",
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "cost amount is smaller than 0":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HA NOI",
				To:        "TP HCM",
				Amount:    -1,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "round trip is empty":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:   pb.TransportationType_TYPE_TRAIN,
				From:   "HA NOI",
				To:     "TP HCM",
				Amount: 0,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "transportation expenses remarks > 100 character":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HA NOI",
				To:        "TP HCM",
				Amount:    0,
				RoundTrip: true,
				Remarks:   randStringBytes(transportExpenseRemarksLimit + 1),
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	default:
		return StepStateToContext(ctx, stepState),
			fmt.Errorf("invalid example argument: %v", invalidArgCase)
	}

	stepState.Request = request
	return StepStateToContext(ctx, stepState), nil
}

func buildCreateTimesheetWithTransportExpensesRequest(ctx context.Context, staffID string, isForCurrentUserID bool) (*pb.CreateTimesheetRequest, error) {
	var err error
	timesheetDate := generateRandomDate()

	listTransportExpenses := []*pb.TransportationExpensesRequest{
		{
			Type:      pb.TransportationType_TYPE_BUS,
			From:      "Ha Noi",
			To:        "TP HCM",
			Amount:    12,
			RoundTrip: true,
			Remarks:   transportExpenseRemark,
		},
		{
			Type:      pb.TransportationType_TYPE_BUS,
			From:      "Da Nang",
			To:        "Quy Nhon",
			Amount:    19,
			RoundTrip: false,
			Remarks:   transportExpenseRemark,
		},
		{
			Type:      pb.TransportationType_TYPE_BUS,
			From:      "Nha Trang",
			To:        "Phu Quoc",
			Amount:    10,
			RoundTrip: false,
			Remarks:   transportExpenseRemark,
		},
	}

	request := &pb.CreateTimesheetRequest{
		StaffId:                    staffID,
		TimesheetDate:              timestamppb.New(timesheetDate),
		LocationId:                 locationIDs[0],
		Remark:                     timesheetRemark,
		ListTransportationExpenses: listTransportExpenses,
	}

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
