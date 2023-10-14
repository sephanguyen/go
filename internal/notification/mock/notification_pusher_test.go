package mock

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/golibs/idutil"

	"firebase.google.com/go/v4/messaging"
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"go.uber.org/multierr"
)

func TestNotificationPusher_SendTokensAndRetrievePushedMessages(t *testing.T) {
	yNotifier := NewNotificationPusher()

	deviceTokens := make([]string, 0)

	notiDummyInput := messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: "dummy",
			Body:  "dummy",
		},
	}

	testCases := []struct {
		Name  string
		Err   *firebase.SendTokensError
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Setup: func(ctx context.Context) {
				deviceTokens = append(deviceTokens, MockNotificationPusherValidDeviceToken+"token1")
				deviceTokens = append(deviceTokens, MockNotificationPusherValidDeviceToken+"token2")
			},
		},
		{
			Name: "error invalid device token",
			Err: &firebase.SendTokensError{
				BatchCombinedError: multierr.Combine(errors.New(MockNotificationPusherInvalidDeviceToken), errors.New(MockNotificationPusherInvalidDeviceToken)),
			},
			Setup: func(ctx context.Context) {
				deviceTokens = make([]string, 0)
				deviceTokens = append(deviceTokens, MockNotificationPusherInvalidDeviceToken+"-token1")
				deviceTokens = append(deviceTokens, MockNotificationPusherInvalidDeviceToken+"-token2")
				deviceTokens = append(deviceTokens, MockNotificationPusherValidDeviceToken+"-token3")
			},
		},
		{
			Name: "error device token with unexpected error",
			Err: &firebase.SendTokensError{
				DirectError: errors.New(MockNotificationPusherDeviceTokenWithUnexpectedError),
			},
			Setup: func(ctx context.Context) {
				deviceTokens = make([]string, 0)
				deviceTokens = append(deviceTokens, MockNotificationPusherDeviceTokenWithUnexpectedError+"-token1")
				deviceTokens = append(deviceTokens, MockNotificationPusherDeviceTokenWithUnexpectedError+"-token2")
			},
		},
		{
			Name: "partial failed",
			Err: &firebase.SendTokensError{
				BatchCombinedError: multierr.Combine(errors.New(MockNotificationPusherInvalidDeviceToken)),
			},
			Setup: func(ctx context.Context) {
				deviceTokens = make([]string, 0)
				deviceTokens = append(deviceTokens, MockNotificationPusherValidDeviceToken+"token1")
				deviceTokens = append(deviceTokens, MockNotificationPusherInvalidDeviceToken+"-token2")
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(context.Background())
			_, _, err := yNotifier.SendTokens(context.Background(), &notiDummyInput, deviceTokens)
			assert.Equal(t, testCase.Err, err)

			for _, deviceToken := range deviceTokens {
				if !strings.Contains(deviceToken, MockNotificationPusherInvalidDeviceToken) &&
					!strings.Contains(deviceToken, MockNotificationPusherDeviceTokenWithUnexpectedError) {
					retrievePushedMessages, errRetrieve := yNotifier.RetrievePushedMessages(context.Background(), deviceToken, 1, &types.Timestamp{})
					assert.Nil(t, errRetrieve)
					assert.NotEqual(t, 0, len(retrievePushedMessages))

					if len(retrievePushedMessages) > 0 {
						for _, retrievePushedMessage := range retrievePushedMessages {
							assert.Equal(t, notiDummyInput.Notification.Title, retrievePushedMessage.Notification.Title)
							assert.Equal(t, notiDummyInput.Notification.Body, retrievePushedMessage.Notification.Body)
						}
					}
				}
			}
		})
	}

}

func TestNotificationPusher_SendTokenAndRetrievePushedMessages(t *testing.T) {
	yNotifier := NewNotificationPusher()
	notiDummyInput := messaging.Message{
		Notification: &messaging.Notification{
			Title: "dummy",
			Body:  "dummy",
		},
	}
	deviceToken := "token"
	testCases := []struct {
		Name  string
		Err   error
		Setup func(ctx context.Context)
	}{
		{
			Name: "happy case",
			Err:  nil,
			Setup: func(ctx context.Context) {
				deviceToken = "token-happy-case"
			},
		},
		{
			Name: "error device token with unexpected error",
			Err:  errors.New(MockNotificationPusherDeviceTokenWithUnexpectedError),
			Setup: func(ctx context.Context) {
				deviceToken = MockNotificationPusherDeviceTokenWithUnexpectedError + "-" + idutil.ULIDNow()
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			testCase.Setup(context.Background())
			err := yNotifier.SendToken(context.Background(), &notiDummyInput, deviceToken)
			if testCase.Err == nil {
				assert.Nil(t, err)

				retrievePushedMessages, errRetrieve := yNotifier.RetrievePushedMessages(context.Background(), deviceToken, 1, &types.Timestamp{})
				assert.Nil(t, errRetrieve)
				assert.NotEqual(t, 0, len(retrievePushedMessages))

				if len(retrievePushedMessages) > 0 {
					for _, retrievePushedMessage := range retrievePushedMessages {
						assert.Equal(t, notiDummyInput.Notification.Title, retrievePushedMessage.Notification.Title)
						assert.Equal(t, notiDummyInput.Notification.Body, retrievePushedMessage.Notification.Body)
					}
				}
			} else {
				assert.Equal(t, testCase.Err, err)
			}
		})
	}

}
