package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	vc_consumers "github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/application/consumers"

	"go.uber.org/zap"
)

type LiveRoomSubscriber struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	vc_consumers.SubscriberHandler
}

func (l *LiveRoomSubscriber) Subscribe() error {
	l.Logger.Info("[LiveRoomEvent]: Subscribing to ",
		zap.String("subject", constants.SubjectLiveRoomUpdated),
		zap.String("group", constants.QueueLiveRoom),
		zap.String("durable", constants.DurableLiveRoom))

	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamLiveRoom, constants.DurableLiveRoom),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverLiveRoom),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "LiveRoomSubscription",
	}

	_, err := l.JSM.QueueSubscribe(
		constants.SubjectLiveRoomUpdated,
		constants.QueueLiveRoom,
		opts,
		l.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectLiveRoomUpdated,
			constants.QueueLiveRoom,
			err,
		)
	}
	return nil
}
