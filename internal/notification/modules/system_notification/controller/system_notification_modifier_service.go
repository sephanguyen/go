package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/commands/payloads"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure/repo"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

type SystemNotificationModifierService struct {
	npb.UnimplementedSystemNotificationModifierServiceServer
	DB database.Ext // put it here to enable using transaction in all command handlers

	SystemNotificationCommandHandler interface {
		UpsertSystemNotification(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationPayload) error
		SetSystemNotificationStatus(ctx context.Context, db database.QueryExecer, payload *payloads.SetSystemNotificationStatusPayload) error
	}
	SystemNotificationRecipientCommandHandler interface {
		UpsertSystemNotificationRecipients(ctx context.Context, db database.QueryExecer, payload *payloads.UpsertSystemNotificationRecipientPayload) error
	}
}

func NewSystemNotificationModifierService(db database.Ext) *SystemNotificationModifierService {
	return &SystemNotificationModifierService{
		DB: db,
		SystemNotificationCommandHandler: &commands.SystemNotificationCommandHandler{
			SystemNotificationRepo: &repo.SystemNotificationRepo{},
		},
		SystemNotificationRecipientCommandHandler: &commands.SystemNotificationRecipientCommandHandler{
			SystemNotificationRecipientRepo: &repo.SystemNotificationRecipientRepo{},
		},
	}
}
