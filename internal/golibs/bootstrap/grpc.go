package bootstrap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"

	"github.com/golang/protobuf/proto" //nolint:staticcheck
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	errNotBaseService = errors.New("not a base service (does not implement bootstrap.BaseServicer interface)")
	defaultMaxTimeout = time.Second * 5
)

// mockery --dir internal/golibs/bootstrap --name AllServicer --structname MockAllService --filename internal/golibs/bootstrap/mocks_test.go --print
// To gen mockery only
type AllServicer[T any] interface {
	BaseServicer[T]
	GRPCServicer[T]
	HTTPServicer[T]
	NatsServicer[T]
	KafkaServicer[T]
	MonitorServicer[T]
}

// Every service must implement this interface
type BaseServicer[T any] interface {
	// Service name and cmdline name
	ServerName() string

	// Setup all customized dependencies of your service
	// not including those managed by platform, aka whatever
	// available in Resource (Agora,Firebase,S3),
	// maybe create some complex service obj that may share between grpc/nats/http
	InitDependencies(T, *Resources) error

	// Shutdown your own dependencies, not including those managed
	// by platform
	GracefulShutdown(context.Context)
}

// GRPCServicer must be implemented to bootstrap a GRPC server.
type GRPCServicer[T any] interface {
	// Register your implementation to the provided grpc server
	// You don't need to start any net Listener
	SetupGRPC(context.Context, *grpc.Server, T, *Resources) error
	WithUnaryServerInterceptors(T, *Resources) []grpc.UnaryServerInterceptor
	WithStreamServerInterceptors(T, *Resources) []grpc.StreamServerInterceptor
	WithServerOptions() []grpc.ServerOption
}

func (b *bootstrapper[T]) setupGRPCService(ctx context.Context, servicer GRPCServicer[T], c *T, rsc *Resources) (*grpc.Server, error) {
	b.unaryInterceptors = append(servicer.WithUnaryServerInterceptors(*c, rsc), b.unaryInterceptors...)

	// add recovery interceptor
	cc, err := extract[configs.CommonConfig](c, "Common")
	if err != nil {
		return nil, fmt.Errorf("setupGRPCService: failed to extract configs.CommonConfig, error: %w", err)
	}

	grpcStream := servicer.WithStreamServerInterceptors(*c, rsc)
	if cc.Environment != "local" {
		b.unaryInterceptors = interceptors.WithUnaryServerRecovery(b.unaryInterceptors)
		grpcStream = interceptors.WithStreamServerRecovery(grpcStream)
	}

	// add timeout interceptor.
	// v1 or the new v2, depending on the flag in config
	if cc.GRPC.HandlerTimeoutV2Enabled {
		timeoutInterceptor := interceptors.TimeoutUnaryServerInterceptorV2(rsc.Logger(), cc.GRPC.HandlerTimeoutV2)
		b.unaryInterceptors = append(b.unaryInterceptors, timeoutInterceptor)
	} else if cc.GRPC.HandlerTimeout.Seconds() >= 0 { // if `HandlerTimeout` = -1s, the service will bypass timeout policy.
		if cc.GRPC.HandlerTimeout.Seconds() == 0 {
			cc.GRPC.HandlerTimeout = defaultMaxTimeout
		}

		b.unaryInterceptors = append(b.unaryInterceptors, interceptors.TimeoutUnaryServerInterceptor(cc.GRPC.HandlerTimeout))
	}

	opts := []grpc.ServerOption{
		grpc_middleware.WithUnaryServerChain(b.unaryInterceptors...),
		grpc_middleware.WithStreamServerChain(grpcStream...),
	}

	opts = append(opts, servicer.WithServerOptions()...)
	grpcServer := grpc.NewServer(opts...)
	err = servicer.SetupGRPC(ctx, grpcServer, *c, rsc)
	if err != nil {
		return nil, fmt.Errorf("setupGRPCService: %w", err)
	}

	return grpcServer, nil
}

func (b *bootstrapper[T]) gracefulShutdown(ctx context.Context, s BaseServicer[T], rsc *Resources) {
	s.GracefulShutdown(ctx)
}

// DefaultUnaryServerInterceptor returns the default unary server interceptors for a GPRC service.
func DefaultUnaryServerInterceptor(
	rsc *Resources,
) []grpc.UnaryServerInterceptor {
	clonedLogger := rsc.Logger().WithOptions(zap.AddStacktrace(zap.FatalLevel))
	return []grpc.UnaryServerInterceptor{
		grpc_ctxtags.UnaryServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.UnaryServerInterceptor(clonedLogger, grpc_zap.WithLevels(codeToLevel)),
		PayloadUnaryServerInterceptor(clonedLogger, func(_ context.Context, _ string, _ interface{}) bool { return true }),
	}
}

