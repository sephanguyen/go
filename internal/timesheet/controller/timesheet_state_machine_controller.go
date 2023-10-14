package controller

import (
	"context"
	"strings"

	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TimesheetStateMachineController struct {
	TimesheetStateMachineService interface {
		DeleteTimesheet(ctx context.Context, timesheetID string) error
		SubmitTimesheet(ctx context.Context, timesheetID string) error
		ApproveTimesheet(ctx context.Context, timesheetIDs []string) error
		CancelApproveTimesheet(ctx context.Context, timesheetID string) error
		ConfirmTimesheet(ctx context.Context, timesheetIDs []string) error
		CancelSubmissionTimesheet(ctx context.Context, timesheetID string) error
	}
	MastermgmtConfigurationService interface {
		CheckPartnerTimesheetServiceIsOn(ctx context.Context) (bool, error)
	}
}

func (c *TimesheetStateMachineController) DeleteTimesheet(ctx context.Context, request *pb.DeleteTimesheetRequest) (*pb.DeleteTimesheetResponse, error) {
	timesheetServiceStatus, err := c.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !timesheetServiceStatus {
		return nil, status.Errorf(codes.PermissionDenied, "don't have permission to modify timesheet")
	}
	err = validateTimesheetIDRequest(request.TimesheetId)
	if err != nil {
		return nil, err
	}

	err = c.TimesheetStateMachineService.DeleteTimesheet(ctx, request.TimesheetId)
	if err != nil {
		return nil, err
	}

	return &pb.DeleteTimesheetResponse{
		Success: true,
	}, nil
}

func (c *TimesheetStateMachineController) SubmitTimesheet(ctx context.Context, request *pb.SubmitTimesheetRequest) (*pb.SubmitTimesheetResponse, error) {
	timesheetServiceStatus, err := c.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !timesheetServiceStatus {
		return nil, status.Errorf(codes.PermissionDenied, "don't have permission to modify timesheet")
	}

	err = validateTimesheetIDRequest(request.TimesheetId)
	if err != nil {
		return nil, err
	}

	err = c.TimesheetStateMachineService.SubmitTimesheet(ctx, request.TimesheetId)
	if err != nil {
		return nil, err
	}

	return &pb.SubmitTimesheetResponse{
		Success: true,
	}, nil
}

func (c *TimesheetStateMachineController) ApproveTimesheet(ctx context.Context, request *pb.ApproveTimesheetRequest) (*pb.ApproveTimesheetResponse, error) {
	timesheetServiceStatus, err := c.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !timesheetServiceStatus {
		return nil, status.Errorf(codes.PermissionDenied, "don't have permission to modify timesheet")
	}

	err = validateApproveConfirmTimesheetIDsRequest(request.TimesheetIds)
	if err != nil {
		return nil, err
	}

	err = c.TimesheetStateMachineService.ApproveTimesheet(ctx, request.TimesheetIds)
	if err != nil {
		return nil, err
	}

	return &pb.ApproveTimesheetResponse{
		Success: true,
	}, nil
}

func (c *TimesheetStateMachineController) ConfirmTimesheet(ctx context.Context, request *pb.ConfirmTimesheetRequest) (*pb.ConfirmTimesheetResponse, error) {
	timesheetServiceStatus, err := c.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !timesheetServiceStatus {
		return nil, status.Errorf(codes.PermissionDenied, "don't have permission to modify timesheet")
	}

	err = validateApproveConfirmTimesheetIDsRequest(request.TimesheetIds)
	if err != nil {
		return nil, err
	}

	err = c.TimesheetStateMachineService.ConfirmTimesheet(ctx, request.TimesheetIds)
	if err != nil {
		return nil, err
	}

	return &pb.ConfirmTimesheetResponse{
		Success: true,
	}, nil
}

func (c *TimesheetStateMachineController) CancelApproveTimesheet(ctx context.Context, request *pb.CancelApproveTimesheetRequest) (*pb.CancelApproveTimesheetResponse, error) {

	timesheetServiceStatus, err := c.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !timesheetServiceStatus {
		return nil, status.Errorf(codes.PermissionDenied, "don't have permission to modify timesheet")
	}

	err = validateTimesheetIDRequest(request.TimesheetId)
	if err != nil {
		return nil, err
	}

	err = c.TimesheetStateMachineService.CancelApproveTimesheet(ctx, request.TimesheetId)
	if err != nil {
		return nil, err
	}

	return &pb.CancelApproveTimesheetResponse{
		Success: true,
	}, nil
}

func validateTimesheetIDRequest(timesheetID string) error {
	if strings.TrimSpace(timesheetID) == "" {
		return status.Error(codes.InvalidArgument, "timesheet id cannot be empty")
	}

	return nil
}

func validateApproveConfirmTimesheetIDsRequest(timesheetIDs []string) error {
	if len(timesheetIDs) == 0 {
		return status.Error(codes.InvalidArgument, "timesheet ids cannot be empty")
	}

	return nil
}

func (c *TimesheetStateMachineController) CancelSubmissionTimesheet(ctx context.Context, request *pb.CancelSubmissionTimesheetRequest) (*pb.CancelSubmissionTimesheetResponse, error) {

	timesheetServiceStatus, err := c.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !timesheetServiceStatus {
		return nil, status.Errorf(codes.PermissionDenied, "don't have permission to modify timesheet")
	}

	err = validateTimesheetIDRequest(request.TimesheetId)
	if err != nil {
		return nil, err
	}

	err = c.TimesheetStateMachineService.CancelSubmissionTimesheet(ctx, request.TimesheetId)
	if err != nil {
		return nil, err
	}

	return &pb.CancelSubmissionTimesheetResponse{
		Success: true,
	}, nil
}
