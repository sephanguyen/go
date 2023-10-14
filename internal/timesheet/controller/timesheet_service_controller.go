package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TimesheetServiceController struct {
	TimesheetService interface {
		CreateTimesheet(ctx context.Context, timesheet *dto.Timesheet) (string, error)
		UpdateTimesheet(ctx context.Context, timesheet *dto.Timesheet) error
		CountTimesheets(ctx context.Context, req *dto.TimesheetCountReq) (*dto.TimesheetCountOut, error)
		CountTimesheetsV2(ctx context.Context, req *dto.TimesheetCountV2Req) (*dto.TimesheetCountV2Out, error)
		CountSubmittedTimesheets(ctx context.Context, req *dto.CountSubmittedTimesheetsReq) (*dto.CountSubmittedTimesheetsResp, error)
	}

	MastermgmtConfigurationService interface {
		CheckPartnerTimesheetServiceIsOn(ctx context.Context) (bool, error)
	}

	ConfirmationWindowService interface {
		CheckModifyConditionByTimesheetDateAndLocation(ctx context.Context, timesheetDate *timestamppb.Timestamp, locationID string) (bool, error)
		CheckModifyConditionByTimesheetID(ctx context.Context, timesheetID string) (bool, error)
	}
}

func (t *TimesheetServiceController) CreateTimesheet(ctx context.Context, request *pb.CreateTimesheetRequest) (*pb.CreateTimesheetResponse, error) {
	serviceStatus, err := t.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !serviceStatus {
		return nil, status.Errorf(codes.PermissionDenied, "current partner doesn't have permission to modify timesheet")
	}

	isModify, err := t.ConfirmationWindowService.CheckModifyConditionByTimesheetDateAndLocation(ctx, request.TimesheetDate, request.LocationId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !isModify {
		return nil, status.Errorf(codes.FailedPrecondition, "all data in this period have been confirm")
	}

	timesheetReq := dto.NewTimesheetFromRPCCreateRequest(request)

	err = timesheetReq.ValidateCreateInfo()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	timesheetReq.NormalizedData()

	timesheetID, err := t.TimesheetService.CreateTimesheet(ctx, timesheetReq)
	if err != nil {
		return nil, err
	}

	return &pb.CreateTimesheetResponse{TimesheetId: timesheetID}, nil
}

func (t *TimesheetServiceController) UpdateTimesheet(ctx context.Context, request *pb.UpdateTimesheetRequest) (*pb.UpdateTimesheetResponse, error) {
	checkResult, err := t.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !checkResult {
		return nil, status.Errorf(codes.PermissionDenied, "current partner doesn't have permission to modify timesheet")
	}

	isModify, err := t.ConfirmationWindowService.CheckModifyConditionByTimesheetID(ctx, request.TimesheetId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !isModify {
		return nil, status.Errorf(codes.FailedPrecondition, "all data in this period have been confirm")
	}

	timesheetReq := dto.NewTimesheetFromRPCUpdateRequest(request)

	err = timesheetReq.ValidateUpdateInfo()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = t.TimesheetService.UpdateTimesheet(ctx, timesheetReq)
	if err != nil {
		return nil, err
	}

	return &pb.UpdateTimesheetResponse{Success: true}, nil
}

func (t *TimesheetServiceController) CountTimesheets(ctx context.Context, req *pb.CountTimesheetsRequest) (*pb.CountTimesheetsResponse, error) {
	checkResult, err := t.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !checkResult {
		return nil, status.Errorf(codes.PermissionDenied, "current partner doesn't have permission to modify timesheet")
	}

	reqDto := dto.NewTimesheetCountReqFromRPCCreateRequest(req)

	res, err := t.TimesheetService.CountTimesheets(ctx, reqDto)
	if err != nil {
		return nil, err
	}

	return &pb.CountTimesheetsResponse{
		AllCount:       res.AllCount,
		DraftCount:     res.DraftCount,
		SubmittedCount: res.SubmittedCount,
		ApprovedCount:  res.ApprovedCount,
		ConfirmedCount: res.ConfirmedCount,
	}, nil
}

func (t *TimesheetServiceController) CountTimesheetsV2(ctx context.Context, req *pb.CountTimesheetsV2Request) (*pb.CountTimesheetsV2Response, error) {
	checkResult, err := t.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !checkResult {
		return nil, status.Errorf(codes.PermissionDenied, "current partner doesn't have permission to modify timesheet")
	}

	reqDto := dto.NewTimesheetCountReqFromRPCCreateV2Request(req)

	res, err := t.TimesheetService.CountTimesheetsV2(ctx, reqDto)
	if err != nil {
		return nil, err
	}

	return &pb.CountTimesheetsV2Response{
		AllCount:       res.AllCount,
		DraftCount:     res.DraftCount,
		SubmittedCount: res.SubmittedCount,
		ApprovedCount:  res.ApprovedCount,
		ConfirmedCount: res.ConfirmedCount,
	}, nil
}

func (t *TimesheetServiceController) CountSubmittedTimesheets(ctx context.Context, req *pb.CountSubmittedTimesheetsRequest) (*pb.CountSubmittedTimesheetsResponse, error) {
	reqDto := dto.NewCountSubmittedTimesheetsRequest(req)

	checkResult, err := t.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res, err := t.TimesheetService.CountSubmittedTimesheets(ctx, reqDto)

	if err != nil {
		return nil, err
	}

	return &pb.CountSubmittedTimesheetsResponse{
		Count:     res.Count,
		IsEnabled: checkResult,
	}, nil
}
