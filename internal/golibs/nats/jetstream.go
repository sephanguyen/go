package nats

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/manabie-com/backend/internal/golibs"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/stringutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"github.com/google/go-cmp/cmp"
	nats "github.com/nats-io/nats.go"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

var (
	jetstreamProcessedMessageSubject = tag.MustNewKey("jetstream_subject_name")
	jetstreamProcessedMessageQueue   = tag.MustNewKey("jetstream_queue_name")
	jetstreamProcessedMessageStatus  = tag.MustNewKey("consumer_handler_status")

	jetstreamProcessedMessagesCounter = stats.Int64("jetstream/processed_messages", "Jetstream processed messages", stats.UnitDimensionless)
	jetStreamProcessedMessagesLatency = stats.Float64("jetstream/processed_messages_latency", "Jetstream processed messages latency", stats.UnitMilliseconds)

	JetstreamProcessedMessagesView = &view.View{
		Name:        "jetstream/processed_messages",
		Description: "Count of processed messages by subject and queue name",
		TagKeys: []tag.Key{
			jetstreamProcessedMessageSubject,
			jetstreamProcessedMessageQueue,
			jetstreamProcessedMessageStatus,
		},
		Measure:     jetstreamProcessedMessagesCounter,
		Aggregation: view.Count(),
	}

	JetstreamProcessedMessagesLatencyView = &view.View{
		Name:        "jetstream/processed_messages_latency",
		Description: "Distribution of processed messages latency, by subject and queue name",
		TagKeys:     []tag.Key{jetstreamProcessedMessageSubject, jetstreamProcessedMessageQueue},
		Measure:     jetStreamProcessedMessagesLatency,
		Aggregation: view.Distribution(5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000),
	}

	errMsgAck            = errors.New("jetstream: cannot ack message")
	errProtoUnmarshalMsg = errors.New("jetstream: cannot unmarshal message into protobuf message")
	errHandler           = errors.New("jetstream: consumer handler fail to process message")
)

type MsgsHandler func(msgs []*nats.Msg) error

var ErrConsumerAlreadyExists = errors.New("consumer is already exists")

type JetStreamManagement interface {
	ConnectToJS()
	GetJS() nats.JetStreamContext
	UpsertStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) error
	UpsertConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) error
	PublishAsyncContext(ctx context.Context, subject string, data []byte, opts ...nats.PubOpt) (string, error)
	PublishContext(ctx context.Context, subject string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error)
	TracedPublishAsync(ctx context.Context, spanName, subject string, data []byte, opts ...nats.PubOpt) (string, error)
	TracedPublish(ctx context.Context, spanName, subject string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error)
	Subscribe(subject string, option Option, cb MsgHandler) (*Subscription, error)
	QueueSubscribe(subject, queue string, option Option, cb MsgHandler) (*Subscription, error)
	PullSubscribe(subject, durable string, cb MsgsHandler, option Option) error
	Close()

	RegisterReconnectHandler(h nats.ConnHandler)
	RegisterDisconnectErrHandler(h nats.ConnErrHandler)
}

type jetStreamManagementImpl struct {
	sync.RWMutex
	url           string
	user          string
	password      string
	conn          *nats.Conn
	js            nats.JetStreamContext
	logger        *zap.Logger
	opts          []nats.Option
	startTime     time.Time
	maxReconnect  int
	reconnectWait time.Duration
	isLocal       bool
	subs          []*nats.Subscription

	disconnectErrHandlers []nats.ConnErrHandler
	reconnectHandlers     []nats.ConnHandler
}

type Option struct {
	JetStreamOptions []JSSubOption
	SpanName         string
	SkipMsgOlderThan *time.Duration // when set, skip messages whose age is older than this value
	PullOpt          PullSubscribeOption
}

type PullSubscribeOption struct {
	BatchSize int
	FetchSize int
}

type PayloadMsg struct {
	Data         []byte
	ResourcePath string
	UserID       string
}

type Subscription struct {
	JetStreamSub *nats.Subscription
}

type JSSubOption interface {
	configureSubscribeOption(opts *jsSubOptions) error
}

type jsSubOptFn func(opts *jsSubOptions) error

func (opt jsSubOptFn) configureSubscribeOption(opts *jsSubOptions) error {
	return opt(opts)
}

type jsSubOptions struct {
	consumerConfig *nats.ConsumerConfig
	streamName     string
	subOption      []nats.SubOpt
}

