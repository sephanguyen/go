package kafka

import (
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	defaultMaxAttempts             = 10
	defaultMaxBytes                = 1e6
	defaultReconnectAttempts       = 10
	defaultWaitingTimeToReconnect  = 15 * time.Second
	defaultRetryLogicAttempts      = 10
	defaultWaitingTimeToRetryLogic = 30 * time.Second
	defaultHeartbeatIntervalTime   = 3 * time.Second
	defaultRebalanceTimeout        = 60 * time.Second
	defaultStrictCommit            = true
)

type kafkaConsumerOption struct {
	kafkaConsumerConfig *kafka.ReaderConfig
	reconnectAttempts   int
	retryLogicAttempts  int

	waitingTimeToReconnect  time.Duration
	waitingTimeToRetryLogic time.Duration

	strictCommit bool
}

type Option struct {
	SpanName             string
	KafkaConsumerOptions []KafkaConsumerOption
}

// nolint:revive
type KafkaConsumerOption interface {
	configureConsumerOption(opts *kafkaConsumerOption) error
}

type kafkaConsumerOptFn func(opts *kafkaConsumerOption) error

func newDefaultKafkaConsumerOption() []KafkaConsumerOption {
	return []KafkaConsumerOption{
		MaxBytes(defaultMaxBytes),
		MaxAttempts(defaultMaxAttempts),
		HeartbeatIntervalTime(defaultHeartbeatIntervalTime),
		RebalanceTimeout(defaultRebalanceTimeout),
		ReconnectAttempts(defaultReconnectAttempts),
		RetryLogicAttempts(defaultRetryLogicAttempts),
		WaitingTimeToReconnect(defaultWaitingTimeToReconnect),
		WaitingTimeToRetryLogic(defaultWaitingTimeToRetryLogic),
		kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
			opts.strictCommit = defaultStrictCommit
			return nil
		}),
	}
}

func (opt kafkaConsumerOptFn) configureConsumerOption(opts *kafkaConsumerOption) error {
	return opt(opts)
}

func brokersOption(n []string) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.kafkaConsumerConfig.Brokers = n
		return nil
	})
}

func topicOption(n string) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.kafkaConsumerConfig.Topic = n
		return nil
	})
}

func consumerGroupIDOption(n string) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.kafkaConsumerConfig.GroupID = n
		return nil
	})
}

// MaxBytes indicates to the broker the maximum batch size that the consumer
// will accept. The broker will truncate a message to satisfy this maximum, so
// choose a value that is high enough for your largest message size.
//
// Default: 1MB
func MaxBytes(n int) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.kafkaConsumerConfig.MaxBytes = n
		return nil
	})
}

// Limit of how many attempts to connect will be made before returning the error.
//
// The default is to try 10 times.
func MaxAttempts(n int) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.kafkaConsumerConfig.MaxAttempts = n
		return nil
	})
}

// RebalanceTimeout optionally sets the length of time the coordinator will wait
// for members to join as part of a rebalance.  For kafka servers under higher
// load, it may be useful to set this value higher.
//
// Default: 60s
func RebalanceTimeout(n time.Duration) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.kafkaConsumerConfig.RebalanceTimeout = n
		return nil
	})
}

// HeartbeatInterval sets the optional frequency at which the reader sends the consumer
// group heartbeat update.
//
// Default: 3s
func HeartbeatIntervalTime(n time.Duration) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.kafkaConsumerConfig.HeartbeatInterval = n
		return nil
	})
}

// Limit of how many attempts to reconnect to Kafka cluster.
//
// The default is to try 5 times.
func ReconnectAttempts(n int) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.reconnectAttempts = n
		return nil
	})
}

// Limit of how many attempts to retry handle a message if it's failed.
//
// The default is to try 10 times.
func RetryLogicAttempts(n int) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.retryLogicAttempts = n
		return nil
	})
}

// Waiting time before trying to reconnect to Kafka cluster.
//
// Default: 10s
func WaitingTimeToReconnect(n time.Duration) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.waitingTimeToReconnect = n
		return nil
	})
}

// Waiting time before trying to retry handle a message if it's failed.
//
// Default: 30s
func WaitingTimeToRetryLogic(n time.Duration) KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.waitingTimeToRetryLogic = n
		return nil
	})
}

// By default (doesn't use this option) ONLY COMMIT message when it has been processed successfully.
// Meaning the consumer will be stuck at the failed processed message.
//
// If set this option, the message will be COMMITTED AUTOMATICALLY even if failed handle, and process the next message.
func AutoCommit() KafkaConsumerOption {
	return kafkaConsumerOptFn(func(opts *kafkaConsumerOption) error {
		opts.strictCommit = false
		return nil
	})
}
