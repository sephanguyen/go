package helpers

import (
	"context"
	"fmt"

	npb "github.com/manabie-com/backend/pkg/manabuf/notificationmgmt/v1"
)

func (helper *CommunicationHelper) UpdateDeviceToken(token string, deviceToken string, userID string) error {
	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), token)
	defer cancel()

	_, err := npb.NewNotificationModifierServiceClient(helper.NotificationMgmtGRPCConn).UpdateUserDeviceToken(ctx, &npb.UpdateUserDeviceTokenRequest{
		DeviceToken:       deviceToken,
		UserId:            userID,
		AllowNotification: true,
	})

	if err != nil {
		return fmt.Errorf("err npb.UpdateUserDeviceToken: %v", err)
	}

	return nil
}
