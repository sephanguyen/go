package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationModifierService) SetStatusForUserNotifications(ctx context.Context, req *npb.SetStatusForUserNotificationsRequest) (*npb.SetStatusForUserNotificationsResponse, error) {
	if req.Status == cpb.UserNotificationStatus_USER_NOTIFICATION_STATUS_NONE {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid request Status %v", req.Status))
	}
	userID := interceptors.UserIDFromContext(ctx)
	err := svc.UserNotificationRepo.SetStatus(ctx, svc.DB, database.Text(userID), database.TextArray(req.UserNotificationIds), database.Text(req.Status.String()))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("UserInfoNotificationRepo.SetStatus: %v", err))
	}
	return &npb.SetStatusForUserNotificationsResponse{}, nil
}
