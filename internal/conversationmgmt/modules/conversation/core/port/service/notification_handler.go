package service

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
)

type NotificationHandler interface {
	PushNotification(ctx context.Context, message *domain.OfflineMessage) error
}
