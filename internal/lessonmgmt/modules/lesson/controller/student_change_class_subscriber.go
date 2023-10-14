package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/consumers"

	"go.uber.org/zap"
)

type StudentChangeClassSubscriber struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (s *StudentChangeClassSubscriber) Subscribe() error {
	s.Logger.Info("[StudentChangeClassSubscriber]: Subscribing to ",
		zap.String("subject", constants.SubjectMasterMgmtClassUpserted),
		zap.String("group", constants.QueueStudentSubscriptionChangeClassEventNats),
		zap.String("durable", constants.DurableStudentSubscriptionChangeClassEventNats),
	)
	return s.subscribeStudentChangeClass()
}

func (s *StudentChangeClassSubscriber) subscribeStudentChangeClass() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamMasterMgmtClass, constants.DurableStudentSubscriptionChangeClassEventNats),
			nats.MaxDeliver(10),
			nats.DeliverSubject(constants.DeliverStudentSubscriptionChangeClassEventNats),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "subscribeStudentChangeClass",
	}
	_, err := s.JSM.QueueSubscribe(
		constants.SubjectMasterMgmtClassUpserted,
		constants.QueueStudentSubscriptionChangeClassEventNats,
		opts,
		s.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectMasterMgmtClassUpserted,
			constants.QueueStudentSubscriptionChangeClassEventNats,
			err,
		)
	}
	return nil
}
