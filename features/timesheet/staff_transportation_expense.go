package timesheet

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
)

func (s *Suite) userUpsertStaffTransportationExpense(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Request == nil {
		stepState.Request = &pb.UpsertStaffTransportationExpenseRequest{}
	}
	stepState.RequestSentAt = time.Now()
	stepState.Response, stepState.ResponseErr =
		pb.NewStaffTransportationExpenseServiceClient(s.TimesheetConn).UpsertStaffTransportationExpense(contextWithToken(ctx), stepState.Request.(*pb.UpsertStaffTransportationExpenseRequest))
	return StepStateToContext(ctx, stepState), nil
}

func removeFromStaffTransportExpensesSlicesChangesOrder(req []*pb.StaffTransportationExpenseRequest, elementIndex int) []*pb.StaffTransportationExpenseRequest {
	if elementIndex > len(req) {
		return nil
	}
	req[elementIndex] = req[len(req)-1]
	req[len(req)-1] = (*pb.StaffTransportationExpenseRequest)(nil)
	req = req[:len(req)-1]
	return req
}

func (s *Suite) newInsertStaffTransportationExpenseConfig(ctx context.Context, recordNumber int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	staffID, err := getOneStaffIDInDB(ctx, stepState.CurrentUserID, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StaffID = staffID

	locationId, err := getOneLocationIDInDB(ctx, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	listUpdateTransportExpenses := make([]*pb.StaffTransportationExpenseRequest, 0, recordNumber)
	for i := 0; i < recordNumber; i++ {

		transportExpenseTemp := &pb.StaffTransportationExpenseRequest{
			LocationId: locationId,
			Type:       pb.TransportationType_TYPE_BUS,
			From:       "HN",
			To:         "HCM",
			RoundTrip:  true,
			CostAmount: 9,
			Remarks:    "",
		}

		listUpdateTransportExpenses = append(listUpdateTransportExpenses, transportExpenseTemp)
	}

	request := &pb.UpsertStaffTransportationExpenseRequest{
		StaffId:                         staffID,
		ListStaffTransportationExpenses: listUpdateTransportExpenses,
	}
	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newUpdateStaffTransportationExpenseConfig(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	recordNumberToUpdate := 5
	staffID, err := getOneStaffIDInDB(ctx, stepState.CurrentUserID, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StaffID = staffID

	locationId, err := getOneLocationIDInDB(ctx, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	listUpdateTransportExpenses := make([]*pb.StaffTransportationExpenseRequest, 0, recordNumberToUpdate)
	for i := 0; i < recordNumberToUpdate; i++ {
		staffTransportExpense, err := initStaffTransportExpenses(ctx, staffID, locationId, strconv.Itoa(constants.ManabieSchool))

		if err != nil {
			return nil, err
		}

		transportExpenseTemp := &pb.StaffTransportationExpenseRequest{
			Id:         staffTransportExpense.ID.String,
			LocationId: locationId,
			Type:       pb.TransportationType_TYPE_TRAIN,
			From:       "TPHCM",
			To:         "HNVN",
			RoundTrip:  false,
			CostAmount: 12,
			Remarks:    "staff transportation expense remark",
		}

		listUpdateTransportExpenses = append(listUpdateTransportExpenses, transportExpenseTemp)
	}

	request := &pb.UpsertStaffTransportationExpenseRequest{
		StaffId:                         staffID,
		ListStaffTransportationExpenses: listUpdateTransportExpenses,
	}
	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newUpsertStaffTransportationExpenseConfig(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	recordNumberToUpdate := 5
	staffID, err := getOneStaffIDInDB(ctx, stepState.CurrentUserID, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StaffID = staffID

	locationId, err := getOneLocationIDInDB(ctx, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	listUpdateTransportExpenses := make([]*pb.StaffTransportationExpenseRequest, 0, recordNumberToUpdate)

	transportExpenseTemp := &pb.StaffTransportationExpenseRequest{
		LocationId: locationId,
		Type:       pb.TransportationType_TYPE_BUS,
		From:       "DN",
		To:         "QN",
		RoundTrip:  true,
		CostAmount: 22,
		Remarks:    "Upsert staff TE config",
	}

	listUpdateTransportExpenses = append(listUpdateTransportExpenses, transportExpenseTemp)

	transportExpenseTemp2 := &pb.StaffTransportationExpenseRequest{
		LocationId: locationId,
		Type:       pb.TransportationType_TYPE_BUS,
		From:       "TPCT",
		To:         "TPHP",
		RoundTrip:  false,
		CostAmount: 18,
		Remarks:    "Upsert staff TE config",
	}

	listUpdateTransportExpenses = append(listUpdateTransportExpenses, transportExpenseTemp2)

	for i := 0; i < recordNumberToUpdate; i++ {
		staffTransportExpense, err := initStaffTransportExpenses(ctx, staffID, locationId, strconv.Itoa(constants.ManabieSchool))

		if err != nil {
			return nil, err
		}

		transportExpenseTemp := &pb.StaffTransportationExpenseRequest{
			Id:         staffTransportExpense.ID.String,
			LocationId: locationId,
			Type:       pb.TransportationType_TYPE_TRAIN,
			From:       "TPHCM",
			To:         "HNVN",
			RoundTrip:  false,
			CostAmount: 12,
			Remarks:    "staff transportation expense remark",
		}

		listUpdateTransportExpenses = append(listUpdateTransportExpenses, transportExpenseTemp)
	}

	request := &pb.UpsertStaffTransportationExpenseRequest{
		StaffId:                         staffID,
		ListStaffTransportationExpenses: listUpdateTransportExpenses,
	}
	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) newDeleteStaffTransportationExpenseConfig(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	recordNumberToUpdate := 5
	staffID, err := getOneStaffIDInDB(ctx, stepState.CurrentUserID, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))

	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}
	stepState.StaffID = staffID

	locationId, err := getOneLocationIDInDB(ctx, strconv.FormatInt(int64(stepState.CurrentSchoolID), 10))
	if err != nil {
		return StepStateToContext(ctx, stepState), err
	}

	listUpdateTransportExpenses := make([]*pb.StaffTransportationExpenseRequest, 0, recordNumberToUpdate)

	for i := 0; i < recordNumberToUpdate; i++ {
		staffTransportExpense, err := initStaffTransportExpenses(ctx, staffID, locationId, strconv.Itoa(constants.ManabieSchool))

		if err != nil {
			return nil, err
		}

		//only get first 2 records and delete 3 remain records
		if i < 2 {
			transportExpenseTemp := &pb.StaffTransportationExpenseRequest{
				Id:         staffTransportExpense.ID.String,
				LocationId: locationId,
				Type:       pb.TransportationType_TYPE_TRAIN,
				From:       "TPHCM",
				To:         "HNVN",
				RoundTrip:  false,
				CostAmount: 12,
				Remarks:    "staff transportation expense remark",
			}

			listUpdateTransportExpenses = append(listUpdateTransportExpenses, transportExpenseTemp)
		}
	}

	request := &pb.UpsertStaffTransportationExpenseRequest{
		StaffId:                         staffID,
		ListStaffTransportationExpenses: listUpdateTransportExpenses,
	}
	stepState.Request = request

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) verifyStaffTransportationConfigNumberAfterUpsert(ctx context.Context, recordNumber int) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	if stepState.ResponseErr != nil {
		return ctx, stepState.ResponseErr
	}
	if stepState.Response != nil {
		if !stepState.Response.(*pb.UpsertStaffTransportationExpenseResponse).Success {
			return ctx, fmt.Errorf("error cannot upsert staff transportation expense config record")
		}

		currentConfigRecordNumber, err := s.countStaffTransportationConfigValue(ctx)
		if err != nil {
			return StepStateToContext(ctx, stepState), err
		}

		if currentConfigRecordNumber != recordNumber {
			return ctx, fmt.Errorf("number staff transportation expense config record is wrong")
		}
	}

	return StepStateToContext(ctx, stepState), nil
}

func (s *Suite) countStaffTransportationConfigValue(ctx context.Context) (int, error) {
	stepState := StepStateFromContext(ctx)
	var count int

	stmt := `
		SELECT
			count(staff_id)
		FROM
			staff_transportation_expense
		WHERE
			staff_id = $1
		AND
			deleted_at IS NULL
		`
	err := s.TimesheetDB.QueryRow(ctx, stmt, stepState.StaffID).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Suite) removeAllStaffTransportationExpense(ctx context.Context) (context.Context, error) {
	stepState := StepStateFromContext(ctx)

	stmt := `
		DELETE FROM
			staff_transportation_expense
		WHERE
			staff_id = $1`
	_, err := s.TimesheetDB.Exec(ctx, stmt, stepState.StaffID)
	if err != nil {
		return ctx, err
	}

	return StepStateToContext(ctx, stepState), nil
}
