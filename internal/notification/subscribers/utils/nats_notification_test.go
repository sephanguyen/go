package utils

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetRawNotificationNatsMessage(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		t.Parallel()
		natsNoti := GenSampleNatsNotification()
		raw, err := GetRawNotificationNatsMessage(natsNoti.NotificationConfig.Notification)
		assert.Nil(t, err)

		expectedRaw := fmt.Sprintf(`{"blocks":[{"key":"aip8i","text":"%s","type":"unstyled","depth":0,"inlineStyleRanges":[],"entityRanges":[],"data":{}}],"entityMap":{}}`, natsNoti.NotificationConfig.Notification.Message)
		assert.Equal(t, expectedRaw, raw)
	})

	t.Run("null message", func(t *testing.T) {
		t.Parallel()
		natsNoti := GenSampleNatsNotification()
		raw, err := GetRawNotificationNatsMessage(natsNoti.NotificationConfig.Notification)
		assert.Nil(t, err)

		expectedRaw := fmt.Sprintf(`{"blocks":[{"key":"aip8i","text":"%s","type":"unstyled","depth":0,"inlineStyleRanges":[],"entityRanges":[],"data":{}}],"entityMap":{}}`, natsNoti.NotificationConfig.Notification.Message)
		assert.Equal(t, expectedRaw, raw)
	})

	t.Run("null content", func(t *testing.T) {
		t.Parallel()
		natsNoti := GenSampleNatsNotification()
		raw, err := GetRawNotificationNatsMessage(natsNoti.NotificationConfig.Notification)
		assert.Nil(t, err)

		expectedRaw := fmt.Sprintf(`{"blocks":[{"key":"aip8i","text":"%s","type":"unstyled","depth":0,"inlineStyleRanges":[],"entityRanges":[],"data":{}}],"entityMap":{}}`, natsNoti.NotificationConfig.Notification.Message)
		assert.Equal(t, expectedRaw, raw)
	})

	t.Run("null message and content", func(t *testing.T) {
		t.Parallel()
		natsNoti := GenSampleNatsNotification()
		natsNoti.NotificationConfig.Notification.Content = ""
		natsNoti.NotificationConfig.Notification.Title = ""
		_, err := GetRawNotificationNatsMessage(natsNoti.NotificationConfig.Notification)
		expectedErr := errors.New("request Notification.Message.Content is null")
		assert.Equal(t, expectedErr, err)
	})
}
