package subscribers

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/notification/consts"
	"github.com/manabie-com/backend/internal/notification/subscribers/mappers"
	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

//nolint
func (i *NotificationSubscriber) ProcessPushNotification(ctx context.Context, data *ypb.NatsCreateNotificationRequest) error {
	notiCpb, err := mappers.NatsNotificationToPb(data)

	if err != nil {
		return fmt.Errorf("%v toNotificationCpb", err.Error())
	}

	// switch send time to upsert data
	switch data.SendTime.Type {
	case consts.NotificationTypeScheduled:
		// scheduled
		_, err := i.notificationService.UpsertNotification(ctx, &npb.UpsertNotificationRequest{
			Notification: notiCpb,
		})
		return err

	case consts.NotificationTypeImmediate:
		// send immediately
		if data.NotificationConfig.PermanentStorage {
			res, err := i.notificationService.UpsertNotification(ctx, &npb.UpsertNotificationRequest{
				Notification: notiCpb,
			})
			if err != nil {
				return err
			}

			_, err = i.notificationService.SendNotification(ctx, &npb.SendNotificationRequest{
				NotificationId: res.NotificationId,
			})

			return err
		} else {
			return i.notificationService.SendNotificationToTargetWithoutSave(ctx, notiCpb)
		}
	default:
		return fmt.Errorf("send time type does not support")
	}
}
