package dto

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestNewStaffTransportExpensesFromRPCRequest(t *testing.T) {
	t.Parallel()

	var (
		listStaffTransportationExpenseReq = &pb.StaffTransportationExpenseRequest{
			Id:         "id_1",
			LocationId: "location_1",
			Type:       pb.TransportationType_TYPE_BUS,
			From:       "HN",
			To:         "HCM",
			CostAmount: 5,
			RoundTrip:  true,
			Remarks:    "",
		}
		listStaffTransportationExpensesExpect = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "TYPE_BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HCM",
			CostAmount:         5,
			RoundTrip:          true,
			Remarks:            "",
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new staff transportation expense from rpc request",
			request:      listStaffTransportationExpenseReq,
			expectedResp: listStaffTransportationExpensesExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewStaffTransportExpenseFromRPCRequest("staff_1", testcase.request.(*pb.StaffTransportationExpenseRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewStaffTransportExpensesFromEntity(t *testing.T) {
	t.Parallel()

	var (
		staffTransportationExpensesExpect = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HCM",
			CostAmount:         5,
			RoundTrip:          true,
			Remarks:            "",
		}
		staffTransportationExpensesExpectDeleted = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HCM",
			CostAmount:         5,
			RoundTrip:          true,
			Remarks:            "",
			IsDeleted:          true,
		}

		staffTransportationExpenses = &entity.StaffTransportationExpense{
			ID:                 database.Text("id_1"),
			StaffID:            database.Text("staff_1"),
			LocationID:         database.Text("location_1"),
			TransportationType: database.Text("BUS"),
			TransportationFrom: database.Text("HN"),
			TransportationTo:   database.Text("HCM"),
			CostAmount:         database.Int4(5),
			RoundTrip:          database.Bool(true),
			Remarks:            database.Text(""),
			CreatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:          pgtype.Timestamptz{Status: pgtype.Null},
		}
		staffTransportationExpensesDeleted = &entity.StaffTransportationExpense{
			ID:                 database.Text("id_1"),
			StaffID:            database.Text("staff_1"),
			LocationID:         database.Text("location_1"),
			TransportationType: database.Text("BUS"),
			TransportationFrom: database.Text("HN"),
			TransportationTo:   database.Text("HCM"),
			CostAmount:         database.Int4(5),
			RoundTrip:          database.Bool(true),
			Remarks:            database.Text(""),
			CreatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:          database.Timestamptz(time.Now()),
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new staff transportation expense from entity success",
			request:      staffTransportationExpenses,
			expectedResp: staffTransportationExpensesExpect,
		},
		{
			name:         "new staff transportation expense from entity set deleted",
			request:      staffTransportationExpensesDeleted,
			expectedResp: staffTransportationExpensesExpectDeleted,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewStaffTransportExpensesFromEntity(testcase.request.(*entity.StaffTransportationExpense))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestStaffTransportationExpenses_IsEqual(t *testing.T) {
	t.Parallel()

	var (
		staffTransportationExpensesA = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HCM",
			CostAmount:         5,
			RoundTrip:          true,
			Remarks:            "",
		}
		staffTransportationExpensesB = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HCM",
			CostAmount:         5,
			RoundTrip:          true,
			Remarks:            "",
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		toCompare    interface{}
		expectedResp interface{}
	}{
		{
			name:         "is equal success",
			request:      staffTransportationExpensesA,
			toCompare:    staffTransportationExpensesB,
			expectedResp: true,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*StaffTransportationExpenses).IsEqual(testcase.toCompare.(*StaffTransportationExpenses))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestStaffTransportationExpenses_ToEntity(t *testing.T) {
	t.Parallel()

	var (
		staffTransportationExpenses = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HCM",
			CostAmount:         5,
			RoundTrip:          true,
			Remarks:            "",
		}

		staffTransportationExpensesExpect = &entity.StaffTransportationExpense{
			ID:                 database.Text("id_1"),
			StaffID:            database.Text("staff_1"),
			LocationID:         database.Text("location_1"),
			TransportationType: database.Text("BUS"),
			TransportationFrom: database.Text("HN"),
			TransportationTo:   database.Text("HCM"),
			CostAmount:         database.Int4(5),
			RoundTrip:          database.Bool(true),
			Remarks:            database.Text(""),
			CreatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:          pgtype.Timestamptz{Status: pgtype.Null},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "convert to entity success",
			request:      staffTransportationExpenses,
			expectedResp: staffTransportationExpensesExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*StaffTransportationExpenses).ToEntity()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestStaffTransportationExpenses_ValidateUpsertInfo(t *testing.T) {
	t.Parallel()

	var (
		staffTransportationExpensesWithEmptyLocationId = &StaffTransportationExpenses{
			ID:         "id_1",
			StaffID:    "staff_1",
			LocationID: "",
		}

		staffTransportationExpensesWithEmptyType = &StaffTransportationExpenses{
			ID:         "id_1",
			StaffID:    "staff_1",
			LocationID: "location_1",
		}

		staffTransportationExpensesWithEmptyFrom = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
		}

		staffTransportationExpensesWithEmptyTo = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
		}

		staffTransportationExpensesWithCostSmallThanZero = &StaffTransportationExpenses{
			ID:                 "id_1",
			StaffID:            "staff_1",
			LocationID:         "location_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HN",
			CostAmount:         -1,
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{

		{
			name:         "fail with empty location id",
			request:      staffTransportationExpensesWithEmptyLocationId,
			expectedResp: fmt.Errorf("location id must not be empty"),
		},
		{
			name:         "fail with empty type",
			request:      staffTransportationExpensesWithEmptyType,
			expectedResp: fmt.Errorf("transportation type must not be empty"),
		},
		{
			name:         "fail with empty from",
			request:      staffTransportationExpensesWithEmptyFrom,
			expectedResp: fmt.Errorf("transportation from must not be nil"),
		},
		{
			name:         "fail with empty to",
			request:      staffTransportationExpensesWithEmptyTo,
			expectedResp: fmt.Errorf("transportation to must not be nil"),
		},
		{
			name:         "fail with cost small than error",
			request:      staffTransportationExpensesWithCostSmallThanZero,
			expectedResp: fmt.Errorf("transportation cost amount must be greater than 0"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*StaffTransportationExpenses).ValidateUpsertInfo()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestListStaffTransportationExpenses_ValidateUpsertInfo(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name: "happy case",
			request: &ListStaffTransportationExpenses{
				{
					ID:                 "id_1",
					StaffID:            "staff_1",
					LocationID:         "location_1",
					TransportationType: "BUS",
					TransportationFrom: "HN",
					TransportationTo:   "HN",
					CostAmount:         1,
				},
			},
			expectedResp: nil,
		},
		{
			name: "error case list staff transportation expense length over limit",
			request: &ListStaffTransportationExpenses{
				{
					ID: "id_1",
				},
				{
					ID: "id_2",
				},
				{
					ID: "id_3",
				},
				{
					ID: "id_4",
				},
				{
					ID: "id_5",
				},
				{
					ID: "id_6",
				},
				{
					ID: "id_7",
				},
				{
					ID: "id_8",
				},
				{
					ID: "id_9",
				},
				{
					ID: "id_10",
				},
				{
					ID: "id_11",
				},
			},
			expectedResp: fmt.Errorf("list staff transportation expenses config must be limit to 10 rows"),
		},
		{
			name: "error case list staff transportation expense validation error",
			request: &ListStaffTransportationExpenses{
				{
					ID:                 "id_1",
					StaffID:            "staff_1",
					LocationID:         "location_1",
					TransportationType: "BUS",
					TransportationFrom: "HN",
					TransportationTo:   "HN",
					CostAmount:         0,
				},
			},
			expectedResp: fmt.Errorf("transportation cost amount must be greater than 0"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*ListStaffTransportationExpenses).ValidateUpsertInfo()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestListStaffTransportationExpenses_ToEntities(t *testing.T) {
	t.Parallel()

	var (
		listStaffTransportationExpensesDto = &ListStaffTransportationExpenses{
			{
				ID:                 "id_1",
				StaffID:            "staff_1",
				LocationID:         "location_1",
				TransportationType: "BUS",
				TransportationFrom: "HN",
				TransportationTo:   "HCM",
				CostAmount:         5,
				RoundTrip:          true,
				Remarks:            "",
			},
		}
		listStaffTransportationExpensesE = []*entity.StaffTransportationExpense{
			{
				ID:                 database.Text("id_1"),
				StaffID:            database.Text("staff_1"),
				LocationID:         database.Text("location_1"),
				TransportationType: database.Text("BUS"),
				TransportationFrom: database.Text("HN"),
				TransportationTo:   database.Text("HCM"),
				CostAmount:         database.Int4(5),
				RoundTrip:          database.Bool(true),
				Remarks:            database.Text(""),
				CreatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
				UpdatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
				DeletedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			},
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "list staff transportation expenses to entities success",
			request:      listStaffTransportationExpensesDto,
			expectedResp: listStaffTransportationExpensesE,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*ListStaffTransportationExpenses).ToEntities()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewListStaffTransportExpensesFromRPCRequest(t *testing.T) {
	t.Parallel()

	var (
		listStaffTransportationExpenseReq = []*pb.StaffTransportationExpenseRequest{
			{
				Id:         "id_1",
				LocationId: "location_1",
				Type:       pb.TransportationType_TYPE_BUS,
				From:       "HN",
				To:         "HCM",
				CostAmount: 5,
				RoundTrip:  true,
				Remarks:    "",
			},
		}
		listStaffTransportationExpensesExpect = ListStaffTransportationExpenses{
			{
				ID:                 "id_1",
				StaffID:            "staff_1",
				LocationID:         "location_1",
				TransportationType: "TYPE_BUS",
				TransportationFrom: "HN",
				TransportationTo:   "HCM",
				CostAmount:         5,
				RoundTrip:          true,
				Remarks:            "",
			},
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new list staff transportation expense from rpc request",
			request:      listStaffTransportationExpenseReq,
			expectedResp: listStaffTransportationExpensesExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewListStaffTransportExpensesFromRPCRequest("staff_1", testcase.request.([]*pb.StaffTransportationExpenseRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewListStaffTransportExpensesFromEntities(t *testing.T) {
	t.Parallel()

	var (
		listStaffTransportationExpensesExpect = ListStaffTransportationExpenses{
			{
				ID:                 "id_1",
				StaffID:            "staff_1",
				LocationID:         "location_1",
				TransportationType: "BUS",
				TransportationFrom: "HN",
				TransportationTo:   "HCM",
				CostAmount:         5,
				RoundTrip:          true,
				Remarks:            "",
			},
		}

		listStaffTransportationExpensesE = []*entity.StaffTransportationExpense{
			{
				ID:                 database.Text("id_1"),
				StaffID:            database.Text("staff_1"),
				LocationID:         database.Text("location_1"),
				TransportationType: database.Text("BUS"),
				TransportationFrom: database.Text("HN"),
				TransportationTo:   database.Text("HCM"),
				CostAmount:         database.Int4(5),
				RoundTrip:          database.Bool(true),
				Remarks:            database.Text(""),
				CreatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
				UpdatedAt:          pgtype.Timestamptz{Status: pgtype.Null},
				DeletedAt:          pgtype.Timestamptz{Status: pgtype.Null},
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new list staff transportation expense from entities",
			request:      listStaffTransportationExpensesE,
			expectedResp: listStaffTransportationExpensesExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewListStaffTransportExpensesFromEntities(testcase.request.([]*entity.StaffTransportationExpense))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
