package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/consumers"

	"go.uber.org/zap"
)

type LessonDefaultChatStateSubscriber struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (l *LessonDefaultChatStateSubscriber) Subscribe() error {
	l.Logger.Info("[LessonDefaultChatStateEvent]: Subscribing to ",
		zap.String("subject", constants.SubjectLessonUpdated),
		zap.String("group", constants.QueueLessonDefaultChatState),
		zap.String("durable", constants.DurableLessonDefaultChatState))

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLesson, constants.DurableLessonDefaultChatState),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverLessonDefaultChatState),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "LessonDefaultChatStateSubscription",
	}

	_, err := l.JSM.QueueSubscribe(
		constants.SubjectLessonUpdated,
		constants.QueueLessonDefaultChatState,
		opts,
		l.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectLessonUpdated,
			constants.QueueLessonDefaultChatState,
			err,
		)
	}
	return nil
}