func Bind(streamName string, consumerName string) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.streamName = streamName
		opts.consumerConfig.Durable = consumerName
		opts.subOption = append(opts.subOption, nats.Bind(streamName, consumerName))
		return nil
	})
}

func ManualAck() JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.subOption = append(opts.subOption, nats.ManualAck())
		return nil
	})
}

func MaxDeliver(n int) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.MaxDeliver = n
		opts.subOption = append(opts.subOption, nats.MaxDeliver(n))
		return nil
	})
}

func FitlerSubject(n string) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.FilterSubject = n
		return nil
	})
}

func AckWait(n time.Duration) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.AckWait = n
		opts.subOption = append(opts.subOption, nats.AckWait(n))
		return nil
	})
}

// AckAll when acking a sequence number, this implicitly acks all sequences
// below this one as well.
func AckAll() JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.AckPolicy = nats.AckAllPolicy
		opts.subOption = append(opts.subOption, nats.AckAll())
		return nil
	})
}

// AckExplicit requires ack or nack for all messages.
func AckExplicit() JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.AckPolicy = nats.AckExplicitPolicy
		opts.subOption = append(opts.subOption, nats.AckExplicit())
		return nil
	})
}

func StartTime(startTime time.Time) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.DeliverPolicy = nats.DeliverByStartTimePolicy
		opts.consumerConfig.OptStartTime = &startTime
		opts.subOption = append(opts.subOption, nats.StartTime(startTime))
		return nil
	})
}

func DeliverNew() JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.DeliverPolicy = nats.DeliverNewPolicy
		return nil
	})
}

func DeliverLast() JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.DeliverPolicy = nats.DeliverLastPolicy
		return nil
	})
}

func StartSequence(seq uint64) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.DeliverPolicy = nats.DeliverByStartSequencePolicy
		opts.consumerConfig.OptStartSeq = seq
		opts.subOption = append(opts.subOption, nats.StartSequence(seq))
		return nil
	})
}

func checkDurName(dur string) error {
	if strings.Contains(dur, ".") {
		return nats.ErrInvalidDurableName
	}
	return nil
}

func Durable(consumer string) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		if opts.consumerConfig.Durable != "" {
			return fmt.Errorf("nats: option Durable set more than once")
		}
		if err := checkDurName(consumer); err != nil {
			return err
		}

		opts.consumerConfig.Durable = consumer
		opts.subOption = append(opts.subOption, nats.Durable(consumer))
		return nil
	})
}

func DeliverSubject(subject string) JSSubOption {
	return jsSubOptFn(func(opts *jsSubOptions) error {
		opts.consumerConfig.DeliverSubject = subject
		opts.subOption = append(opts.subOption, nats.DeliverSubject(subject))
		return nil
	})
}

// bool: whether this message be should be retried
type MsgHandler func(ctx context.Context, data []byte) (bool, error)

func NewJetStreamManagement(url, user, password string, maxReconnect int, reconnectWait time.Duration, isLocal bool, zapLogger *zap.Logger) (JetStreamManagement, error) {
	if url == "" {
		return nil, errors.New("missing url")
	}

	if user == "" {
		return nil, errors.New("missing user")
	}

	if password == "" {
		return nil, errors.New("missing password")
	}

	if zapLogger == nil {
		return nil, errors.New("missing logger")
	}

	n := &jetStreamManagementImpl{
		url:           url,
		user:          user,
		password:      password,
		logger:        zapLogger,
		maxReconnect:  maxReconnect,
		reconnectWait: reconnectWait,
		isLocal:       isLocal,
	}

	return n, nil
}

func (n *jetStreamManagementImpl) RegisterReconnectHandler(h nats.ConnHandler) {
	n.reconnectHandlers = append(n.reconnectHandlers, h)
}

func (n *jetStreamManagementImpl) RegisterDisconnectErrHandler(h nats.ConnErrHandler) {
	n.disconnectErrHandlers = append(n.disconnectErrHandlers, h)
}

func (n *jetStreamManagementImpl) disconnectErrHandler(c *nats.Conn, err error) {
	n.logger.Warn("Disconnected from nats server", zap.Error(err), zap.String("URL", c.ConnectedUrl()))
	for _, handler := range n.disconnectErrHandlers {
		handler(c, err)
	}
}

func (n *jetStreamManagementImpl) reconnectHandler(c *nats.Conn) {
	n.logger.Info("Reconnected to nats server", zap.String("URL", c.ConnectedUrl()))
	for _, handler := range n.reconnectHandlers {
		handler(c)
	}
}

