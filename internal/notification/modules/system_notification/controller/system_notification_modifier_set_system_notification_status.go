package controller

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *SystemNotificationModifierService) SetSystemNotificationStatus(ctx context.Context, req *npb.SetSystemNotificationStatusRequest) (*npb.SetSystemNotificationStatusResponse, error) {
	if req.GetSystemNotificationId() == "" {
		return nil, status.Error(codes.InvalidArgument, "missing SystemNotificationID")
	}

	payload := &payloads.SetSystemNotificationStatusPayload{
		SystemNotificationID: req.GetSystemNotificationId(),
		Status:               req.GetStatus().String(),
	}

	err := svc.SystemNotificationCommandHandler.SetSystemNotificationStatus(ctx, svc.DB, payload)

	if err != nil {
		// return nil, status.Error(codes.Internal, fmt.Sprintf("failed SetSystemNotificationStatus: %v", err))
		return nil, fmt.Errorf("failed SetSystemNotificationStatus: %+v", err)
	}

	return &npb.SetSystemNotificationStatusResponse{}, nil
}
