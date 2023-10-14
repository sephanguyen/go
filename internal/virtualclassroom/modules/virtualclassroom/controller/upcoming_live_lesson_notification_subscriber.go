package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/consumers"

	"go.uber.org/zap"
)

type UpcomingLiveLessonNotificationSubscription struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (u *UpcomingLiveLessonNotificationSubscription) Subscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUpcomingLiveLessonNotification, constants.DurableUpcomingLiveLessonNotification),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverUpcomingLiveLessonNotification),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "UpcomingLiveLessonNotification",
	}
	_, err := u.JSM.QueueSubscribe(
		constants.SubjectUpcomingLiveLessonNotification,
		constants.QueueUpcomingLiveLessonNotification,
		opts,
		u.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectUpcomingLiveLessonNotification,
			constants.QueueUpcomingLiveLessonNotification,
			err,
		)
	}
	return nil
}
