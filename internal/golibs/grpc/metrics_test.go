package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestNewGRPCServiceMetrics(t *testing.T) {
	t.Run("pass empty ServerInterceptor options", func(t *testing.T) {
		srv := NewGRPCServiceMetrics()
		var grpcUnary []grpc.UnaryServerInterceptor
		var grpcStream []grpc.StreamServerInterceptor
		grpcUnaryRes, grpcStreamRes := srv.FillInterServerInterceptor(grpcUnary, grpcStream)
		assert.Len(t, grpcUnaryRes, 1)
		assert.Len(t, grpcStreamRes, 1)
	})

	t.Run("pass not empty ServerInterceptor options", func(t *testing.T) {
		srv := NewGRPCServiceMetrics()
		grpcUnary := make([]grpc.UnaryServerInterceptor, 1)
		grpcStream := make([]grpc.StreamServerInterceptor, 1)
		grpcUnaryRes, grpcStreamRes := srv.FillInterServerInterceptor(grpcUnary, grpcStream)
		assert.Len(t, grpcUnaryRes, len(grpcUnary)+1)
		assert.Len(t, grpcStreamRes, len(grpcStream)+1)
	})
}
