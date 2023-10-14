package service

import (
	"context"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
)

func (svc *notificationHandlerServiceImpl) PushNotification(_ context.Context, _ *domain.OfflineMessage) error {
	// TODO: Update logic push notification for learner if they are offline
	return nil
}
