package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/class/application/consumers"

	"go.uber.org/zap"
)

type StudentPackageSubscriber struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (s *StudentPackageSubscriber) Subscribe() error {
	s.Logger.Info("[StudentPackageSubscriber]: Subscribing to ",
		zap.String("subject", constants.SubjectStudentPackageV2EventNats),
		zap.String("group", constants.QueueStudentSubscriptionClassEventNats),
		zap.String("durable", constants.DurableStudentSubscriptionClassEventNats),
	)
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStudentPackageEventNatsV2, constants.DurableStudentSubscriptionClassEventNats),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverStudentSubscriptionClassEventNats),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "StudentPackageV2",
	}
	_, err := s.JSM.QueueSubscribe(
		constants.SubjectStudentPackageV2EventNats,
		constants.QueueStudentSubscriptionClassEventNats,
		opts,
		s.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectStudentPackageV2EventNats,
			constants.QueueStudentSubscriptionClassEventNats,
			err,
		)
	}
	return nil
}
