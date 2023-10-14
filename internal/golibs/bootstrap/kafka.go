package bootstrap

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/kafka"

	"go.uber.org/zap"
)

// KafkaServicer represents a service that uses Kafka.
// Users should implements this interface with their server struct
// when they want to use Kafka.
type KafkaServicer[T any] interface {
	// InitConsumers should be implemented by users to use Kafka.
	// It should register all the necessary consumer to kafka topic.
	InitKafkaConsumers(context.Context, T, *Resources) error
}

func initKafka(_ context.Context, config interface{}, rsc *Resources) error {
	c, err := extract[configs.KafkaClusterConfig](config, kafkaFieldName)
	if err != nil {
		return ignoreErrFieldNotFound(err)
	}

	_ = rsc.WithKafkaC(c)
	return nil
}

// Kafkaer handles the connection to Kafka server.
type Kafkaer interface {
	// NewKafkaManagement returns a new kafka.NewKafkaManagement instance.
	NewKafkaManagement(zapLogger *zap.Logger, c *configs.KafkaClusterConfig) (kafka.KafkaManagement, error)
}

// kafkaImpl implements Kafkaer.
type kafkaImpl struct{}

func newKafkaImpl() *kafkaImpl {
	return &kafkaImpl{}
}

func (n *kafkaImpl) NewKafkaManagement(zapLogger *zap.Logger, c *configs.KafkaClusterConfig) (kafka.KafkaManagement, error) {
	return kafka.NewKafkaManagement(c.Address, c.IsLocal, c.ObjectNamePrefix, zapLogger)
}
