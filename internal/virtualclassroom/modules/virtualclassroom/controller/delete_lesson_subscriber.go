package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/consumers"

	"go.uber.org/zap"
)

type LessonDeletedSubscription struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.LessonDeletedHandler
}

func (l *LessonDeletedSubscription) Subscribe() error {
	l.Logger.Info("[DeleteLessonEvent]: Subscribing to ",
		zap.String("subject", constants.SubjectLessonDeleted),
		zap.String("group", constants.QueueLessonDeleted),
		zap.String("durable", constants.DurableLessonDeleted))

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLesson, constants.DurableLessonDeleted),
			nats.DeliverSubject(constants.DeliverLessonDeleted),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "DeleteLessonSubscription",
	}
	_, err := l.JSM.QueueSubscribe(constants.SubjectLessonDeleted, constants.QueueLessonDeleted, opts, l.Handle)
	if err != nil {
		return fmt.Errorf("s.JSM.Subscribe: %w", err)
	}
	return nil
}
