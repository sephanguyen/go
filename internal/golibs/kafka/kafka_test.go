package kafka

import (
	"context"
	"fmt"
	"math"
	"sync"
	"testing"

	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	dummyError = fmt.Errorf("dummy error")
)

func TestNewKafkaManagement(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		_, err := NewKafkaManagement("address", true, "obj_prefix", zap.NewExample())
		assert.NoError(t, err)
	})

	t.Run("fail due to miss address", func(t *testing.T) {
		_, err := NewKafkaManagement("", true, "obj_prefix", zap.NewExample())
		assert.Equal(t, "kafka: missing address", err.Error())
	})

	t.Run("fail due to miss logger", func(t *testing.T) {
		_, err := NewKafkaManagement("address", false, "obj_prefix", nil)
		assert.Equal(t, "kafka: missing logger", err.Error())
	})
}

func Test_isKafkaLeaderConnConneted(t *testing.T) {
	mockConn := NewMockConn(t)

	k := kafkaManagementImpl{
		conn:   mockConn,
		logger: zap.NewExample(),
	}

	t.Run("happy case", func(t *testing.T) {
		mockConn.On("ApiVersions").Once().Return([]kafka.ApiVersion{}, nil)
		isConnected := k.isKafkaConnConneted()

		assert.Equal(t, isConnected, true)
	})

	t.Run("error get metadata", func(t *testing.T) {
		mockConn.On("ApiVersions").Once().Return(nil, fmt.Errorf("network error"))
		isConnected := k.isKafkaConnConneted()

		assert.Equal(t, isConnected, false)
	})

	t.Run("nil conn", func(t *testing.T) {
		k.conn = nil
		isConnected := k.isKafkaConnConneted()

		assert.Equal(t, isConnected, false)
	})
}

func TestGenNewConsumerGroupID(t *testing.T) {
	k := kafkaManagementImpl{
		objectNamePrefix: "local.manabie.",
	}
	t.Run("fail due to miss address", func(t *testing.T) {
		consumerGroupID := k.GenNewConsumerGroupID("spike", "topic-A")
		assert.Equal(t, consumerGroupID, "local.manabie.spike.consumer-group.topic-A")
	})
}

func TestConnectToKafka(t *testing.T) {
	mockKafkaer := NewMockKafkaer(t)

	k := kafkaManagementImpl{
		objectNamePrefix: "local.manabie.",
		address:          "127.0.0.1:9092",
		logger:           zap.NewExample(),
		mutex:            &sync.Mutex{},
		kafkaer:          mockKafkaer,
	}

	t.Run("happy case", func(t *testing.T) {
		mockConn := NewMockConn(t)
		mockKafkaer.On("DialConn", "tcp", k.address).Return(mockConn, nil)

		k.ConnectToKafka()
	})
}

func TestPublishContext(t *testing.T) {
	mockKafkaer := NewMockKafkaer(t)

	k := kafkaManagementImpl{
		objectNamePrefix: "local.manabie.",
		address:          "127.0.0.1:9092",
		logger:           zap.NewExample(),
		mutex:            &sync.Mutex{},
		kafkaer:          mockKafkaer,
		writerMap:        make(map[string]Writer),
	}

	topicName := "example-topic"

	userID := "user-id"
	resourcePath := "resource-path"
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserID:       userID,
			ResourcePath: resourcePath,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(context.Background(), claim)

	t.Run("happy case", func(t *testing.T) {
		mockWriter := NewMockWriter(t)
		mockKafkaer.On("NewWriter", kafka.WriterConfig{
			Brokers:     []string{k.address},
			Topic:       "local.manabie." + topicName,
			Balancer:    &kafka.LeastBytes{},
			MaxAttempts: math.MaxInt32,
		}).Return(mockWriter)

		mockWriter.On("WriteMessages", ctx, mock.Anything).Return(nil)

		err := k.PublishContext(ctx, topicName, []byte("key"), []byte("value"))

		assert.NoError(t, err)
	})

	t.Run("empty topic", func(t *testing.T) {
		err := k.PublishContext(ctx, "", []byte("key"), []byte("value"))

		assert.Error(t, err, fmt.Errorf("kafka: empty topic to publish"))
	})
}

