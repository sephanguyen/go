package grpc

import (
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
)

type ServerMetrics struct {
	grpc_prometheus.ServerMetrics
}

func (s *ServerMetrics) FillInterServerInterceptor(
	grpcUnary []grpc.UnaryServerInterceptor,
	grpcStream []grpc.StreamServerInterceptor,
) (
	[]grpc.UnaryServerInterceptor,
	[]grpc.StreamServerInterceptor,
) {
	if grpcUnary == nil {
		grpcUnary = make([]grpc.UnaryServerInterceptor, 0, 1)
	}
	grpcUnary = append(grpcUnary, s.UnaryServerInterceptor())

	if grpcStream == nil {
		grpcStream = make([]grpc.StreamServerInterceptor, 0, 1)
	}
	grpcStream = append(grpcStream, s.StreamServerInterceptor())

	return grpcUnary, grpcStream
}

func (s ServerMetrics) Register(server *grpc.Server) {
	s.InitializeMetrics(server)
}

func NewGRPCServiceMetrics() *ServerMetrics {
	srvMetrics := grpc_prometheus.NewServerMetrics()
	srvMetrics.EnableHandlingTimeHistogram()
	return &ServerMetrics{
		*srvMetrics,
	}
}
