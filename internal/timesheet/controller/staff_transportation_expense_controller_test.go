package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	mock_staff_transport_expense_services "github.com/manabie-com/backend/mock/timesheet/service/staff_transportation_expense"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStaffTransportationExpenseController_UpsertStaffTransportationExpenseConfig(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	staffTransportationExpenseService := new(mock_staff_transport_expense_services.MockStaffTransportationExpenseServiceImpl)

	ctl := &StaffTransportationExpenseController{
		StaffTransportationExpenseService: staffTransportationExpenseService,
	}

	staffTransportationExpenseConfig := &pb.StaffTransportationExpenseRequest{
		Id:         "transport-expense-id",
		LocationId: "location-transport-expense-id",
		Type:       pb.TransportationType_TYPE_BUS,
		From:       "HCM",
		To:         "DN",
		CostAmount: 9900,
		RoundTrip:  true,
		Remarks:    "",
	}

	staffTransportationExpenseConfigInvalid := &pb.StaffTransportationExpenseRequest{
		LocationId: "location-transport-expense-id",
		Type:       pb.TransportationType_TYPE_BUS,
		From:       "HCM",
		To:         "DN",
		CostAmount: 9900,
		RoundTrip:  true,
		Remarks:    "",
	}

	ListStaffTransportationExpenses := []*pb.StaffTransportationExpenseRequest{
		staffTransportationExpenseConfig,
	}

	ListStaffTransportationExpensesInvalid := []*pb.StaffTransportationExpenseRequest{
		staffTransportationExpenseConfigInvalid,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.UpsertStaffTransportationExpenseRequest{
				StaffId:                         "staff_id",
				ListStaffTransportationExpenses: ListStaffTransportationExpenses,
			},
			expectedErr:  nil,
			expectedResp: &pb.UpsertStaffTransportationExpenseResponse{Success: true},
			setup: func(ctx context.Context) {
				staffTransportationExpenseService.On("UpsertConfig", ctx, mock.Anything, mock.Anything).
					Return(nil).Once()
			},
		},
		{
			name: "error case invalid request",
			ctx:  ctx,
			req: &pb.UpsertStaffTransportationExpenseRequest{
				ListStaffTransportationExpenses: ListStaffTransportationExpensesInvalid,
			},
			expectedErr:  status.Error(codes.InvalidArgument, "staff id must not be empty"),
			expectedResp: (*pb.UpsertStaffTransportationExpenseResponse)(nil),
			setup:        func(ctx context.Context) {},
		},

		{
			name: "error case upsert staff transportation expense config failed",
			ctx:  ctx,
			req: &pb.UpsertStaffTransportationExpenseRequest{
				StaffId:                         "staff_id",
				ListStaffTransportationExpenses: ListStaffTransportationExpenses,
			},
			expectedErr:  status.Error(codes.Internal, fmt.Sprintf("err upsert staff transportation expense")),
			expectedResp: (*pb.UpsertStaffTransportationExpenseResponse)(nil),
			setup: func(ctx context.Context) {
				staffTransportationExpenseService.On("UpsertConfig", ctx, mock.Anything, mock.Anything).Return(status.Error(codes.Internal, fmt.Sprintf("err upsert staff transportation expense"))).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.UpsertStaffTransportationExpenseRequest)
			resp, err := ctl.UpsertStaffTransportationExpense(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
