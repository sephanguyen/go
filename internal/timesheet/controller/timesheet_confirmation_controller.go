package controller

import (
	"context"
	"strings"

	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimesheetConfirmationController struct {
	TimesheetConfirmationWindowService interface {
		GetPeriod(ctx context.Context, date *timestamppb.Timestamp) (*dto.TimesheetConfirmationPeriod, error)
		ConfirmPeriod(ctx context.Context, request *pb.ConfirmTimesheetWithLocationRequest) error
		GetTimesheetLocationList(ctx context.Context, request *dto.GetTimesheetLocationListReq) (*dto.GetTimesheetLocationListOut, error)
		GetNonConfirmedLocationCount(ctx context.Context, request *dto.GetNonConfirmedLocationCountReq) (*dto.GetNonConfirmedLocationCountOut, error)
	}
}

func (c *TimesheetConfirmationController) GetConfirmationPeriodByDate(ctx context.Context, request *pb.GetTimesheetConfirmationPeriodByDateRequest) (*pb.GetTimesheetConfirmationPeriodByDateResponse, error) {
	err := dto.ValidateGetPeriodInfo(request)

	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	periodDto, err := c.TimesheetConfirmationWindowService.GetPeriod(ctx, request.Date)

	if err != nil {
		return nil, err
	}

	return dto.NewTimesheetConfirmationPeriodToRPCResponse(periodDto), nil
}

func (c *TimesheetConfirmationController) ConfirmTimesheet(ctx context.Context, request *pb.ConfirmTimesheetWithLocationRequest) (*pb.ConfirmTimesheetWithLocationResponse, error) {

	err := validateConfirmTimesheetRequest(request)
	if err != nil {
		return nil, err
	}

	err = c.TimesheetConfirmationWindowService.ConfirmPeriod(ctx, request)
	if err != nil {
		return nil, err
	}

	return &pb.ConfirmTimesheetWithLocationResponse{
		Success: true,
	}, nil
}

func (c *TimesheetConfirmationController) GetTimesheetLocationList(ctx context.Context, request *pb.GetTimesheetLocationListRequest) (*pb.GetTimesheetLocationListResponse, error) {
	reqDto := dto.NewGetTimesheetLocationListRequest(request)
	reqDto.ConvertTimeToJPTimezone()

	res, err := c.TimesheetConfirmationWindowService.GetTimesheetLocationList(ctx, reqDto)
	if err != nil {
		return nil, err
	}

	return &pb.GetTimesheetLocationListResponse{
		Locations:          dto.ConvertTimesheetLocationListToRPC(res.Locations),
		LocationsAggregate: dto.ConvertTimesheetLocationAggregateToRPC(res.LocationAggregate),
	}, nil
}

func (c *TimesheetConfirmationController) GetNonConfirmedLocationCount(ctx context.Context, request *pb.GetNonConfirmedLocationCountRequest) (*pb.GetNonConfirmedLocationCountResponse, error) {
	reqDto := dto.NewGetNonConfirmedLocationCountRequest(request)
	reqDto.ConvertTimeToJPTimezone()

	res, err := c.TimesheetConfirmationWindowService.GetNonConfirmedLocationCount(ctx, reqDto)
	if err != nil {
		return nil, err
	}

	return dto.ConvertGetNonConfirmedLocationCountOutToRPC(res), nil
}

func validateConfirmTimesheetRequest(request *pb.ConfirmTimesheetWithLocationRequest) error {
	if len(request.LocationIds) == 0 {
		return status.Error(codes.InvalidArgument, "location ids cannot be empty")
	}

	if strings.TrimSpace(request.PeriodId) == "" {
		return status.Error(codes.InvalidArgument, "period id must not be empty")
	}

	return nil
}
