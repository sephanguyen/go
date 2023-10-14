package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/manabie-com/backend/internal/golibs"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/protocol"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func Test_newConsumer(t *testing.T) {
	options := []kafkaConsumerOption{
		{
			strictCommit: true,
		},
		{
			strictCommit: false,
		},
	}

	t.Run("happy case", func(t *testing.T) {
		for _, opt := range options {
			consumer := newConsumer(opt, nil, false, zap.NewExample())

			switch consumer.(type) {
			case *looseConsumer:
				assert.Equal(t, opt.strictCommit, false)
			case *strictConsumer:
				assert.Equal(t, opt.strictCommit, true)
			}
		}
	})
}

func Test_readMessage(t *testing.T) {
	options := []kafkaConsumerOption{
		{
			strictCommit: true,
		},
		{
			strictCommit: false,
		},
	}

	consumers := make([]consumer, 0)
	mockReaders := make([]*MockReader, 0)
	for _, opt := range options {
		mockReader := NewMockReader(t)
		consumers = append(consumers, newConsumer(opt, mockReader, false, zap.NewExample()))

		mockReaders = append(mockReaders, mockReader)
	}

	t.Run("happy case", func(t *testing.T) {
		for idx, consumer := range consumers {
			returnMsg := kafka.Message{
				Key: []byte(fmt.Sprintf("key-msg-%d", idx)),
			}

			ctx, _ := consumer.getContextWithCancel()
			switch consumer.(type) {
			case *looseConsumer:
				msg := &kafka.Message{}
				mockReaders[idx].On("ReadMessage", ctx).Once().Return(returnMsg, nil)

				err := consumer.readMessage(msg)
				assert.NoError(t, err)
				assert.Equal(t, returnMsg, *msg)
			case *strictConsumer:
				msg := &kafka.Message{}
				mockReaders[idx].On("FetchMessage", ctx).Once().Return(returnMsg, nil)

				err := consumer.readMessage(msg)
				assert.NoError(t, err)
				assert.Equal(t, returnMsg, *msg)
			}
		}
	})

	t.Run("failed case", func(t *testing.T) {
		for idx, consumer := range consumers {
			returnMsg := kafka.Message{}

			ctx, _ := consumer.getContextWithCancel()
			switch consumer.(type) {
			case *looseConsumer:
				msg := &kafka.Message{}
				mockReaders[idx].On("ReadMessage", ctx).Once().Return(returnMsg, dummyError)

				err := consumer.readMessage(msg)
				assert.Error(t, err, fmt.Errorf("kafka: error reader.ReadMessage: %s", dummyError.Error()))
				assert.Equal(t, returnMsg, *msg)
			case *strictConsumer:
				msg := &kafka.Message{}
				mockReaders[idx].On("FetchMessage", ctx).Once().Return(returnMsg, dummyError)

				err := consumer.readMessage(msg)
				assert.Error(t, err, fmt.Errorf("kafka: error reader.FetchMessage: %s", dummyError.Error()))
				assert.Equal(t, returnMsg, *msg)
			}
		}
	})
}