// DefaultStreamServerInterceptor returns the default stream server interceptors for a GPRC service.
func DefaultStreamServerInterceptor(
	rsc *Resources,
) []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpc_ctxtags.StreamServerInterceptor(grpc_ctxtags.WithFieldExtractor(grpc_ctxtags.CodeGenRequestFieldExtractor)),
		grpc_zap.StreamServerInterceptor(
			rsc.Logger().WithOptions(zap.AddStacktrace(zap.FatalLevel)),
			grpc_zap.WithLevels(codeToLevel),
		),
	}
}

// PayloadUnaryServerInterceptor returns a new unary server interceptors that logs the payloads of requests.
// It is a customized version of grpc_zap.PayloadUnaryServerInterceptor.
//
// This *only* works when placed *after* the `grpc_zap.UnaryServerInterceptor`. However, the logging can be done to a
// separate instance of the logger.
func PayloadUnaryServerInterceptor(logger *zap.Logger, decider grpc_logging.ServerPayloadLoggingDecider) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if !decider(ctx, info.FullMethod, info.Server) {
			return resp, err
		}

		// Use the provided zap.Logger for logging but use the fields from context.
		logEntry := logger.With(append(serverCallFields(info.FullMethod), ctxzap.TagsToFields(ctx)...)...)

		if err != nil {
			logEntry.Check(zapcore.ErrorLevel, "payload of failed grpc call").
				Write(protoMessageAsZapField("grpc.request.content", req))
		}
		return resp, err
	}
}

func serverCallFields(fullMethodString string) []zapcore.Field {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	return []zapcore.Field{
		grpc_zap.SystemField,
		grpc_zap.ServerField,
		zap.String("grpc.service", service),
		zap.String("grpc.method", method),
	}
}

func protoMessageAsZapField(key string, pb interface{}) zap.Field {
	p, ok := pb.(proto.Message)
	if !ok {
		return zap.String(key, "not a valid proto.Message")
	}
	if p == nil {
		return zap.String(key, "nil proto.Message")
	}
	return zap.Reflect(key, &jsonpbObjectMarshaler{pb: p})
}

type jsonpbObjectMarshaler struct {
	pb proto.Message
}

func (j *jsonpbObjectMarshaler) MarshalLogObject(e zapcore.ObjectEncoder) error {
	// ZAP jsonEncoder deals with AddReflect by using json.MarshalObject. The same thing applies for consoleEncoder.
	return e.AddReflected("msg", j)
}

func (j *jsonpbObjectMarshaler) MarshalJSON() ([]byte, error) {
	b := &bytes.Buffer{}
	if err := grpc_zap.JsonPbMarshaller.Marshal(b, j.pb); err != nil {
		return nil, fmt.Errorf("jsonpb serializer failed: %v", err)
	}
	return b.Bytes(), nil
}

// codeToLevel is similar to grpc_zap.DefaultCodeToLevel, but bump all non-OK
// levels to at least zap.WarnLevel.
func codeToLevel(code codes.Code) zapcore.Level {
	switch code {
	case codes.OK:
		return zap.InfoLevel
	case codes.Canceled:
		return zap.WarnLevel
	case codes.Unknown:
		return zap.ErrorLevel
	case codes.InvalidArgument:
		return zap.WarnLevel
	case codes.DeadlineExceeded:
		return zap.WarnLevel
	case codes.NotFound:
		return zap.WarnLevel
	case codes.AlreadyExists:
		return zap.WarnLevel
	case codes.PermissionDenied:
		return zap.WarnLevel
	case codes.Unauthenticated:
		return zap.WarnLevel
	case codes.ResourceExhausted:
		return zap.WarnLevel
	case codes.FailedPrecondition:
		return zap.WarnLevel
	case codes.Aborted:
		return zap.WarnLevel
	case codes.OutOfRange:
		return zap.WarnLevel
	case codes.Unimplemented:
		return zap.ErrorLevel
	case codes.Internal:
		return zap.ErrorLevel
	case codes.Unavailable:
		return zap.WarnLevel
	case codes.DataLoss:
		return zap.ErrorLevel
	default:
		return grpc_zap.DefaultCodeToLevel(code)
	}
}