func (n *jetStreamManagementImpl) ConnectToJS() {
	conn, err := nats.Connect(
		n.url,
		nats.UserInfo(n.user, n.password),
		nats.ReconnectWait(n.reconnectWait),
		nats.MaxReconnects(n.maxReconnect),
		nats.DisconnectErrHandler(n.disconnectErrHandler),
		nats.ReconnectHandler(n.reconnectHandler),
	)
	if err != nil {
		n.logger.Fatal(err.Error())
	}

	js, err := conn.JetStream()
	if err != nil {
		n.logger.Fatal(err.Error())
	}

	n.Lock()
	defer n.Unlock()

	n.conn = conn
	n.js = js
	n.logger.Info("nats jetstream connected")
}

func (n *jetStreamManagementImpl) GetJS() nats.JetStreamContext {
	n.RLock()
	defer n.RUnlock()
	js := n.js

	return js
}

func (n *jetStreamManagementImpl) UpsertStream(cfg *nats.StreamConfig, opts ...nats.JSOpt) error {
	// we just need 1 replicas in local, it help save our resource
	if n.isLocal {
		cfg.Replicas = 1
	}

	info, err := n.js.StreamInfo(cfg.Name)

	// case 1: we already have stream and need to update stream
	if err == nil {
		if info.Config.Retention != cfg.Retention {
			err := n.js.DeleteStream(cfg.Name)
			if err != nil {
				return fmt.Errorf("failed to delete stream: %s, err: %s", cfg.Name, err)
			}

			_, err = n.js.AddStream(cfg, opts...)
			if err != nil {
				return fmt.Errorf("failed to add new stream: %s", err)
			}
		} else if !stringutil.SliceEqual(info.Config.Subjects, cfg.Subjects) ||
			info.Config.Replicas != cfg.Replicas ||
			info.Config.MaxAge != cfg.MaxAge ||
			info.Config.MaxBytes != cfg.MaxBytes {
			_, err = n.js.UpdateStream(cfg, opts...)
			if err != nil {
				return fmt.Errorf("failed to update stream: %s", err)
			}
		}
		return nil
	}

	// case 2: Stream is not exists before => create a new stream
	if err != nil && err == nats.ErrStreamNotFound {
		_, err = n.js.AddStream(cfg, opts...)
		if err != nil {
			return fmt.Errorf("failed to add new stream: %s", err)
		}
		return nil
	}

	// case 3: Another error when get stream info
	return fmt.Errorf("error getting stream info: %s", err)
}

func (n *jetStreamManagementImpl) UpsertConsumer(stream string, cfg *nats.ConsumerConfig, opts ...nats.JSOpt) error {
	info, err := n.js.ConsumerInfo(stream, cfg.Durable)
	if err != nil {
		// case 1: consumer is not exits before => create a new consumer
		if err == nats.ErrConsumerNotFound {
			_, err = n.js.AddConsumer(stream, cfg, opts...)
			return err
		}

		// case 2: another error when get consumer info
		return err
	}

	// case 3: if consumer is already exists && have different config with current config => Delete and Recreate consumer
	var needDelete bool
	if info.Config.MaxDeliver != cfg.MaxDeliver ||
		info.Config.AckWait != cfg.AckWait ||
		info.Config.DeliverGroup != cfg.DeliverGroup ||
		info.Config.OptStartTime != cfg.OptStartTime ||
		info.Config.AckPolicy != cfg.AckPolicy ||
		info.Config.DeliverSubject != cfg.DeliverSubject ||
		info.Config.FilterSubject != cfg.FilterSubject ||
		info.Config.DeliverPolicy != cfg.DeliverPolicy ||
		info.Config.OptStartSeq != cfg.OptStartSeq {
		needDelete = true
		n.logger.Sugar().Warnf("deleting consumer %s because: %s", cfg.Durable, cmp.Diff(info.Config, *cfg))
	}
	if needDelete {
		err = n.js.DeleteConsumer(stream, cfg.Durable)
		if err != nil {
			return fmt.Errorf("err when delete consumer: %v", err)
		}
		_, err = n.js.AddConsumer(stream, cfg, opts...)
		return err
	}
	return nil
}

