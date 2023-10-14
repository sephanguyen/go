package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure"
	systemNotification "github.com/manabie-com/backend/internal/notification/modules/system_notification/util/mapper/systemnotification"
)

type SystemNotificationCommandHandler struct {
	SystemNotificationRepo infrastructure.SystemNotificationRepo
}

func (cmd *SystemNotificationCommandHandler) UpsertSystemNotification(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationPayload) error {
	// check if this system notification existed
	notification, err := cmd.SystemNotificationRepo.FindByReferenceID(ctx, db, payload.SystemNotification.ReferenceID)
	if err != nil {
		return fmt.Errorf("failed FindByReferenceID: %+v", err)
	}

	if notification == nil {
		payload.SystemNotification.SystemNotificationID = idutil.ULIDNow()
	} else {
		payload.SystemNotification.SystemNotificationID = notification.SystemNotificationID.String
	}

	entity, err := systemNotification.ToEntity(payload.SystemNotification)
	if err != nil {
		return err
	}

	err = cmd.SystemNotificationRepo.UpsertSystemNotification(ctx, db, entity)
	if err != nil {
		return fmt.Errorf("failed UpsertSystemNotification: %+v", err)
	}

	return nil
}
