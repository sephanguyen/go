package controller

import (
	"context"

	pt "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ImportMasterDataController struct {
	ImportTimesheetConfigService interface {
		ImportTimesheetConfig(ctx context.Context, payload []byte) ([]*pt.ImportTimesheetConfigError, error)
	}
}

func (c *ImportMasterDataController) ImportTimesheetConfig(ctx context.Context, req *pt.ImportTimesheetConfigRequest) (*pt.ImportTimesheetConfigResponse, error) {
	if len(req.Payload) == 0 {
		return nil, status.Error(codes.InvalidArgument, "missing payload for import timesheet config")
	}

	rowsError, err := c.ImportTimesheetConfigService.ImportTimesheetConfig(ctx, req.Payload)
	if err != nil {
		return nil, err
	}

	return &pt.ImportTimesheetConfigResponse{Errors: rowsError}, nil
}
