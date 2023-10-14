package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	nats_service_utils "github.com/manabie-com/backend/internal/timesheet/service/nats"
	pb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AutoCreateTimesheetFlagController struct {
	JSM                   nats.JetStreamManagement
	AutoCreateFlagService interface {
		UpsertFlag(ctx context.Context, autoCreateFlag *dto.AutoCreateTimesheetFlag) error
	}

	MastermgmtConfigurationService interface {
		CheckPartnerTimesheetServiceIsOn(ctx context.Context) (bool, error)
	}
}

func (c *AutoCreateTimesheetFlagController) UpdateAutoCreateTimesheetFlag(ctx context.Context, request *pb.UpdateAutoCreateTimesheetFlagRequest) (*pb.UpdateAutoCreateTimesheetFlagResponse, error) {

	serviceStatus, err := c.MastermgmtConfigurationService.CheckPartnerTimesheetServiceIsOn(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	if !serviceStatus {
		return nil, status.Errorf(codes.FailedPrecondition, "don't have permission to modify timesheet")
	}

	autoCreateFlagReq := dto.NewAutoCreateTimeSheetFlagFromRPCUpdateRequest(request)

	err = autoCreateFlagReq.ValidateUpsertInfo()
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	err = c.AutoCreateFlagService.UpsertFlag(ctx, autoCreateFlagReq)
	if err != nil {
		return nil, err
	}

	err = nats_service_utils.PublishUpdateAutoCreateFlagEvent(ctx, &pb.NatsUpdateAutoCreateTimesheetFlagRequest{
		FlagOn:  autoCreateFlagReq.FlagOn,
		StaffId: autoCreateFlagReq.StaffID,
	}, c.JSM)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "jsm error: %v", err)
	}

	return &pb.UpdateAutoCreateTimesheetFlagResponse{Successful: true}, nil
}
