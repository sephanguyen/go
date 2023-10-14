package kafka

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/try"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type consumer interface {
	getRunning() bool
	setRunning(isRunning bool)
	getContextWithCancel() (context.Context, context.CancelFunc)
	getReader() Reader

	readMessage(msg *kafka.Message) error
	handleMessage(spanName string, handleMsg MsgHandler, msg *kafka.Message) (bool, error)
	completeMessage(msg *kafka.Message) error
}

func newConsumer(
	option kafkaConsumerOption,
	reader Reader,
	isRunning bool,
	logger *zap.Logger,
) consumer {
	consumerCtx, consumerCancelFunc := context.WithCancel(context.Background())
	if option.strictCommit {
		return &strictConsumer{
			commonConsumer: commonConsumer{
				option:             option,
				isRunning:          isRunning,
				reader:             reader,
				consumerContext:    consumerCtx,
				consumerCancelFunc: consumerCancelFunc,
				logger:             logger,
			},
		}
	}

	return &looseConsumer{
		commonConsumer: commonConsumer{
			option:             option,
			isRunning:          isRunning,
			reader:             reader,
			consumerContext:    consumerCtx,
			consumerCancelFunc: consumerCancelFunc,
			logger:             logger,
		},
	}
}

type commonConsumer struct {
	logger             *zap.Logger
	option             kafkaConsumerOption
	isRunning          bool
	reader             Reader
	consumerContext    context.Context
	consumerCancelFunc context.CancelFunc
}

func (c *commonConsumer) getRunning() bool { return c.isRunning }

func (c *commonConsumer) setRunning(isRunning bool) { c.isRunning = isRunning }

func (c *commonConsumer) getContextWithCancel() (context.Context, context.CancelFunc) {
	return c.consumerContext, c.consumerCancelFunc
}

func (c *commonConsumer) getReader() Reader {
	return c.reader
}

func (c *commonConsumer) handleMessage(spanName string, handleMsg MsgHandler, msg *kafka.Message) (err error) {
	err = try.DoWithCtx(c.consumerContext, func(_ context.Context, attempt int) (bool, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
		defer cancel()
		traceCarrier, claim := TraceCarrierAndClaimInfoFromMessageHeaders(msg.Headers)
		logger := c.logger.With(
			zap.String("key", string(msg.Key)),
			zap.String("topic", msg.Topic),
			zap.Int("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
		)

		// Assign resource_path and user_id inside context.
		ctx = interceptors.ContextWithJWTClaims(ctx, claim)

		var span interceptors.TimedSpan
		var hasSpan bool
		if traceCarrier.GetAllValues() != nil {
			hasSpan = true
			// Assign trace info inside context.
			ctx = ContextWithTraceCarrier(ctx, traceCarrier)
			ctx, span = interceptors.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
			defer span.End()
			span.SetAttributes(attribute.KeyValue{
				Key:   "message_payload",
				Value: attribute.StringValue(string(msg.Value)),
			})
		}

		payload := msg.Value
		isRetryLogic, err := handleMsg(ctx, payload)
		if err != nil {
			if hasSpan {
				span.RecordError(err)
			}

			err = fmt.Errorf("kafka handleMsg: [%w], raw: [%v]", errHandler, err)
			logger.Error("kafka: handleMsg error", zap.Error(err))
			logger.Error(fmt.Sprintf("kafka: payload of failed handled message: [%v]", string(msg.Value)), zap.Error(err))
		}

		isCanRetryLogic := attempt < c.option.retryLogicAttempts
		if err != nil && isCanRetryLogic && isRetryLogic {
			c.logger.Warn(fmt.Sprintf("kafka: handling logic occurred error, retrying after %f second(s)...", c.option.waitingTimeToRetryLogic.Seconds()), zap.Error(err))
			time.Sleep(c.option.waitingTimeToRetryLogic)

			return isRetryLogic, err
		}

		return false, err
	})

	return err
}

type looseConsumer struct {
	commonConsumer
}

func (c *looseConsumer) readMessage(msg *kafka.Message) (err error) {
	err = try.DoWithCtx(c.consumerContext, func(_ context.Context, attempt int) (bool, error) {
		var err error
		*msg, err = c.reader.ReadMessage(c.consumerContext)
		if err == nil {
			return false, nil
		}

		isCanRetryConnect := attempt < c.option.reconnectAttempts
		if err != nil && isCanRetryConnect && !errors.Is(err, io.EOF) {
			c.logger.Warn(fmt.Sprintf("kafka: reading message with network error, retrying after %f second(s)...", c.option.waitingTimeToReconnect.Seconds()), zap.Error(err))
			time.Sleep(c.option.waitingTimeToReconnect)

			return true, err
		}

		return false, err
	})
	if err != nil {
		return fmt.Errorf("kafka: error reader.ReadMessage: %s", err.Error())
	}

	return
}

func (c *looseConsumer) handleMessage(spanName string, handleMsg MsgHandler, msg *kafka.Message) (isContinueToNextMsg bool, err error) {
	err = c.commonConsumer.handleMessage(spanName, handleMsg, msg)
	isContinueToNextMsg = true

	return
}

func (c *looseConsumer) completeMessage(_ *kafka.Message) error {
	return nil
}

type strictConsumer struct {
	commonConsumer
}

func (c *strictConsumer) readMessage(msg *kafka.Message) (err error) {
	err = try.DoWithCtx(c.consumerContext, func(_ context.Context, attempt int) (bool, error) {
		var err error

		// Remember to commit message after processed if use FetchMessage.
		*msg, err = c.reader.FetchMessage(c.consumerContext)
		if err == nil {
			return false, nil
		}

		isCanRetryConnect := attempt < c.option.reconnectAttempts
		if err != nil && isCanRetryConnect && !errors.Is(err, io.EOF) {
			c.logger.Warn(fmt.Sprintf("kafka: fetching message with network error, retrying after %f second(s)...", c.option.waitingTimeToReconnect.Seconds()), zap.Error(err))
			time.Sleep(c.option.waitingTimeToReconnect)

			return true, err
		}

		return false, err
	})
	if err != nil {
		return fmt.Errorf("kafka: error reader.FetchMessage: %s", err.Error())
	}

	return
}

func (c *strictConsumer) handleMessage(spanName string, handleMsg MsgHandler, msg *kafka.Message) (isContinueToNextMsg bool, err error) {
	err = c.commonConsumer.handleMessage(spanName, handleMsg, msg)
	isContinueToNextMsg = true
	if err != nil {
		isContinueToNextMsg = false
	}

	return
}

func (c *strictConsumer) completeMessage(msg *kafka.Message) error {
	err := try.DoWithCtx(c.consumerContext, func(_ context.Context, attempt int) (bool, error) {
		err := c.reader.CommitMessages(c.consumerContext, *msg)
		if err == nil {
			return false, nil
		}

		isCanRetryConnect := attempt < c.option.reconnectAttempts
		if err != nil && isCanRetryConnect && !errors.Is(err, io.EOF) {
			c.logger.Warn(fmt.Sprintf("kafka: commit messages with network error, retrying after %f second(s)...", c.option.waitingTimeToReconnect.Seconds()), zap.Error(err))
			time.Sleep(c.option.waitingTimeToReconnect)

			return true, err
		}
		return false, err
	})
	if err != nil {
		// Can't commit message after retry -> return error
		return fmt.Errorf("kafka: error reader.CommitMessages: %s", err.Error())
	}

	return nil
}
