package services

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/entities"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/jackc/pgtype"
)

func (svc *NotificationModifierService) pushNotificationToUsers(ctx context.Context, db database.QueryExecer, noti *entities.InfoNotification, notiMsg *entities.InfoNotificationMsg, userIDs []string) (int, int, error) {
	if len(userIDs) == 0 {
		return 0, 0, nil
	}

	var ids pgtype.TextArray
	_ = ids.Set(userIDs)

	userDeviceTokens, err := svc.UserDeviceTokenRepo.FindByUserIDs(ctx, db, ids)
	if err != nil {
		return 0, 0, fmt.Errorf("svc.UserRepo.Retrieve: %v", err)
	}
	if len(userDeviceTokens) == 0 {
		return 0, 0, nil
	}
	success, failure, err := svc.PushNotificationService.PushNotificationForUser(ctx, userDeviceTokens, noti, notiMsg)
	if err != nil {
		if success == 0 {
			return 0, 0, fmt.Errorf("svc.NotificationService.PushNotificationForUser full failure: %v", err)
		}

		logger := ctxzap.Extract(ctx)
		logger.Sugar().Errorf("svc.NotificationService.PushNotificationForUser partial failure: %v", err)
	}

	return success, failure, nil
}
