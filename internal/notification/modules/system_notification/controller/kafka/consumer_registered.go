package kafka

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/notification/infra/metrics"

	"go.uber.org/zap"
)

type ConsumersRegistered struct {
	kafkaMgmt kafka.KafkaManagement
	logger    *zap.Logger
	db        database.Ext
	metrics   metrics.NotificationMetrics
}

func NewConsumersRegistered(db database.Ext, kafkaMgmt kafka.KafkaManagement, logger *zap.Logger, metrics metrics.NotificationMetrics) *ConsumersRegistered {
	return &ConsumersRegistered{
		kafkaMgmt: kafkaMgmt,
		logger:    logger,
		db:        db,
		metrics:   metrics,
	}
}

func (r *ConsumersRegistered) Consume() error {
	upsertSystemNotificationConsumer := NewUpsertSystemNotificationConsumer(r.db, r.kafkaMgmt, r.logger, r.metrics)
	err := upsertSystemNotificationConsumer.Consume()
	if err != nil {
		return fmt.Errorf("failed at upsertSystemNotificationConsumer: %+v", err)
	}

	return nil
}
