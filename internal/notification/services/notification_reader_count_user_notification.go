package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (svc *NotificationReaderService) CountUserNotification(ctx context.Context, req *npb.CountUserNotificationRequest) (*npb.CountUserNotificationResponse, error) {
	userID := interceptors.UserIDFromContext(ctx)
	numByStatus, total, err := svc.UserInfoNotificationRepo.CountByStatus(ctx, svc.DB, database.Text(userID), database.Text(req.Status.String()))
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("UserInfoNotificationRepo.CountByStatus %v", err))
	}
	resp := &npb.CountUserNotificationResponse{
		NumByStatus: int32(numByStatus),
		Total:       int32(total),
	}
	return resp, nil
}
