package kafka

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	errHandler = errors.New("kafka: consumer handler fail to process message")
)

const (
	kafkaUserIDHeaderName       = "Kafka-User-ID"
	kafkaResourcePathHeaderName = "Kafka-Resource-Path"
	kafkaSpanHeaderName         = "Kafka-Span-Name"
)

// The first return value is retry mechanism (if error has appeared)
// if error != nil && retry then retry to call this handler function after a waiting time and do it with n attempts value, after that, commit message
// else commit message, and do not retry
type MsgHandler func(ctx context.Context, data []byte) (bool, error)

// nolint:revive
type KafkaManagement interface {
	GetObjectNamePrefix() string

	// GenNewConsumerGroupID return a consumer_group_id for a service, topic name here shouldn't have prefix.
	// The example name: prod.tokyo.notificationmgmt.consumer-group.topic-A.
	GenNewConsumerGroupID(serviceName string, topicName string) string

	// Trying to connect Kafka cluster.
	ConnectToKafka()

	// Value should be an returned value of json.Marshal(v), supporting debug, and check payload.
	PublishContext(ctx context.Context, topic string, key []byte, value []byte) error

	// Value should be an returned value of json.Marshal(v), supporting debug, and check payload.
	// This method will support tracing with Jaeger.
	TracedPublishContext(ctx context.Context, spanName string, topic string, key []byte, value []byte) error

	// This will create a consumer of a consumer group to consume this topic.
	// Noted: Please use GenNewConsumerGroupID to generate consumerGroupID.
	Consume(topic string, consumerGroupID string, option Option, handleMsg MsgHandler) error
	UpsertTopic(topicConfig *kafka.TopicConfig) error

	Close()
}

// map group_id consumer
type consumerGroup struct {
	consumerGroupID string
	consumers       []consumer
}

type kafkaManagementImpl struct {
	address string

	// This is prefix name of any object in kafka: topic, consumer, consumer group, etc...
	objectNamePrefix string

	logger            *zap.Logger
	isLocalEnv        bool
	mutex             *sync.Mutex
	consumerWaitGroup *sync.WaitGroup

	// Connection, supporting for handle everything related to metadata: upsert topic, get topic,...
	conn Conn

	kafkaer Kafkaer

	// Map topic name to a list of consumer groups are running in current service.
	consumerGroupsMap map[string][]*consumerGroup

	// Map topic name to each writer.
	writerMap map[string]Writer
}

func NewKafkaManagement(address string, isLocalEnv bool, objectNamePrefix string, zapLogger *zap.Logger) (KafkaManagement, error) {
	if address == "" {
		return nil, errors.New("kafka: missing address")
	}

	if zapLogger == nil {
		return nil, errors.New("kafka: missing logger")
	}

	if objectNamePrefix == "" {
		zapLogger.Sugar().Warnf("prefix of kafka object name is empty")
	}

	k := &kafkaManagementImpl{
		logger:            zapLogger,
		address:           address,
		isLocalEnv:        isLocalEnv,
		objectNamePrefix:  objectNamePrefix,
		consumerWaitGroup: &sync.WaitGroup{},
		writerMap:         make(map[string]Writer),
		consumerGroupsMap: make(map[string][]*consumerGroup),
		mutex:             &sync.Mutex{},
		kafkaer:           newKafkaImpl(),
	}

	return k, nil
}

func (k *kafkaManagementImpl) isKafkaConnConneted() bool {
	if k.conn == nil {
		return false
	}

	// Trying to get API version - meta data (ping to kafka)
	_, err := k.conn.ApiVersions()
	if err != nil {
		k.logger.Error("kafka: can't ping kafka cluster", zap.Error(err))
		return false
	}
	return true
}

func (k *kafkaManagementImpl) getKafkaConnection() Conn {
	k.mutex.Lock()
	if !k.isKafkaConnConneted() {
		k.ConnectToKafka()
	}
	k.mutex.Unlock()

	return k.conn
}

func (k *kafkaManagementImpl) GetObjectNamePrefix() string {
	return k.objectNamePrefix
}

// GenNewConsumerGroupID return a consumer_group_id for a service, topic name here shouldn't have prefix.
// The example name: prod.tokyo.notificationmgmt.consumer-group.topic-A
func (k *kafkaManagementImpl) GenNewConsumerGroupID(serviceName string, topicName string) string {
	return fmt.Sprintf("%s%s.consumer-group.%s", k.objectNamePrefix, serviceName, topicName)
}