func Test_handleMsg(t *testing.T) {
	options := []kafkaConsumerOption{
		{
			strictCommit: true,
		},
		{
			strictCommit: false,
		},
	}

	consumers := make([]consumer, 0)
	for _, opt := range options {
		consumers = append(consumers, newConsumer(opt, nil, false, zap.NewExample()))
	}

	type TestMessage struct {
		Message string `json:"message"`
	}
	testMessage := TestMessage{
		Message: "message",
	}
	testMessageB, _ := json.Marshal(testMessage)

	var contextFromHandleFunc context.Context
	var testMessageFromHandleFunc TestMessage

	handleFunc := func(ctx context.Context, value []byte) (bool, error) {
		valueRcv := TestMessage{}
		_ = json.Unmarshal(value, &valueRcv)
		testMessageFromHandleFunc = valueRcv
		contextFromHandleFunc = ctx
		return false, nil
	}

	handleFailedFunc := func(ctx context.Context, value []byte) (bool, error) {
		valueRcv := TestMessage{}
		_ = json.Unmarshal(value, &valueRcv)
		testMessageFromHandleFunc = valueRcv
		contextFromHandleFunc = ctx
		return true, dummyError
	}

	userID := "user-id"
	resourcePath := "resource-path"
	kafkaMessage := &kafka.Message{
		Topic:     "topic",
		Partition: 0,
		Offset:    0,
		Key:       []byte("key"),
		Value:     testMessageB,
		Headers: []protocol.Header{
			{
				Key:   kafkaUserIDHeaderName,
				Value: []byte(userID),
			},
			{
				Key:   kafkaResourcePathHeaderName,
				Value: []byte(resourcePath),
			},
			{
				Key:   b3ContextHeader,
				Value: []byte("mock span"),
			},
		},
	}
	t.Run("happy case", func(t *testing.T) {
		for _, consumer := range consumers {
			isContinueNextMsg, err := consumer.handleMessage("unit-test-span", handleFunc, kafkaMessage)
			assert.Equal(t, isContinueNextMsg, true)
			assert.NoError(t, err)
			assert.Equal(t, testMessageFromHandleFunc.Message, testMessage.Message)

			userInfo := golibs.UserInfoFromCtx(contextFromHandleFunc)
			assert.Equal(t, userInfo.UserID, userID)
			assert.Equal(t, userInfo.ResourcePath, resourcePath)
		}
	})

	t.Run("handle failed", func(t *testing.T) {
		for _, consumer := range consumers {
			isContinueNextMsg, err := consumer.handleMessage("unit-test-span", handleFailedFunc, kafkaMessage)

			switch consumer.(type) {
			case *looseConsumer:
				assert.Equal(t, isContinueNextMsg, true)
			case *strictConsumer:
				assert.Equal(t, isContinueNextMsg, false)
			}

			fmt.Printf(err.Error())
			assert.Error(t, err, fmt.Errorf("kafka: handleMsg: %v %w", dummyError, errHandler))
			assert.Equal(t, testMessageFromHandleFunc.Message, testMessage.Message)

			userInfo := golibs.UserInfoFromCtx(contextFromHandleFunc)
			assert.Equal(t, userInfo.UserID, userID)
			assert.Equal(t, userInfo.ResourcePath, resourcePath)
		}
	})
}

func Test_completeMessage(t *testing.T) {
	options := []kafkaConsumerOption{
		{
			strictCommit: true,
		},
		{
			strictCommit: false,
		},
	}

	consumers := make([]consumer, 0)
	mockReaders := make([]*MockReader, 0)
	for _, opt := range options {
		mockReader := NewMockReader(t)
		consumers = append(consumers, newConsumer(opt, mockReader, false, zap.NewExample()))

		mockReaders = append(mockReaders, mockReader)
	}

	t.Run("happy case", func(t *testing.T) {
		for idx, consumer := range consumers {
			msg := &kafka.Message{
				Key: []byte(fmt.Sprintf("key-msg-%d", idx)),
			}

			ctx, _ := consumer.getContextWithCancel()
			switch consumer.(type) {
			case *looseConsumer:
				err := consumer.completeMessage(msg)
				assert.NoError(t, err)
			case *strictConsumer:
				mockReaders[idx].On("CommitMessages", ctx, *msg).Once().Return(nil)

				err := consumer.completeMessage(msg)
				assert.NoError(t, err)
			}
		}
	})

	t.Run("failed case", func(t *testing.T) {
		for idx, consumer := range consumers {
			msg := &kafka.Message{
				Key: []byte(fmt.Sprintf("key-msg-%d", idx)),
			}

			ctx, _ := consumer.getContextWithCancel()
			switch consumer.(type) {
			case *looseConsumer:
				err := consumer.completeMessage(msg)
				assert.NoError(t, err)
			case *strictConsumer:
				mockReaders[idx].On("CommitMessages", ctx, *msg).Once().Return(dummyError)

				err := consumer.completeMessage(msg)
				assert.Error(t, err, fmt.Errorf("kafka: error reader.CommitMessages: %s", dummyError.Error()))
			}
		}
	})
}
