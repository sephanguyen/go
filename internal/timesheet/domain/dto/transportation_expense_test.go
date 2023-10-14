package dto

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestTransportationExpenses_ToEntity(t *testing.T) {
	var (
		transportationExpenses = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HCM",
			CostAmount:         5,
			RoundTrip:          true,
			Remarks:            "",
		}

		transportationExpensesExpect = &entity.TransportationExpense{
			TransportationExpenseID: database.Text("id_1"),
			TimesheetID:             database.Text("ts_1"),
			TransportationType:      database.Text("BUS"),
			TransportationFrom:      database.Text("HN"),
			TransportationTo:        database.Text("HCM"),
			CostAmount:              database.Int4(5),
			RoundTrip:               database.Bool(true),
			Remarks:                 database.Text(""),
			CreatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
			UpdatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
			DeletedAt:               pgtype.Timestamptz{Status: pgtype.Null},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "convert to entity success",
			request:      transportationExpenses,
			expectedResp: transportationExpensesExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*TransportationExpenses).ToEntity()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTransportationExpenses_Validate(t *testing.T) {
	t.Parallel()

	var (
		transportationExpensesWithEmptyType = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
		}

		transportationExpensesWithEmptyFrom = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
			TransportationType: "BUS",
		}

		transportationExpensesWithEmptyTo = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
		}

		transportationExpensesWithCostSmallThanZero = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HN",
			CostAmount:         -1,
		}

		transportationExpensesWithRemarksOverLimit = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   "HN",
			CostAmount:         1,
			Remarks:            strings.Repeat("a", 101),
		}

		transportationExpensesWithTransportationFromOverLimit = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
			TransportationType: "BUS",
			TransportationFrom: strings.Repeat("a", 101),
			TransportationTo:   "HN",
			CostAmount:         1,
			Remarks:            "",
		}

		transportationExpensesWithTransportationToOverLimit = &TransportationExpenses{
			TransportExpenseID: "id_1",
			TimesheetID:        "ts_1",
			TransportationType: "BUS",
			TransportationFrom: "HN",
			TransportationTo:   strings.Repeat("a", 101),
			CostAmount:         1,
			Remarks:            "",
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{

		{
			name:         "fail with empty type",
			request:      transportationExpensesWithEmptyType,
			expectedResp: fmt.Errorf("transportation type must not be empty"),
		},
		{
			name:         "fail with empty from",
			request:      transportationExpensesWithEmptyFrom,
			expectedResp: fmt.Errorf("transportation from must not be nil"),
		},
		{
			name:         "fail with empty to",
			request:      transportationExpensesWithEmptyTo,
			expectedResp: fmt.Errorf("transportation to must not be nil"),
		},
		{
			name:         "fail with cost small than error",
			request:      transportationExpensesWithCostSmallThanZero,
			expectedResp: fmt.Errorf("transportation cost amount must be greater than 0"),
		},
		{
			name:         "fail with remarks over limit",
			request:      transportationExpensesWithRemarksOverLimit,
			expectedResp: fmt.Errorf("transportation remarks must limit to 100 characters"),
		},
		{
			name:         "fail with transportation from over limit",
			request:      transportationExpensesWithTransportationFromOverLimit,
			expectedResp: fmt.Errorf("transportation from must limit to 100 characters"),
		},
		{
			name:         "fail with transportation to over limit",
			request:      transportationExpensesWithTransportationToOverLimit,
			expectedResp: fmt.Errorf("transportation to must limit to 100 characters"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*TransportationExpenses).Validate()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestListTransportationExpenses_ToEntities(t *testing.T) {
	var (
		transportationExpenses = &ListTransportationExpenses{
			{
				TransportExpenseID: "id_1",
				TimesheetID:        "ts_1",
				TransportationType: "BUS",
				TransportationFrom: "HN",
				TransportationTo:   "HCM",
				CostAmount:         5,
				RoundTrip:          true,
				Remarks:            "",
			},
		}

		transportationExpensesExpect = []*entity.TransportationExpense{
			{
				TransportationExpenseID: database.Text("id_1"),
				TimesheetID:             database.Text("ts_1"),
				TransportationType:      database.Text("BUS"),
				TransportationFrom:      database.Text("HN"),
				TransportationTo:        database.Text("HCM"),
				CostAmount:              database.Int4(5),
				RoundTrip:               database.Bool(true),
				Remarks:                 database.Text(""),
				CreatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
				UpdatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
				DeletedAt:               pgtype.Timestamptz{Status: pgtype.Null},
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "convert to entities success",
			request:      transportationExpenses,
			expectedResp: transportationExpensesExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*ListTransportationExpenses).ToEntities()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestListTransportationExpenses_Validate(t *testing.T) {
	var (
		transportationExpenses = &ListTransportationExpenses{
			{
				TransportExpenseID: "id_1",
				TimesheetID:        "ts_1",
				TransportationType: "BUS",
				TransportationFrom: "HN",
				TransportationTo:   "HCM",
				CostAmount:         5,
				RoundTrip:          true,
				Remarks:            "",
			},
		}
		transportationExpensesOverLimit = &ListTransportationExpenses{
			{
				TransportExpenseID: "id_1",
			},
			{
				TransportExpenseID: "id_2",
			},
			{
				TransportExpenseID: "id_3",
			},
			{
				TransportExpenseID: "id_4",
			},
			{
				TransportExpenseID: "id_5",
			},
			{
				TransportExpenseID: "id_6",
			},
			{
				TransportExpenseID: "id_7",
			},
			{
				TransportExpenseID: "id_8",
			},
			{
				TransportExpenseID: "id_9",
			},
			{
				TransportExpenseID: "id_10",
			},
			{
				TransportExpenseID: "id_11",
			},
		}
		transportationExpensesHasError = &ListTransportationExpenses{
			{
				TransportExpenseID: "id_1",
				TimesheetID:        "ts_1",
				TransportationType: "BUS",
				TransportationFrom: "HN",
				TransportationTo:   "HCM",
				CostAmount:         -1,
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
			name:         "list transportation expenses validate success",
			request:      transportationExpenses,
			expectedResp: nil,
		},
		{
			name:         "list transportation expenses over limit",
			request:      transportationExpensesOverLimit,
			expectedResp: fmt.Errorf("list transportation expenses must be limit to 10 rows"),
		},
		{
			name:         "list transportation expenses has error",
			request:      transportationExpensesHasError,
			expectedResp: fmt.Errorf("transportation cost amount must be greater than 0"),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*ListTransportationExpenses).Validate()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewListTransportExpensesFromRPCRequest(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name: "new list transport expenses from rpc request",
			request: []*pb.TransportationExpensesRequest{
				{
					TransportationExpenseId: "id_1",
					Type:                    pb.TransportationType_TYPE_BUS,
					From:                    "HN",
					To:                      "HCM",
					Amount:                  5,
					RoundTrip:               true,
					Remarks:                 "",
				},
			},
			expectedResp: ListTransportationExpenses{
				{
					TransportExpenseID: "id_1",
					TimesheetID:        "ts_1",
					TransportationType: "TYPE_BUS",
					TransportationFrom: "HN",
					TransportationTo:   "HCM",
					CostAmount:         5,
					RoundTrip:          true,
					Remarks:            "",
				},
			},
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewListTransportExpensesFromRPCRequest("ts_1", testcase.request.([]*pb.TransportationExpensesRequest))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestNewListTransportExpensesFromEntities(t *testing.T) {
	t.Parallel()
	var (
		transportationExpensesExpect = ListTransportationExpenses{
			{
				TransportExpenseID: "id_1",
				TimesheetID:        "ts_1",
				TransportationType: "BUS",
				TransportationFrom: "HN",
				TransportationTo:   "HCM",
				CostAmount:         5,
				Remarks:            "",
			},
		}

		transportationExpensesRequest = []*entity.TransportationExpense{
			{
				TransportationExpenseID: database.Text("id_1"),
				TimesheetID:             database.Text("ts_1"),
				TransportationType:      database.Text("BUS"),
				TransportationFrom:      database.Text("HN"),
				TransportationTo:        database.Text("HCM"),
				CostAmount:              database.Int4(5),
				Remarks:                 database.Text(""),
				CreatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
				UpdatedAt:               pgtype.Timestamptz{Status: pgtype.Null},
				DeletedAt:               pgtype.Timestamptz{Status: pgtype.Null},
			},
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new list transport expenses from entities",
			request:      transportationExpensesRequest,
			expectedResp: transportationExpensesExpect,
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewListTransportExpensesFromEntities(testcase.request.([]*entity.TransportationExpense))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
