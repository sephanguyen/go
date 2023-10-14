package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/nats"
	subscribers "github.com/manabie-com/backend/internal/notification/subscribers"
	ypb "github.com/manabie-com/backend/pkg/manabuf/yasuo/v1"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type PushNotificationEvent struct {
	nats                   nats.JetStreamManagement
	zapLog                 *zap.Logger
	notificationSubscriber *subscribers.NotificationSubscriber
}

func NewPushNotificationEvent(nats nats.JetStreamManagement, logger *zap.Logger, notificationSubscriber *subscribers.NotificationSubscriber) *PushNotificationEvent {
	return &PushNotificationEvent{
		nats:                   nats,
		zapLog:                 logger,
		notificationSubscriber: notificationSubscriber,
	}
}

func (i *PushNotificationEvent) StartSubscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(StreamNotification, DurableNotification),
			nats.DeliverSubject(DeliverNotification),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}
	_, err := i.nats.QueueSubscribe(SubjectNotificationCreated, QueueNotification, opts, i.HandleMessage)
	if err != nil {
		return err
	}
	i.zapLog.Info(fmt.Sprintf("start subscribe from %v", SubjectNotificationCreated))
	return nil
}

// nolint
func (i *PushNotificationEvent) HandleMessage(cxt context.Context, data []byte) (bool, error) {
	// parse []byte to data
	var notification ypb.NatsCreateNotificationRequest
	err := proto.Unmarshal(data, &notification)
	if err != nil {
		i.zapLog.Error(fmt.Sprintf("Notification-Subscribe-Error: [data: %v] [error: %v]", notification.TracingId, err.Error()))
		return false, err
	}

	for _, el := range notification.SendingMethods {
		switch el {
		case "push_notification":
			// validate client id
			if err := ValidateNatsMessage(&notification); err != nil {
				i.zapLog.Error(fmt.Sprintf("Notification-Subscribe-Error: [data: %v] [error: %v]", notification.TracingId, err.Error()))
				return false, err
			}

			err = i.notificationSubscriber.ProcessPushNotification(cxt, &notification)
			if err != nil {
				i.zapLog.Error(fmt.Sprintf("Notification-Subscribe-Error: [data: %v] [error: %v]", notification.TracingId, err.Error()))
				return false, err
			}
		default:
			i.zapLog.Error(fmt.Sprintf("Notification-Subscribe-Error: [data: %v ] [error: unsupported method]", notification.TracingId))
			return false, fmt.Errorf("method is not support")
		}
	}

	// log that subscribe get message
	i.zapLog.Info(fmt.Sprintf("Notification-Subscribe-Success: [%v]", notification.TracingId))
	return false, nil
}