func (n *jetStreamManagementImpl) PublishAsyncContext(ctx context.Context, subject string, data []byte, opts ...nats.PubOpt) (string, error) {
	userInfo := golibs.UserInfoFromCtx(ctx)
	p := npb.DataInMessage{
		Payload:      data,
		ResourcePath: userInfo.ResourcePath,
		UserId:       userInfo.UserID,
	}
	payload, err := proto.Marshal(&p)
	if err != nil {
		n.logger.Error("marshal resource_path from context fail", zap.Error(err))
		return "", err
	}

	opts = append(opts, nats.MsgId(idutil.ULIDNow()))

	pubAck, err := n.js.PublishAsync(subject, payload, opts...)
	if err != nil {
		n.logger.Error("publish async message is error", zap.Error(err), zap.String("subject", subject))
		return "", err
	}
	return pubAck.Msg().Header.Get("Nats-Msg-Id"), err
}

func (n *jetStreamManagementImpl) TracedPublishAsync(ctx context.Context, spanName, subject string, data []byte, opts ...nats.PubOpt) (string, error) {
	ctx, span := interceptors.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))

	userInfo := golibs.UserInfoFromCtx(ctx)
	p := npb.DataInMessage{
		Payload:      data,
		ResourcePath: userInfo.ResourcePath,
		UserId:       userInfo.UserID,
	}
	traceInfo := TraceInfoFromContext(ctx)
	p.TraceInfo = traceInfo

	payload, err := proto.Marshal(&p)
	if err != nil {
		n.logger.Error("marshal DataInMessage from []byte fail", zap.Error(err))
		span.RecordError(err)
		span.End()
		return "", err
	}

	opts = append(opts, nats.MsgId(idutil.ULIDNow()))

	pubAck, err := n.js.PublishAsync(subject, payload, opts...)
	if err != nil {
		n.logger.Error("traced publish async message is error", zap.Error(err))
		span.RecordError(err)
		span.End()
		return "", err
	}
	go func() {
		select {
		case <-pubAck.Ok():
		case err := <-pubAck.Err():
			span.RecordError(err)
		}
		span.End()
	}()
	return pubAck.Msg().Header.Get("Nats-Msg-Id"), err
}

func (n *jetStreamManagementImpl) PublishContext(ctx context.Context, subject string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	userInfo := golibs.UserInfoFromCtx(ctx)

	p := npb.DataInMessage{
		Payload:      data,
		ResourcePath: userInfo.ResourcePath,
		UserId:       userInfo.UserID,
	}
	payload, err := proto.Marshal(&p)
	if err != nil {
		n.logger.Error("marshal resource_path from context fail", zap.Error(err))
		return nil, err
	}

	opts = append(opts, nats.MsgId(idutil.ULIDNow()))

	pubAck, err := n.js.Publish(subject, payload, opts...)
	if err != nil {
		n.logger.Error("publish message is error", zap.Error(err))
	}
	return pubAck, err
}

func (n *jetStreamManagementImpl) TracedPublish(ctx context.Context, spanName, subject string, data []byte, opts ...nats.PubOpt) (*nats.PubAck, error) {
	ctx, span := interceptors.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindProducer))
	defer span.End()
	userInfo := golibs.UserInfoFromCtx(ctx)
	p := npb.DataInMessage{
		Payload:      data,
		ResourcePath: userInfo.ResourcePath,
		UserId:       userInfo.UserID,
	}
	traceInfo := TraceInfoFromContext(ctx)
	p.TraceInfo = traceInfo
	payload, err := proto.Marshal(&p)
	if err != nil {
		n.logger.Error("marshal DataInMessage from []byte fail", zap.Error(err))
		return nil, err
	}

	opts = append(opts, nats.MsgId(idutil.ULIDNow()))

	pubAck, err := n.js.Publish(subject, payload, opts...)
	if err != nil {
		n.logger.Error("traced publish async message is error", zap.Error(err))
	}
	return pubAck, err
}

func (n *jetStreamManagementImpl) Subscribe(subject string, option Option, cb MsgHandler) (*Subscription, error) {
	option.JetStreamOptions = append([]JSSubOption{AckExplicit()}, option.JetStreamOptions...)
	if n.isLocal {
		option.JetStreamOptions = append(option.JetStreamOptions, AckWait(2*time.Second))
	}
	consumerConfig := &nats.ConsumerConfig{}
	o := jsSubOptions{consumerConfig: consumerConfig}
	o.consumerConfig.FilterSubject = subject

	for _, v := range option.JetStreamOptions {
		if err := v.configureSubscribeOption(&o); err != nil {
			return nil, err
		}
	}

	s, err := n.js.Subscribe(subject, n.logicInJS(subject, "", option.SpanName, option.SkipMsgOlderThan, cb), o.subOption...)
	if err != nil {
		return nil, fmt.Errorf("subscribe: %w", err)
	}

	sub := Subscription{JetStreamSub: s}

	n.subs = append(n.subs, s)

	return &sub, err
}