func TestTracedPublishContext(t *testing.T) {
	mockKafkaer := NewMockKafkaer(t)

	k := kafkaManagementImpl{
		objectNamePrefix: "local.manabie.",
		address:          "127.0.0.1:9092",
		logger:           zap.NewExample(),
		mutex:            &sync.Mutex{},
		kafkaer:          mockKafkaer,
		writerMap:        make(map[string]Writer),
	}

	topicName := "example-topic"

	userID := "user-id"
	resourcePath := "resource-path"
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			UserID:       userID,
			ResourcePath: resourcePath,
		},
	}
	ctx := interceptors.ContextWithJWTClaims(context.Background(), claim)

	spanName := "unit-test-span"

	t.Run("happy case", func(t *testing.T) {
		mockWriter := NewMockWriter(t)
		mockKafkaer.On("NewWriter", kafka.WriterConfig{
			Brokers:     []string{k.address},
			Topic:       "local.manabie." + topicName,
			Balancer:    &kafka.LeastBytes{},
			MaxAttempts: math.MaxInt32,
		}).Return(mockWriter)

		ctxCheck, span := interceptors.StartSpan(context.Background(), spanName, trace.WithSpanKind(trace.SpanKindProducer))
		defer span.End()
		mockWriter.On("WriteMessages", ctxCheck, mock.Anything).Return(nil)

		err := k.TracedPublishContext(context.Background(), spanName, topicName, []byte("key"), []byte("value"))

		assert.NoError(t, err)
	})

	t.Run("empty topic", func(t *testing.T) {
		err := k.TracedPublishContext(ctx, "", "", []byte("key"), []byte("value"))

		assert.Error(t, err, fmt.Errorf("kafka: empty topic to publish"))
	})
}

func TestUpsertTopic(t *testing.T) {
	mockConn := NewMockConn(t)

	k := kafkaManagementImpl{
		conn:   mockConn,
		logger: zap.NewExample(),
		mutex:  &sync.Mutex{},
	}

	topicName := "example-topic"

	t.Run("happy case", func(t *testing.T) {
		mockConn.On("ApiVersions").Once().Return([]kafka.ApiVersion{}, nil)
		mockConn.On("CreateTopics", kafka.TopicConfig{
			Topic: topicName,
		}).Once().Return(nil)
		err := k.UpsertTopic(&kafka.TopicConfig{
			Topic: topicName,
		})

		assert.NoError(t, err)
	})
}

func TestClose(t *testing.T) {
	mockKafkaer := NewMockKafkaer(t)
	mockConn := NewMockConn(t)

	numReaders := 10
	numWriters := 5

	mockReaders := make([]*MockReader, 0)
	mockWriters := make([]*MockWriter, 0)
	mockConsumers := make([]*mockConsumer, 0)

	k := kafkaManagementImpl{
		objectNamePrefix: "local.manabie.",
		address:          "127.0.0.1:9092",
		logger:           zap.NewExample(),
		mutex:            &sync.Mutex{},
		kafkaer:          mockKafkaer,
		writerMap:        make(map[string]Writer),
		consumerGroupsMap: map[string][]*consumerGroup{
			"local.manabie.example-topic": {
				{
					consumerGroupID: "",
					// consumers:       make([]*consumer, 0),
				},
			},
		},
		conn:              mockConn,
		consumerWaitGroup: &sync.WaitGroup{},
	}

	for i := 0; i < numReaders; i++ {
		mockConsumer := newMockConsumer(t)
		mockConsumers = append(mockConsumers, mockConsumer)
		k.consumerGroupsMap["local.manabie.example-topic"][0].consumers = append(k.consumerGroupsMap["local.manabie.example-topic"][0].consumers, mockConsumer)
	}

	for i := 0; i < numWriters; i++ {
		mockWriter := NewMockWriter(t)
		mockWriters = append(mockWriters, mockWriter)

		key := fmt.Sprintf("local.manabie.example-topic-%d", i)
		k.writerMap[key] = mockWriter
	}

	t.Run("happy case", func(t *testing.T) {
		for i := 0; i < numReaders; i++ {
			consumerCtx, consumerCancel := context.WithCancel(context.Background())
			mockReader := NewMockReader(t)
			mockConsumers[i].On("setRunning", false).Once().Return()
			mockConsumers[i].On("getContextWithCancel").Once().Return(consumerCtx, consumerCancel)
			mockConsumers[i].On("getReader").Once().Return(mockReader)

			mockReaders = append(mockReaders, mockReader)
		}
		for i := 0; i < numReaders; i++ {
			mockReaders[i].On("Close").Once().Return(nil)
		}
		for i := 0; i < numWriters; i++ {
			mockWriters[i].On("Close").Once().Return(nil)
		}
		mockConn.On("Close").Return(nil)

		k.Close()
	})
}
