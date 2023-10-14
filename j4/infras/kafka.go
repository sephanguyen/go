package infras

import (
	"github.com/manabie-com/backend/internal/golibs/kafka"
	"github.com/manabie-com/backend/internal/golibs/logger"
)

func (c *Connections) ConnectKafka(cfg *ManabieJ4Config) error {
	if c.Kafka == nil {
		logger := logger.NewZapLogger("info", cfg.KafkaConfig.IsLocal)
		kafka, err := kafka.NewKafkaManagement(cfg.KafkaConfig.Address, cfg.KafkaConfig.IsLocal, cfg.KafkaConfig.ObjectNamePrefix, logger)
		if err != nil {
			return err
		}
		kafka.ConnectToKafka()

		c.Kafka = kafka
	}
	return nil
}
