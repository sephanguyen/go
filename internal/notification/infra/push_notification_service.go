package infra

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/firebase"
	"github.com/manabie-com/backend/internal/notification/config"
	"github.com/manabie-com/backend/internal/notification/entities"
	"github.com/manabie-com/backend/internal/notification/infra/metrics"
	"github.com/manabie-com/backend/internal/notification/services/utils"

	"firebase.google.com/go/v4/messaging"
	"github.com/gogo/protobuf/types"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
)

type PushNotificationService interface {
	RetrievePushedMessages(ctx context.Context, deviceToken string, limit int, since *types.Timestamp) ([]utils.RetrievedPushNotificationMsg, error)
	PushNotificationForUser(ctx context.Context, users entities.UserDeviceTokens, notification *entities.InfoNotification, notificationMsg *entities.InfoNotificationMsg) (success, failure int, err error)
}

type pushNotificationServiceImpl struct {
	notificationPusher  firebase.NotificationPusher
	NotificationMetrics metrics.NotificationMetrics
}

func NewPushNotificationService(notificationPusher firebase.NotificationPusher, metric metrics.NotificationMetrics) PushNotificationService {
	return &pushNotificationServiceImpl{
		notificationPusher:  notificationPusher,
		NotificationMetrics: metric,
	}
}

func (svc *pushNotificationServiceImpl) RetrievePushedMessages(
	ctx context.Context, deviceToken string, limit int, since *types.Timestamp) ([]utils.RetrievedPushNotificationMsg, error) {
	msges, err := svc.notificationPusher.RetrievePushedMessages(ctx, deviceToken, limit, since)
	if err != nil {
		return nil, err
	}
	return fcmToPushedMessage(msges), nil
}

func (svc *pushNotificationServiceImpl) PushNotificationForUser(ctx context.Context, userDeviceTokens entities.UserDeviceTokens, notification *entities.InfoNotification, notificationMsg *entities.InfoNotificationMsg) (
	success, failure int, err error) {
	tokens := make([]string, 0, len(userDeviceTokens))
	for _, u := range userDeviceTokens {
		if u.AllowNotification.Bool && u.DeviceToken.String != "" {
			tokens = append(tokens, u.DeviceToken.String)
		}
	}

	isMuteNotification := notification.IsMuteMode()
	successCount := 0
	failureCount := 0
	var errTenantSendTokens *firebase.SendTokensError

	switch {
	case len(tokens) == 0:
		return 0, 0, nil
	case len(tokens) == 1:
		msg := toMessage(notification, notificationMsg, isMuteNotification)
		if err := svc.notificationPusher.SendToken(ctx, msg, tokens[0]); err != nil {
			svc.NotificationMetrics.RecordPushNotificationErrors(metrics.StatusFail, 1)
			return 0, 1, fmt.Errorf("svc.PushNotificationService.SendToken: %w", err)
		}
		svc.NotificationMetrics.RecordPushNotificationErrors(metrics.StatusOK, 1)
		return 1, 0, nil
	default:
		logger := ctxzap.Extract(ctx)
		logger.Sugar().Infof("svc.PushNotificationService.SendTokens: number of device tokens to be sent is %v", len(tokens))
		msg := toMulticastMessage(notification, notificationMsg, isMuteNotification)
		successCount, failureCount, errTenantSendTokens = svc.notificationPusher.SendTokens(ctx, msg, tokens)
		logger.Sugar().Infof("svc.PushNotificationService.SendTokens: number of device tokens sent success is/are %v", successCount)
		logger.Sugar().Infof("svc.PushNotificationService.SendTokens: number of device tokens sent failed is/are %v", failureCount)
		svc.NotificationMetrics.RecordPushNotificationErrors(metrics.StatusOK, float64(successCount))
		svc.NotificationMetrics.RecordPushNotificationErrors(metrics.StatusFail, float64(failureCount))

		if errTenantSendTokens != nil {
			if errTenantSendTokens.BatchCombinedError != nil {
				logger.Sugar().Errorf("svc.PushNotificationService.SendTokens batch error: %v", errTenantSendTokens.BatchCombinedError)
			}
			if errTenantSendTokens.DirectError != nil {
				return successCount, failureCount, fmt.Errorf("svc.PushNotificationService.SendTokens - SendMulticast error: %v", errTenantSendTokens.DirectError)
			}
		}
	}

	return successCount, failureCount, nil
}

func fcmToPushedMessage(fmcMsges []*messaging.MulticastMessage) (ret []utils.RetrievedPushNotificationMsg) {
	for _, item := range fmcMsges {
		converted := utils.RetrievedPushNotificationMsg{
			Tokens: item.Tokens,
			Data:   item.Data,
		}
		if item.Notification != nil {
			converted.Body = item.Notification.Body
			converted.Title = item.Notification.Title
		}
		ret = append(ret, converted)
	}
	return ret
}

func toMessage(notification *entities.InfoNotification, notificationMsg *entities.InfoNotificationMsg, mute bool) *messaging.Message {
	title := notificationMsg.Title.String
	contentEnt, _ := notificationMsg.GetContent()
	content := contentEnt.GetText()

	title, content = limitLength(title, content)
	firebaseMessage := &messaging.Message{
		Data: toMessageData(notification),
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Headers: map[string]string{
				"apns-priority": "10",
			},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{Sound: "default"},
			},
		},
	}

	if !mute {
		firebaseMessage.Notification = &messaging.Notification{
			Title: title,
			Body:  content,
		}
	}

	return firebaseMessage
}

func toMessageData(notification *entities.InfoNotification) map[string]string {
	var data string
	_ = notification.Data.AssignTo(&data)

	return map[string]string{
		"id":                 notification.NotificationID.String,
		"type":               notification.Type.String,
		"data":               data,
		"click_action":       firebase.ClickAction,
		"notification_event": notification.Event.String,
	}
}

func limitLength(title string, content string) (string, string) {
	if len(title)+len(content) <= config.CharacterLengthMax {
		return title, content
	}

	if len(title) > config.CharacterLengthMax/2 {
		title = title[:(config.CharacterLengthMax/2 - 3)]
		title += "..."
	}

	if len(content) > config.CharacterLengthMax/2 {
		content = content[:(config.CharacterLengthMax/2 - 3)]
		content += "..."
	}

	return title, content
}

func toMulticastMessage(notification *entities.InfoNotification, notificationMsg *entities.InfoNotificationMsg, mute bool) *messaging.MulticastMessage {
	title := notificationMsg.Title.String
	contentEnt, _ := notificationMsg.GetContent()
	content := contentEnt.GetText()

	title, content = limitLength(title, content)
	firebaseMessage := &messaging.MulticastMessage{
		Data: toMessageData(notification),
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{Sound: "default"},
			},
		},
	}

	if !mute {
		firebaseMessage.Notification = &messaging.Notification{
			Title: title,
			Body:  content,
		}
	}

	return firebaseMessage
}
