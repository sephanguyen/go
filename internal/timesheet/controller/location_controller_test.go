package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	mock_location_service "github.com/manabie-com/backend/mock/timesheet/service/location"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLocationController_GetConfirmationPeriodByDate(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	locationService := new(mock_location_service.MockLocationServiceImpl)

	ctl := &LocationController{
		LocationService: locationService,
	}

	locationResonse1 := pb.Location{
		LocationId: "location-1",
		Name:       "location 1",
	}

	locationResonse2 := pb.Location{
		LocationId: "location-2",
		Name:       "location 2",
	}

	listLocationResponse := []*pb.Location{&locationResonse1, &locationResonse2}

	locationDto1 := &dto.Location{
		LocationID: "location-1",
		Name:       "location 1",
	}

	locationDto2 := &dto.Location{
		LocationID: "location-2",
		Name:       "location 2",
	}

	listLocationDTO := []*dto.Location{locationDto1, locationDto2}

	testCases := []TestCase{
		{
			name: "happy case",
			ctx:  ctx,
			req: &pb.GetGrantedLocationsOfStaffRequest{
				StaffId: "staff-1",
				Limit:   10,
			},
			expectedErr: nil,
			expectedResp: &pb.GetGrantedLocationsOfStaffResponse{
				Locations: listLocationResponse,
			},
			setup: func(ctx context.Context) {
				locationService.On("GetListGrantedLocationOfStaff", ctx, mock.Anything, mock.Anything, mock.Anything).
					Return(listLocationDTO, nil).Once()
			},
		},
		{
			name: "error case invalid staff id",
			ctx:  ctx,
			req: &pb.GetGrantedLocationsOfStaffRequest{
				Limit: 10,
			},
			expectedErr:  status.Error(codes.InvalidArgument, "staff id must be not empty"),
			expectedResp: (*pb.GetGrantedLocationsOfStaffResponse)(nil),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "error case invalid limit number",
			ctx:  ctx,
			req: &pb.GetGrantedLocationsOfStaffRequest{
				StaffId: "staff-1",
			},
			expectedErr:  status.Error(codes.InvalidArgument, "limit number must be not empty"),
			expectedResp: (*pb.GetGrantedLocationsOfStaffResponse)(nil),
			setup:        func(ctx context.Context) {},
		},
		{
			name: "error case get granted location failed",
			ctx:  ctx,
			req: &pb.GetGrantedLocationsOfStaffRequest{
				StaffId: "staff-1",
				Limit:   10,
			},
			expectedErr:  status.Error(codes.Internal, "err get granted location"),
			expectedResp: (*pb.GetGrantedLocationsOfStaffResponse)(nil),
			setup: func(ctx context.Context) {
				locationService.On("GetListGrantedLocationOfStaff", ctx, mock.Anything, mock.Anything, mock.Anything).
					Return(nil, status.Error(codes.Internal, fmt.Sprintf("err get granted location"))).Once()
			},
		},
	}

	// Do Test
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setup(testCase.ctx)
			req := testCase.req.(*pb.GetGrantedLocationsOfStaffRequest)
			resp, err := ctl.GetGrantedLocationsOfStaff(testCase.ctx, req)
			assert.Equal(t, testCase.expectedErr, err)
			assert.Equal(t, testCase.expectedResp, resp)
		})
	}
}
