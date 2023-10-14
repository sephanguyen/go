package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/consumers"

	"go.uber.org/zap"
)

type CreateLiveLessonRoomSubscriber struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (c *CreateLiveLessonRoomSubscriber) Subscribe() error {
	c.Logger.Info("[CreateLiveLessonRoomEvent]: Subscribing to ",
		zap.String("subject", constants.SubjectLessonCreated),
		zap.String("group", constants.QueueCreateLiveLessonRoom),
		zap.String("durable", constants.DurableCreateLiveLessonRoom))

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLesson, constants.DurableCreateLiveLessonRoom),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverCreateLiveLessonRoom),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "CreateLiveLessonRoomSubscription",
	}

	_, err := c.JSM.QueueSubscribe(
		constants.SubjectLessonCreated,
		constants.QueueCreateLiveLessonRoom,
		opts,
		c.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectLessonCreated,
			constants.QueueCreateLiveLessonRoom,
			err,
		)
	}
	return nil
}
