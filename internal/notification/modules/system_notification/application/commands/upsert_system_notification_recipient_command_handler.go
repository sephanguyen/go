package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure"
	systemNotification "github.com/manabie-com/backend/internal/notification/modules/system_notification/util/mapper/systemnotification"
)

type SystemNotificationRecipientCommandHandler struct {
	SystemNotificationRecipientRepo infrastructure.SystemNotificationRecipientRepo
}

func (cmd *SystemNotificationRecipientCommandHandler) UpsertSystemNotificationRecipients(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationRecipientPayload) error {
	entities, err := systemNotification.ToRecipientEntities(payload.SystemNotificationID, payload.Recipients)
	if err != nil {
		return err
	}
	// delete all existing recipients of this event
	err = cmd.SystemNotificationRecipientRepo.SoftDeleteBySystemNotificationID(ctx, db, payload.SystemNotificationID)
	if err != nil {
		return fmt.Errorf("failed SoftDeleteBySystemNotificationID: %+v", err)
	}

	// then bulk insert all recipients
	err = cmd.SystemNotificationRecipientRepo.BulkInsertSystemNotificationRecipients(ctx, db, entities)
	if err != nil {
		return fmt.Errorf("failed BulkInsertSystemNotificationRecipients: %+v", err)
	}
	return nil
}
