package timesheet

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/internal/golibs/constants"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func removeFromTransportExpensesSlicesChangesOrder(req []*pb.TransportationExpensesRequest, elementIndex int) []*pb.TransportationExpensesRequest {
	if elementIndex > len(req) {
		return nil
	}
	req[elementIndex] = req[len(req)-1]
	req[len(req)-1] = (*pb.TransportationExpensesRequest)(nil)
	req = req[:len(req)-1]
	return req
}

func (s *Suite) buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx context.Context, numOfTransportExpenses int, status string) (*pb.UpdateTimesheetRequest, error) {
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
	listUpdateTransportExpenses := make([]*pb.TransportationExpensesRequest, 0, numOfTransportExpenses)
	for i := 0; i < numOfTransportExpenses; i++ {
		transportExpense, err := initTransportExpenses(ctx, timesheetID, strconv.Itoa(constants.ManabieSchool))
		if err != nil {
			return nil, err
		}

		transportExpenseTemp := &pb.TransportationExpensesRequest{
			TransportationExpenseId: transportExpense.TransportationExpenseID.String,
			Type:                    pb.TransportationType_TYPE_BUS,
			From:                    transportExpense.TransportationFrom.String,
			To:                      transportExpense.TransportationTo.String,
			RoundTrip:               transportExpense.RoundTrip.Bool,
			Amount:                  transportExpense.CostAmount.Int,
			Remarks:                 transportExpense.Remarks.String,
		}

		listUpdateTransportExpenses = append(listUpdateTransportExpenses, transportExpenseTemp)
	}

	owhs := &pb.OtherWorkingHoursRequest{
		TimesheetConfigId: initTimesheetConfigID1,
		StartTime:         timestamppb.Now(),
		EndTime:           timestamppb.Now(),
	}

	request := &pb.UpdateTimesheetRequest{
		TimesheetId:                timesheetID,
		ListTransportationExpenses: listUpdateTransportExpenses,
		ListOtherWorkingHours: []*pb.OtherWorkingHoursRequest{
			owhs,
		},
	}

	return request, nil
}

func (s *Suite) newUpdateTimesheetWithTransportExpensesDataForCurrentStaff(ctx context.Context, action string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	var (
		err     error
		request *pb.UpdateTimesheetRequest
	)

	transportExpense1 := &pb.TransportationExpensesRequest{
		Type:      pb.TransportationType_TYPE_BUS,
		From:      randStringBytes(10),
		To:        randStringBytes(10),
		RoundTrip: true,
		Amount:    10,
		Remarks:   randStringBytes(10),
	}
	transportExpense2 := &pb.TransportationExpensesRequest{
		Type:      pb.TransportationType_TYPE_BUS,
		From:      randStringBytes(10),
		To:        randStringBytes(10),
		RoundTrip: false,
		Amount:    12,
		Remarks:   randStringBytes(10),
	}
	switch action {
	case "insert":
		currentTransportExpensesListLen := 0
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListTransportationExpenses = append(request.ListTransportationExpenses, transportExpense1)

	case "update":
		currentTransportExpensesListLen := 1
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListTransportationExpenses[0].Remarks = randStringBytes(15)

	case "delete":
		currentTransportExpensesListLen := 1
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.ListTransportationExpenses = removeFromTransportExpensesSlicesChangesOrder(request.ListTransportationExpenses, 0)

	case "insert,delete":
		currentTransportExpensesListLen := 5
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.ListTransportationExpenses = removeFromTransportExpensesSlicesChangesOrder(request.ListTransportationExpenses, 0)

		request.ListTransportationExpenses = append(request.ListTransportationExpenses, transportExpense1)

	case "insert,update":
		currentTransportExpensesListLen := 1
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListTransportationExpenses[0].Remarks = randStringBytes(15)

		request.ListTransportationExpenses = append(request.ListTransportationExpenses, transportExpense1)

	case "update,delete":
		currentTransportExpensesListLen := 2
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListTransportationExpenses[0].Remarks = randStringBytes(15)

		request.ListTransportationExpenses = removeFromTransportExpensesSlicesChangesOrder(request.ListTransportationExpenses, 1)

	case "insert,update,delete":
		currentTransportExpensesListLen := 5
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}
		request.ListTransportationExpenses[0].Remarks = randStringBytes(15)

		request.ListTransportationExpenses = removeFromTransportExpensesSlicesChangesOrder(request.ListTransportationExpenses, 1)

		request.ListTransportationExpenses = append(request.ListTransportationExpenses, transportExpense1)
	case "have-10,insert-2,delete-1":
		currentTransportExpensesListLen := 10
		request, err = s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		request.ListTransportationExpenses = removeFromTransportExpensesSlicesChangesOrder(request.ListTransportationExpenses, 0)

		request.ListTransportationExpenses = append(request.ListTransportationExpenses, transportExpense1, transportExpense2)

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

func (s *Suite) userUpdateInvalidTransportExpensesArgsForTimesheet(ctx context.Context, invalidArgCase string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	currentTransportExpensesListLen := 5
	request, err := s.buildDataForUpdateTimesheetWithTransportExpensesRequest(ctx, currentTransportExpensesListLen, pb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String())
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	switch invalidArgCase {
	case "transportation expenses list over 10":

		listTransportExpenseTemp := make([]*pb.TransportationExpensesRequest, 0, listTransportExpensesLimit+1)
		for i := 0; i <= (listTransportExpensesLimit + 1); i++ { // +1 over limit
			transportExpenses := &pb.TransportationExpensesRequest{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HAI PHONG",
				To:        "DA NANG",
				Amount:    31,
				RoundTrip: true,
			}

			listTransportExpenseTemp = append(listTransportExpenseTemp, transportExpenses)
		}
		request.ListTransportationExpenses = listTransportExpenseTemp
	case "transportation type invalid":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_INVALID,
				From:      "HAI PHONG",
				To:        "DA NANG",
				Amount:    10,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "transportation expenses from null":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "",
				To:        "DA NANG",
				Amount:    31,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp
	case "transportation expenses from > 100 character":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      randStringBytes(transportExpenseFromToLimit + 1),
				To:        "DA NANG",
				Amount:    10,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "transportation expenses to null":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "",
				To:        "DA NANG",
				Amount:    31,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp
	case "transportation expenses to > 100 character":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HAI PHONG",
				To:        randStringBytes(transportExpenseFromToLimit + 1),
				Amount:    10,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "cost amount is empty":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HAI PHONG",
				To:        "DA NANG",
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp
	case "cost amount is smaller than 0":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HAI PHONG",
				To:        "DA NANG",
				Amount:    -3,
				RoundTrip: true,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "round trip is empty":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:   pb.TransportationType_TYPE_TRAIN,
				From:   "HAI PHONG",
				To:     "DA NANG",
				Amount: 10,
			},
		}
		request.ListTransportationExpenses = listTransportExpenseTemp

	case "transportation expenses remarks > 100 character":
		listTransportExpenseTemp := []*pb.TransportationExpensesRequest{
			{
				Type:      pb.TransportationType_TYPE_BUS,
				From:      "HAI PHONG",
				To:        "DA NANG",
				Amount:    15,
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
