package interceptors

import (
	"runtime/debug"

	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// WithUnaryServerRecovery wraps grpc_recovery in first and last position of interceptor
func WithUnaryServerRecovery(a []grpc.UnaryServerInterceptor) []grpc.UnaryServerInterceptor {
	customFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "stacktrace from panic: \n"+string(debug.Stack()))
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}

	unaryInterceptors := []grpc.UnaryServerInterceptor{
		grpc_recovery.UnaryServerInterceptor(opts...),
	}

	return append(unaryInterceptors, a...)
}

// WithStreamServerRecovery wraps grpc_recovery in first and last position of interceptor
func WithStreamServerRecovery(a []grpc.StreamServerInterceptor) []grpc.StreamServerInterceptor {
	customFunc := func(p interface{}) (err error) {
		return status.Errorf(codes.Unknown, "stacktrace from panic: \n"+string(debug.Stack()))
	}

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(customFunc),
	}

	streamInterceptors := []grpc.StreamServerInterceptor{
		grpc_recovery.StreamServerInterceptor(opts...),
	}

	return append(streamInterceptors, a...)
}
