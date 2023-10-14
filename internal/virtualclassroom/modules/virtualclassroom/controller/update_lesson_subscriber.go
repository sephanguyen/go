package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/consumers"

	"go.uber.org/zap"
)

type LessonUpdatedSubscription struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (l *LessonUpdatedSubscription) Subscribe() error {
	l.Logger.Info("[UpdateLessonEvent]: Subscribing to ",
		zap.String("subject", constants.SubjectLessonUpdated),
		zap.String("group", constants.QueueLessonUpdated),
		zap.String("durable", constants.DurableLessonUpdated))

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLesson, constants.DurableLessonUpdated),
			nats.DeliverSubject(constants.DeliverLessonUpdated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "UpdateLessonSubscription",
	}
	_, err := l.JSM.QueueSubscribe(constants.SubjectLessonUpdated, constants.QueueLessonUpdated, opts, l.Handle)
	if err != nil {
		return fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	return nil
}
