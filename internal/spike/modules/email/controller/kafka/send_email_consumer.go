package kafka

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/constants"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	spike_consts "github.com/manabie-com/backend/internal/spike/constants"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/consumers"

	"go.uber.org/zap"
)

type SendEmailConsumer struct {
	Logger    *zap.Logger
	KafkaMgmt kafka.KafkaManagement

	consumers.ConsumerHandler
}

func (s *SendEmailConsumer) Consume() error {
	opts := kafka.Option{
		SpanName: "CONSUMER." + constants.EmailSendingTopic,
		KafkaConsumerOptions: []kafka.KafkaConsumerOption{
			kafka.AutoCommit(),
		},
	}

	consumerGroupID := s.KafkaMgmt.GenNewConsumerGroupID(spike_consts.SpikeServiceName, constants.EmailSendingTopic)
	err := s.KafkaMgmt.Consume(constants.EmailSendingTopic, consumerGroupID, opts, s.Handle)
	if err != nil {
		fmt.Printf("Error consuming messages: %v\n", err)
	}
	return nil
}
