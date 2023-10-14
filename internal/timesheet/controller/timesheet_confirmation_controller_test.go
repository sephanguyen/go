package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	mock_ts_cod_services "github.com/manabie-com/backend/mock/timesheet/service/timesheet_confirmation"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestTimesheetConfirmationController_GetConfirmationPeriodByDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tsConfirmationPeriodService := new(mock_ts_cod_services.MockConfirmationWindowServiceImpl)

	ctl := &TimesheetConfirmationController{
		TimesheetConfirmationWindowService: tsConfirmationPeriodService,
	}
	var startDateExpect *timestamp.Timestamp = timestamppb.Now()
	timeNow := time.Now()

	tsConfirmationPeriod := pb.TimesheetConfirmationPeriod{
		Id:        "period-id",
		StartDate: timestamppb.New(timeNow),
		EndDate:   timestamppb.New(timeNow),
	}

	TsConfimPeriodDto := &dto.TimesheetConfirmationPeriod{
		ID:        "period-id",
		StartDate: timeNow,
		EndDate:   timeNow,
	}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.GetTimesheetConfirmationPeriodByDateRequest{
				Date: startDateExpect,
			},
			expectedErr: nil,
			expectedResp: &pb.GetTimesheetConfirmationPeriodByDateResponse{
				TimesheetConfirmationPeriod: &tsConfirmationPeriod,
			},
			setup: func(ctx context.Context) {
				tsConfirmationPeriodService.On("GetPeriod", ctx, mock.Anything).
					Return(TsConfimPeriodDto, nil).Once()
			},
		},
		{
			name:         "error case invalid request",
			ctx:          ctx,
			req:          &pb.GetTimesheetConfirmationPeriodByDateRequest{},
			expectedErr:  status.Error(codes.InvalidArgument, "date must not be empty"),
			expectedResp: (*pb.GetTimesheetConfirmationPeriodByDateResponse)(nil),
			setup:        func(ctx context.Context) {},
		},

		{
			name: "error case get period failed",
			ctx:  ctx,
			req: &pb.GetTimesheetConfirmationPeriodByDateRequest{
				Date: startDateExpect,
			},
			expectedErr:  status.Error(codes.Internal, "err get period"),
			expectedResp: (*pb.GetTimesheetConfirmationPeriodByDateResponse)(nil),
			setup: func(ctx context.Context) {
				tsConfirmationPeriodService.On("GetPeriod", ctx, mock.Anything).
					Return(nil, status.Error(codes.Internal, fmt.Sprintf("err get period"))).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.GetTimesheetConfirmationPeriodByDateRequest)
			resp, err := ctl.GetConfirmationPeriodByDate(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}

func TestTimesheetConfirmationController_ConfirmTimesheet(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	tsConfirmationService := new(mock_ts_cod_services.MockConfirmationWindowServiceImpl)

	ctl := &TimesheetConfirmationController{
		TimesheetConfirmationWindowService: tsConfirmationService,
	}

	confirmRequest := &pb.ConfirmTimesheetWithLocationRequest{
		PeriodId:    "period-1",
		LocationIds: []string{"location_1", "location_2"},
	}

	testCases := []TestCase{
		{
			name:        "happy case",
			ctx:         ctx,
			req:         confirmRequest,
			expectedErr: nil,
			expectedResp: &pb.ConfirmTimesheetWithLocationResponse{
				Success: true,
			},
			setup: func(ctx context.Context) {
				tsConfirmationService.On("ConfirmPeriod", ctx, mock.Anything).
					Return(nil).Once()
			},
		},
		{
			name:         "error case invalid request",
			ctx:          ctx,
			req:          &pb.ConfirmTimesheetWithLocationRequest{},
			expectedErr:  status.Error(codes.InvalidArgument, "location ids cannot be empty"),
			expectedResp: (*pb.ConfirmTimesheetWithLocationResponse)(nil),
			setup:        func(ctx context.Context) {},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.ConfirmTimesheetWithLocationRequest)
			resp, err := ctl.ConfirmTimesheet(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
