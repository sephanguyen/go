package interceptors

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// TimeoutUnaryServerInterceptor returns a new unary server interceptors that sets the context timeout.
func TimeoutUnaryServerInterceptor(d time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		newCtx, cancel := context.WithTimeout(ctx, d)
		defer cancel()

		return handler(newCtx, req)
	}
}

const defaultTimeoutKey string = "default"

// TimeoutUnaryServerInterceptorV2 is similar to TimeoutUnaryServerInterceptor, but allows
// specifying different timeout for different GRPC APIs.
func TimeoutUnaryServerInterceptorV2(l *zap.Logger, c configs.GRPCHandlerTimeoutV2) grpc.UnaryServerInterceptor {
	defaultTimeout, defaultAvailable := c[defaultTimeoutKey]
	if !defaultAvailable {
		l.Warn("missing default value for GRPC timeout, please set \"common.grpc.handler_timeout_v2.default\" in your config")
		defaultTimeout = -1 // means no timeout
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		methodName := info.FullMethod
		timeout, ok := c[methodName]
		if !ok {
			timeout = defaultTimeout
		}
		if timeout < 0 {
			return handler(ctx, req)
		}

		childCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return handler(childCtx, req)
	}
}
