package nats

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/application/consumers"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"go.uber.org/zap"
)

type StaffUpsertedSubscriber struct {
	nats   nats.JetStreamManagement
	logger *zap.Logger
	consumers.ConsumerHandler
}

func NewStaffUpsertedSubscriber(nats nats.JetStreamManagement, logger *zap.Logger) *StaffUpsertedSubscriber {
	return &StaffUpsertedSubscriber{
		nats:            nats,
		logger:          logger,
		ConsumerHandler: &consumers.StaffUpsertedHandler{},
	}
}

func (s *StaffUpsertedSubscriber) StartSubscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamStaff, constants.DurableConversationMgmtUpsertStaff),
			nats.DeliverSubject(constants.DeliverConversationMgmtUpsertStaff),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}
	_, err := s.nats.QueueSubscribe(constants.SubjectUpsertStaff, constants.QueueConversationMgmtUpsertStaff, opts, s.Handle)
	if err != nil {
		return err
	}
	s.logger.Info(fmt.Sprintf("start subscribe from %v", constants.SubjectUpsertStaff))
	return nil
}
