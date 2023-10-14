package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cmd *SystemNotificationCommandHandler) SetSystemNotificationStatus(ctx context.Context, db database.QueryExecer, payload *payloads.SetSystemNotificationStatusPayload) error {
	userID := interceptors.UserIDFromContext(ctx)

	isBelong, err := cmd.SystemNotificationRepo.CheckUserBelongToSystemNotification(ctx, db, userID, payload.SystemNotificationID)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed CheckUserBelongToSystemNotification:%+v", err))
	}

	if !isBelong {
		return status.Error(codes.InvalidArgument, "user does not belong to this system notification")
	}

	err = cmd.SystemNotificationRepo.SetStatus(ctx, db, payload.SystemNotificationID, payload.Status)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed SetStatus: %+v", err))
	}

	return nil
}