func (n *jetStreamManagementImpl) QueueSubscribe(subject, queue string, option Option, cb MsgHandler) (*Subscription, error) {
	option.JetStreamOptions = append([]JSSubOption{AckExplicit()}, option.JetStreamOptions...)
	if n.isLocal {
		option.JetStreamOptions = append(option.JetStreamOptions, AckWait(4*time.Second))
	}
	consumerConfig := &nats.ConsumerConfig{}
	o := jsSubOptions{consumerConfig: consumerConfig}
	o.consumerConfig.FilterSubject = subject

	for _, v := range option.JetStreamOptions {
		if err := v.configureSubscribeOption(&o); err != nil {
			return nil, err
		}
	}

	o.consumerConfig.DeliverGroup = queue

	err := try.Do(func(attempt int) (retry bool, err error) {
		err = n.UpsertConsumer(o.streamName, consumerConfig)
		if err == nil {
			return false, nil
		}
		time.Sleep(1 * time.Second)
		return attempt < 5, err
	})

	if err != nil && err != ErrConsumerAlreadyExists {
		return nil, err
	}

	s, err := n.js.QueueSubscribe(subject, queue, n.logicInJS(subject, queue, option.SpanName, option.SkipMsgOlderThan, cb), o.subOption...)
	if err != nil {
		return nil, fmt.Errorf("QueueSubscribe: %w", err)
	}

	sub := Subscription{JetStreamSub: s}

	n.subs = append(n.subs, s)

	return &sub, err
}

func processMsgCounter(ctx context.Context, subject, queue string, count int64, startTime time.Time, err error) {
	if err == nil {
		elapsedTime := time.Since(startTime)
		latencyMs := float64(elapsedTime) / float64(time.Millisecond)

		_ = stats.RecordWithTags(
			ctx,
			[]tag.Mutator{
				tag.Upsert(jetstreamProcessedMessageSubject, subject),
				tag.Upsert(jetstreamProcessedMessageQueue, queue),
			},
			jetStreamProcessedMessagesLatency.M(latencyMs),
		)
	}

	var processedMsgStatus string
	switch {
	case errors.Is(err, errMsgAck):
		processedMsgStatus = "ACK_ERROR"
	case errors.Is(err, errProtoUnmarshalMsg):
		processedMsgStatus = "PROTO_UNMARSHAL_ERROR"
	case errors.Is(err, errHandler):
		processedMsgStatus = "HANDLER_ERROR"
	default:
		processedMsgStatus = "OK"
	}

	_ = stats.RecordWithTags(
		ctx,
		[]tag.Mutator{
			tag.Upsert(jetstreamProcessedMessageSubject, subject),
			tag.Upsert(jetstreamProcessedMessageQueue, queue),
			tag.Upsert(jetstreamProcessedMessageStatus, processedMsgStatus),
		},
		jetstreamProcessedMessagesCounter.M(count),
	)
}