func (k *kafkaManagementImpl) ConnectToKafka() {
	conn, err := k.kafkaer.DialConn("tcp", k.address)
	if err != nil {
		k.logger.Fatal(err.Error())
	}

	k.conn = conn

	k.logger.Info("kafka connected")
}

func (k *kafkaManagementImpl) PublishContext(ctx context.Context, topic string, key []byte, value []byte) error {
	if topic == "" {
		k.logger.Error("kafka: empty topic to publish")
		return fmt.Errorf("kafka: empty topic to publish")
	}

	// IMPORTANT: re-define the topic name depend on prefix (environment, cluster name)
	// Ex: topic_name = "topic-A"
	// On production of tokyo cluster, the name of this topic should be: "prod.tokyo.topic-A"
	topic = GetTopicNameWithPrefix(topic, k.objectNamePrefix)

	if _, ok := k.writerMap[topic]; !ok {
		brokers := []string{k.address}
		k.writerMap[topic] = k.kafkaer.NewWriter(kafka.WriterConfig{
			Brokers:     brokers,
			Topic:       topic,
			Balancer:    &kafka.LeastBytes{},
			MaxAttempts: math.MaxInt32,
		})
	}

	writer := k.writerMap[topic]

	headers := MessageHeadersFromContext(ctx, false)

	return writer.WriteMessages(ctx, kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
	})
}

func (k *kafkaManagementImpl) TracedPublishContext(ctx context.Context, spanName string, topic string, key []byte, value []byte) error {
	ctx, span := interceptors.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()
	span.SetAttributes(attribute.KeyValue{
		Key:   "message_payload",
		Value: attribute.StringValue(string(value)),
	})
	if topic == "" {
		k.logger.Error("kafka: empty topic to publish")
		return fmt.Errorf("kafka: empty topic to publish")
	}

	// IMPORTANT: re-define the topic name depend on prefix (environment, cluster name)
	// Ex: topic_name = "topic-A"
	// On production of tokyo cluster, the name of this topic should be: "prod.tokyo.topic-A"
	topic = GetTopicNameWithPrefix(topic, k.objectNamePrefix)

	if _, ok := k.writerMap[topic]; !ok {
		brokers := []string{k.address}
		k.writerMap[topic] = k.kafkaer.NewWriter(kafka.WriterConfig{
			Brokers:     brokers,
			Topic:       topic,
			Balancer:    &kafka.LeastBytes{},
			MaxAttempts: math.MaxInt32,
		})
	}

	writer := k.writerMap[topic]

	headers := MessageHeadersFromContext(ctx, true)

	return writer.WriteMessages(ctx, kafka.Message{
		Key:     key,
		Value:   value,
		Headers: headers,
	})
}

func (k *kafkaManagementImpl) Consume(topic string, consumerGroupID string, option Option, handleMsg MsgHandler) error {
	if topic == "" {
		k.logger.Error("kafka: empty topic to consume")
		return fmt.Errorf("kafka: empty topic to consume")
	}

	// IMPORTANT: re-define the topic name depend on prefix (environment, cluster name)
	// Ex: topic_name = "topic-A"
	// On production of tokyo cluster, the name of this topic should be: "prod.tokyo.topic-A"
	topic = GetTopicNameWithPrefix(topic, k.objectNamePrefix)

	if consumerGroupID == "" {
		k.logger.Error("kafka: empty consumer_group_id to consume")
		return fmt.Errorf("kafka: empty consumer_group_id to consume")
	}

	if _, ok := k.consumerGroupsMap[topic]; !ok {
		k.consumerGroupsMap[topic] = make([]*consumerGroup, 0)
	}

	consumerGroups := k.consumerGroupsMap[topic]
	var currentConsumerGroup *consumerGroup
	for _, consumerGroup := range consumerGroups {
		if consumerGroup.consumerGroupID == consumerGroupID {
			currentConsumerGroup = consumerGroup
		}
	}

	if currentConsumerGroup == nil {
		// Create a new consumer group in memory of current service if not exists
		currentConsumerGroup = &consumerGroup{
			consumerGroupID: consumerGroupID,
			consumers:       make([]consumer, 0),
		}
		k.consumerGroupsMap[topic] = append(k.consumerGroupsMap[topic], currentConsumerGroup)
	}

	// Set default options, and get user options
	option.KafkaConsumerOptions = append(
		newDefaultKafkaConsumerOption(),
		append([]KafkaConsumerOption{brokersOption([]string{k.address}), topicOption(topic), consumerGroupIDOption(consumerGroupID)}, option.KafkaConsumerOptions...)...,
	)

	consumerOpts := kafkaConsumerOption{kafkaConsumerConfig: &kafka.ReaderConfig{}}
	for _, v := range option.KafkaConsumerOptions {
		if err := v.configureConsumerOption(&consumerOpts); err != nil {
			return err
		}
	}

	kafkaReader := k.kafkaer.NewReader(*consumerOpts.kafkaConsumerConfig)
	currentConsumerGroup.consumers = append(currentConsumerGroup.consumers, newConsumer(
		consumerOpts,
		kafkaReader,
		false,
		k.logger,
	))

	for _, consumerMember := range currentConsumerGroup.consumers {
		k.mutex.Lock()
		if !consumerMember.getRunning() {
			consumerMember.setRunning(true)
			k.mutex.Unlock()
			// Run a go-routine to consume message
			k.consumerWaitGroup.Add(1)
			go func(consumerMember consumer, handleMsg MsgHandler) {
				defer func() {
					k.consumerWaitGroup.Done()
					k.mutex.Lock()
					consumerMember.setRunning(false)
					k.mutex.Unlock()
				}()
				k.runConsumer(option.SpanName, consumerMember, handleMsg)
			}(consumerMember, handleMsg)
		} else {
			k.mutex.Unlock()
		}
	}

	return nil
}

