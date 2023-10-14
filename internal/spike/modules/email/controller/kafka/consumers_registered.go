package kafka

import (
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/sendgrid"
	"github.com/manabie-com/backend/internal/spike/modules/email/application/consumers"
	"github.com/manabie-com/backend/internal/spike/modules/email/infrastructure/repositories"

	"go.uber.org/zap"
)

type ConsumersRegistered struct {
	db             database.Ext
	kafkaMgmt      kafka.KafkaManagement
	logger         *zap.Logger
	sendGridClient sendgrid.SendGridClient
}

func NewConsumersRegistered(
	db database.Ext,
	kafkaMgmt kafka.KafkaManagement,
	sendGridClient sendgrid.SendGridClient,
	logger *zap.Logger,
) *ConsumersRegistered {
	return &ConsumersRegistered{
		db:             db,
		sendGridClient: sendGridClient,
		kafkaMgmt:      kafkaMgmt,
		logger:         logger,
	}
}

func (r *ConsumersRegistered) Consume() error {
	// Init send email consumer
	sendEmailHandler := &consumers.SendEmailHandler{
		DB:                 r.db,
		SendGridClient:     r.sendGridClient,
		EmailRepo:          &repositories.EmailRepo{},
		EmailRecipientRepo: &repositories.EmailRecipientRepo{},
	}
	c := &SendEmailConsumer{
		KafkaMgmt:       r.kafkaMgmt,
		ConsumerHandler: sendEmailHandler,
	}

	return c.Consume()
}
