package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/util/mapper/systemnotification"
)

type SystemNotificationContentHandler struct {
	SystemNotificationContentRepo infrastructure.SystemNotificationContentRepo
}

func (cmd *SystemNotificationContentHandler) UpsertSystemNotificationContents(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationContentPayload) error {
	// convert to db entities
	entities, err := systemnotification.ToSystemNotificationContentEntities(payload.SystemNotificationID, payload.SystemNotificationContents)
	if err != nil {
		return fmt.Errorf("failed ToSystemNotificationContentEntities: %+v", err)
	}

	// delete all content by system notification ID if exists
	err = cmd.SystemNotificationContentRepo.SoftDeleteBySystemNotificationID(ctx, db, payload.SystemNotificationID)
	if err != nil {
		return fmt.Errorf("failed SoftDeleteBySystemNotificationID: %+v", err)
	}

	// insert all new content by system notification ID
	err = cmd.SystemNotificationContentRepo.BulkInsertSystemNotificationContents(ctx, db, entities)
	if err != nil {
		return fmt.Errorf("failed BulkInsertSystemNotificationContents: %+v", err)
	}

	return nil
}