func (k *kafkaManagementImpl) runConsumer(spanName string, consumerMember consumer, handleMsg MsgHandler) {
	consumerCtx, _ := consumerMember.getContextWithCancel()
	for {
		select {
		case <-consumerCtx.Done():
			k.logger.Warn("kafka: shuting down kafka consumer go-routine")
			return
		default:
			var msg = &kafka.Message{}
			var err error
			// Read a message from Kafka.
			err = consumerMember.readMessage(msg)
			if err != nil {
				// Can't read message after retry -> cancel this go routine.
				k.logger.Error("kafka: error consumer.readMessage", zap.Error(err))
				k.logger.Warn("kafka: shuting down kafka consumer go-routine")
				return
			}

			// Handle logic for received message by client handler.
			isContinueToNextMsg, err := consumerMember.handleMessage(spanName, handleMsg, msg)
			if err != nil {
				k.logger.Error("kafka: error consumer.handleMessage", zap.Error(err))
				if !isContinueToNextMsg {
					// Can't process message after retry -> close reader and go-routine.
					k.logger.Warn("shutting down consumer (for strict mode)")
					return
				}
			}

			// Commit message in case manual commit (use reader.FetchMessage).
			err = consumerMember.completeMessage(msg)
			if err != nil {
				// Can't commit message after retry -> cancel this go routine
				k.logger.Error("kafka: error consumer.completeMessage", zap.Error(err))
				k.logger.Warn("kafka: shuting down kafka consumer go-routine")
				return
			}
		}
	}
}

func (k *kafkaManagementImpl) UpsertTopic(cfg *kafka.TopicConfig) error {
	if k.isLocalEnv {
		cfg.ReplicationFactor = 1
	}

	conn := k.getKafkaConnection()

	err := conn.CreateTopics(*cfg)
	if err != nil {
		k.logger.Error("kafka: failed to create topic "+cfg.Topic, zap.Error(err))
		return err
	}
	return nil
}

func (k *kafkaManagementImpl) Close() {
	k.logger.Warn("kafka: shutting down KafkaManagement...")
	// Close all writers
	k.logger.Warn("kafka: shutting down all kafka writers...")
	for _, writer := range k.writerMap {
		err := writer.Close()
		if err != nil {
			k.logger.Error("kafka: failed to close writer", zap.Error(err))
		}
	}

	// Close all consumers go-routine
	readers := make([]Reader, 0)
	for _, consumerGroups := range k.consumerGroupsMap {
		for _, consumerGroup := range consumerGroups {
			for _, consumer := range consumerGroup.consumers {
				// remove flag is_running
				consumer.setRunning(false)
				// cancel go routine
				_, cancel := consumer.getContextWithCancel()
				cancel()
				// close reader
				readers = append(readers, consumer.getReader())
			}
		}
	}
	// Waiting for all go-routine has shutdown.
	k.consumerWaitGroup.Wait()

	// Close all readers.
	k.logger.Warn("kafka: shutting down all kafka readers...")
	for _, reader := range readers {
		err := reader.Close()
		if err != nil {
			k.logger.Error("kafka: failed to close reader", zap.Error(err))
		}
	}

	err := k.conn.Close()
	if err != nil {
		k.logger.Error("kafka: failed to close leader connection", zap.Error(err))
	}
}
