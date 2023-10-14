package nats

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/application/consumers"
	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"go.uber.org/zap"
)

type UserCreatedSubscriber struct {
	nats   nats.JetStreamManagement
	logger *zap.Logger
	consumers.ConsumerHandler
}

func NewUserCreatedSubscriber(nats nats.JetStreamManagement, logger *zap.Logger) *UserCreatedSubscriber {
	return &UserCreatedSubscriber{
		nats:            nats,
		logger:          logger,
		ConsumerHandler: &consumers.UserCreatedHandler{},
	}
}

func (s *UserCreatedSubscriber) StartSubscribe() error {
	opts := nats.Option{
		JetStreamOptions: []nats.JSSubOption{
			nats.ManualAck(),
			nats.Bind(constants.StreamUser, constants.DurableConversationMgmtUserCreated),
			nats.DeliverSubject(constants.DeliverConversationMgmtUserCreated),
			nats.MaxDeliver(10),
			nats.AckWait(30 * time.Second),
		},
	}
	_, err := s.nats.QueueSubscribe(constants.SubjectUserCreated, constants.QueueConversationMgmtUserCreated, opts, s.Handle)
	if err != nil {
		return err
	}
	s.logger.Info(fmt.Sprintf("start subscribe from %v", constants.SubjectUserCreated))
	return nil
}
