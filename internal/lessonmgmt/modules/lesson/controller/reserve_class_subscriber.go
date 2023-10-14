package controller

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/application/consumers"

	"go.uber.org/zap"
)

type ReserveClassSubscriber struct {
	Logger *zap.Logger
	JSM    nats.JetStreamManagement
	consumers.SubscriberHandler
}

func (s *ReserveClassSubscriber) Subscribe() error {
	s.Logger.Info("[ReserveClassSubscriber]: Subscribing to ",
		zap.String("subject", constants.SubjectMasterMgmtReserveClassUpserted),
		zap.String("group", constants.QueueMasterMgmtReserveClassUpserted),
		zap.String("durable", constants.DurableMasterMgmtReserveClassUpserted),
	)
	return s.subscribeReserveClassEvt()
}

func (s *ReserveClassSubscriber) subscribeReserveClassEvt() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamMasterMgmtReserveClass, constants.DurableMasterMgmtReserveClassUpserted),
			nats.MaxDeliver(10),
			nats.DeliverNew(),
			nats.DeliverSubject(constants.DeliverMasterMgmtReserveClassEvent),
			nats.AckWait(30 * time.Second),
		},
		SpanName: "subscribeReserveClassEvt",
	}
	_, err := s.JSM.QueueSubscribe(
		constants.SubjectMasterMgmtReserveClassUpserted,
		constants.QueueMasterMgmtReserveClassUpserted,
		opts,
		s.Handle,
	)
	if err != nil {
		return fmt.Errorf("error subscribing to subject `%s` on `%s` queue: %w",
			constants.SubjectMasterMgmtReserveClassUpserted,
			constants.QueueMasterMgmtReserveClassUpserted,
			err,
		)
	}
	return nil
}