func (n *jetStreamManagementImpl) handleMsg(subject, queue, spanName string, skipMsgOlderThan *time.Duration, cb MsgHandler, msg *nats.Msg, log *zap.Logger) (logger *zap.Logger, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancel()

	meta, err := msg.Metadata()
	var sequence uint64
	if err == nil {
		sequence = meta.Sequence.Stream
	}

	msgID := msg.Header.Get("Nats-Msg-Id")
	logger = log.With(
		zap.String("message_id", msgID),
		zap.Uint64("sequence", sequence),
	)

	startTime := time.Now()
	defer processMsgCounter(ctx, subject, queue, 1, startTime, err)

	if skipMsgOlderThan != nil {
		if math.Abs(time.Since(meta.Timestamp).Seconds()) > skipMsgOlderThan.Seconds() {
			logger.Warn("DENY_CONSUME_OLD_MESSAGE", zap.Time("time_receive_message", meta.Timestamp))
			return
		}
	}

	var dataInMsg npb.DataInMessage
	if err = proto.Unmarshal(msg.Data, &dataInMsg); err != nil {
		if ackErr := msg.Ack(); ackErr != nil {
			logger.Error("msg.Ack", zap.Error(ackErr))
		}

		err = fmt.Errorf("proto.Unmarshal: %v %w", err, errProtoUnmarshalMsg)
		return
	}

	var span interceptors.TimedSpan
	var hasSpan bool
	if dataInMsg.TraceInfo != nil {
		hasSpan = true
		// assign trace info inside context
		ctx = ContextWithTraceInfo(ctx, dataInMsg.TraceInfo)
		ctx, span = interceptors.StartSpan(ctx, spanName, trace.WithSpanKind(trace.SpanKindConsumer))
		defer span.End()
	}

	// assign resource_path inside context
	// ctx = ContextWithResourcePath(ctx, dataInMsg.ResourcePath)
	claim := &interceptors.CustomClaims{
		Manabie: &interceptors.ManabieClaims{
			ResourcePath: dataInMsg.ResourcePath,
			UserID:       dataInMsg.UserId,
		},
	}
	ctx = interceptors.ContextWithJWTClaims(ctx, claim)

	retry, err := cb(ctx, dataInMsg.Payload)
	if err != nil {
		if hasSpan {
			span.RecordError(err)
		}

		if !retry {
			if ackErr := msg.Ack(); ackErr != nil {
				logger.Error("msg.Ack", zap.Error(ackErr))
			}
		}

		err = fmt.Errorf("cb: %v %w", err, errHandler)
		return
	}

	if err = msg.Ack(); err != nil {
		err = fmt.Errorf("msg.Ack: %v %w", err, errMsgAck)
	}

	return
}

func (n *jetStreamManagementImpl) logicInJS(subject, queue, spanName string, skipMsgOlderThan *time.Duration, cb MsgHandler) func(msg *nats.Msg) {
	return func(msg *nats.Msg) {
		logger := n.logger.With(
			zap.String("subject", subject),
			zap.String("queue", queue),
		)
		if logger, err := n.handleMsg(subject, queue, spanName, skipMsgOlderThan, cb, msg, logger); err != nil {
			logger.Error("n.handleMsg", zap.Error(err))
		}
	}
}

func (n *jetStreamManagementImpl) PullSubscribe(subject, durable string, cb MsgsHandler, option Option) error {
	option.JetStreamOptions = append([]JSSubOption{AckExplicit()}, option.JetStreamOptions...)
	if n.isLocal {
		option.JetStreamOptions = append(option.JetStreamOptions, AckWait(4*time.Second))
	}

	consumerConfig := &nats.ConsumerConfig{}
	o := jsSubOptions{consumerConfig: consumerConfig}
	o.consumerConfig.FilterSubject = subject

	for _, v := range option.JetStreamOptions {
		if err := v.configureSubscribeOption(&o); err != nil {
			return err
		}
	}

	err := try.Do(func(attempt int) (retry bool, err error) {
		err = n.UpsertConsumer(o.streamName, consumerConfig)
		if err == nil {
			return false, nil
		}
		time.Sleep(1 * time.Second)
		return attempt < 5, err
	})
	if err != nil && err != ErrConsumerAlreadyExists {
		return err
	}

	sub, err := n.js.PullSubscribe(subject, durable, o.subOption...)
	if err != nil {
		return err
	}

	go func() {
		for {
			msgs, _ := sub.Fetch(option.PullOpt.FetchSize)
			if len(msgs) == 0 {
				time.Sleep(time.Millisecond * 500)
				continue
			}

			err := cb(msgs)
			if err == nil {
				for _, msg := range msgs {
					if err := msg.Ack(); err != nil {
						n.logger.Error("Ack message is failed", zap.Error(err))
						processMsgCounter(context.Background(), subject, "", 1, time.Now(), errMsgAck)
					}
				}
			} else {
				n.logger.Error("Messages are process failed", zap.Error(err))
				processMsgCounter(context.Background(), subject, "", int64(len(msgs)), time.Now(), errHandler)
			}
		}
	}()

	return nil
}

func (n *jetStreamManagementImpl) Close() {
	n.RLock()
	defer n.RUnlock()
	for _, s := range n.subs {
		info, err := s.ConsumerInfo()
		if err != nil {
			return
		}
		n.logger.Info("Current consumer", zap.String("consumer", info.Name))
		err = s.Drain()
		if err != nil {
			n.logger.Error("Drain consumer have error", zap.Error(err))
		}
	}
	n.conn.Close()
}

func HandlePushMsgFail(ctx context.Context, err error) error {
	ctxzap.Extract(ctx).Error("push msg fail", zap.Error(err))
	return err
}
