package kafka

import (
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
)

func TestConsumerOption(t *testing.T) {
	t.Run("Create consumer with common config", func(t *testing.T) {
		address := "127.0.0.1:9092"
		topic := "topic"
		consumerGroupID := "consumer-group-id"

		option := Option{
			KafkaConsumerOptions: []KafkaConsumerOption{
				brokersOption([]string{address}),
				topicOption(topic),
				consumerGroupIDOption(consumerGroupID),
				MaxBytes(defaultMaxBytes),
				MaxAttempts(defaultMaxAttempts),
				HeartbeatIntervalTime(defaultHeartbeatIntervalTime),
				RebalanceTimeout(defaultRebalanceTimeout),
				ReconnectAttempts(defaultReconnectAttempts),
				RetryLogicAttempts(defaultRetryLogicAttempts),
				WaitingTimeToReconnect(defaultWaitingTimeToReconnect),
				WaitingTimeToRetryLogic(defaultWaitingTimeToRetryLogic),
				AutoCommit(),
			},
		}

		consumerConfig := &kafka.ReaderConfig{}
		o := kafkaConsumerOption{kafkaConsumerConfig: consumerConfig}

		for _, v := range option.KafkaConsumerOptions {
			if err := v.configureConsumerOption(&o); err != nil {
				t.Errorf("error when configureSubscribeOption: %v", err)
			}
		}

		assert.Equal(t, []string{address}, o.kafkaConsumerConfig.Brokers)
		assert.Equal(t, topic, o.kafkaConsumerConfig.Topic)
		assert.Equal(t, consumerGroupID, o.kafkaConsumerConfig.GroupID)
		assert.Equal(t, int(defaultMaxBytes), o.kafkaConsumerConfig.MaxBytes)
		assert.Equal(t, defaultMaxAttempts, o.kafkaConsumerConfig.MaxAttempts)
		assert.Equal(t, defaultHeartbeatIntervalTime, o.kafkaConsumerConfig.HeartbeatInterval)
		assert.Equal(t, defaultRebalanceTimeout, o.kafkaConsumerConfig.RebalanceTimeout)
		assert.Equal(t, defaultReconnectAttempts, o.reconnectAttempts)
		assert.Equal(t, defaultRetryLogicAttempts, o.retryLogicAttempts)
		assert.Equal(t, defaultWaitingTimeToReconnect, o.waitingTimeToReconnect)
		assert.Equal(t, defaultWaitingTimeToRetryLogic, o.waitingTimeToRetryLogic)
		assert.Equal(t, false, o.strictCommit)
	})
}
