package controller

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/application/queries"
	"github.com/manabie-com/backend/internal/notification/modules/system_notification/infrastructure/repo"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

type SystemNotificationReaderService struct {
	npb.UnimplementedSystemNotificationReaderServiceServer

	SystemNotificationQueryHandler interface {
		RetrieveSystemNotifications(ctx context.Context, payload *queries.RetrieveSystemNotificationPayload) *queries.RetrieveSystemNotificationResponse
	}
}

func NewSystemNotificationReaderService(db database.Ext) *SystemNotificationReaderService {
	return &SystemNotificationReaderService{
		SystemNotificationQueryHandler: &queries.SystemNotificationQueryHandler{
			DB:                            db,
			SystemNotificationRepo:        &repo.SystemNotificationRepo{},
			SystemNotificationContentRepo: &repo.SystemNotificationContentRepo{},
		},
	}
}
