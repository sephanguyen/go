package mock

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manabie-com/backend/internal/golibs/firebase"

	"firebase.google.com/go/v4/messaging"
	"github.com/gogo/protobuf/types"
	"go.uber.org/multierr"
)

const (
	MockNotificationPusherDeviceTokenWithUnexpectedError = "device_token_with_unexpected_error"
	MockNotificationPusherInvalidDeviceToken             = "device_token_invalid"
	MockNotificationPusherValidDeviceToken               = "valid_device_token"
)

type NotificationPusher struct {
	pushedMulticastMessages map[string][]*messaging.MulticastMessage
}

func NewNotificationPusher() *NotificationPusher {
	return &NotificationPusher{
		pushedMulticastMessages: make(map[string][]*messaging.MulticastMessage),
	}
}

func (n *NotificationPusher) SendTokens(_ context.Context, msg *messaging.MulticastMessage, deviceTokens []string) (int, int, *firebase.SendTokensError) {
	batchResponses := &messaging.BatchResponse{
		SuccessCount: 0,
		FailureCount: 0,
		Responses:    []*messaging.SendResponse{},
	}
	for _, deviceToken := range deviceTokens {
		switch {
		case strings.Contains(deviceToken, MockNotificationPusherValidDeviceToken):
			n.pushedMulticastMessages[deviceToken] = append(n.pushedMulticastMessages[deviceToken], msg)
			batchResponses.SuccessCount++
			err := &messaging.SendResponse{
				Success:   true,
				MessageID: deviceToken,
				Error:     nil,
			}
			batchResponses.Responses = append(batchResponses.Responses, err)
		case strings.Contains(deviceToken, MockNotificationPusherInvalidDeviceToken):
			batchResponses.FailureCount++
			err := &messaging.SendResponse{
				Success:   false,
				MessageID: deviceToken,
				Error:     errors.New(MockNotificationPusherInvalidDeviceToken),
			}
			batchResponses.Responses = append(batchResponses.Responses, err)
		default:
			return 0, len(deviceTokens), &firebase.SendTokensError{
				DirectError: errors.New(MockNotificationPusherDeviceTokenWithUnexpectedError),
			}
		}
	}

	errRet := &firebase.SendTokensError{}
	successCount := batchResponses.SuccessCount
	failureCount := batchResponses.FailureCount

	for _, respDetail := range batchResponses.Responses {
		if respDetail.Error != nil {
			errRet.BatchCombinedError = multierr.Combine(errRet.BatchCombinedError, respDetail.Error)
		}
	}

	// critical and should be direct error
	// if failureCount == len(deviceTokens) && successCount == 0 {
	// 	errRet.DirectError = multierr.Combine(errRet.DirectError, fmt.Errorf("error when call fcm.Client.SendMulticast(): %w", errors.New(MockNotificationPusherDeviceTokenWithUnexpectedError)))
	// 	failureCount = len(deviceTokens) - successCount
	// }

	// For case all success -> error returned should be nil
	if errRet.DirectError == nil && errRet.BatchCombinedError == nil {
		errRet = nil
	}

	return successCount, failureCount, errRet
}

func (n *NotificationPusher) SendToken(_ context.Context, msg *messaging.Message, deviceToken string) error {
	if strings.Contains(deviceToken, MockNotificationPusherInvalidDeviceToken) {
		return nil
	}
	if strings.Contains(deviceToken, MockNotificationPusherDeviceTokenWithUnexpectedError) {
		return fmt.Errorf(MockNotificationPusherDeviceTokenWithUnexpectedError)
	}

	// convert to MulticastMessage type to using one array for mock
	n.pushedMulticastMessages[deviceToken] = append(n.pushedMulticastMessages[deviceToken], &messaging.MulticastMessage{
		Tokens: []string{msg.Token},
		Data:   msg.Data,
		Notification: &messaging.Notification{
			Title:    msg.Notification.Title,
			Body:     msg.Notification.Body,
			ImageURL: msg.Notification.ImageURL,
		},
		Android: msg.Android,
		Webpush: msg.Webpush,
		APNS:    msg.APNS,
	})

	return nil
}

func (n *NotificationPusher) RetrievePushedMessages(_ context.Context, deviceToken string, _ int, _ *types.Timestamp) ([]*messaging.MulticastMessage, error) {
	return n.pushedMulticastMessages[deviceToken], nil
}
