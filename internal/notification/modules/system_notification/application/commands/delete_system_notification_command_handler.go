package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure"
)

type DeleteSystemNotificationCommandHandler struct {
	infrastructure.SystemNotificationRepo
	infrastructure.SystemNotificationRecipientRepo
	infrastructure.SystemNotificationContentRepo
}

func (cmd *DeleteSystemNotificationCommandHandler) SoftDeleteSystemNotification(ctx context.Context, db database.QueryExecer, softDeletePayload *payloads.SoftDeleteSystemNotificationPayload) error {
	// get system notification by reference ID
	sn, err := cmd.SystemNotificationRepo.FindByReferenceID(ctx, db, softDeletePayload.ReferenceID)
	if err != nil {
		return fmt.Errorf("err FindByReferenceID: %+v", err)
	}

	if sn == nil {
		return fmt.Errorf("not found System Notification by ReferenceID %s", softDeletePayload.ReferenceID)
	}

	snID := sn.SystemNotificationID.String
	if snID == "" {
		return fmt.Errorf("err SystemNotificationID is empty")
	}

	// soft delete it's recipients
	err = cmd.SystemNotificationRecipientRepo.SoftDeleteBySystemNotificationID(ctx, db, snID)
	if err != nil {
		return fmt.Errorf("err soft delete recipient: %+v", err)
	}

	// soft delete it's contents
	err = cmd.SystemNotificationContentRepo.SoftDeleteBySystemNotificationID(ctx, db, snID)
	if err != nil {
		return fmt.Errorf("err soft delete content: %+v", err)
	}

	// soft delete system notification
	_ = sn.DeletedAt.Set(time.Now())
	err = cmd.SystemNotificationRepo.UpsertSystemNotification(ctx, db, sn)
	if err != nil {
		return fmt.Errorf("err soft delete system notification: %+v", err)
	}

	return nil
}
