package utils

import (
	"errors"
	"fmt"

	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"
)

func GetRawNotificationNatsMessage(natsNoti *ypb.NatsNotification) (string, error) {
	// validate message
	if !(natsNoti.Message != "" && natsNoti.GetContent() != "") {
		return "", errors.New("request Notification.Message.Content is null")
	}

	raw := fmt.Sprintf(`{"blocks":[{"key":"aip8i","text":"%s","type":"unstyled","depth":0,"inlineStyleRanges":[],"entityRanges":[],"data":{}}],"entityMap":{}}`, natsNoti.Message)
	return raw, nil
}
