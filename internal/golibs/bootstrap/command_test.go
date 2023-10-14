package bootstrap

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	prometheus "github.com/prometheus/client_golang/prometheus"
	view "go.opencensus.io/stats/view"
	"google.golang.org/grpc"
)

// TestRegisterable compile-time checks only.
func TestRegisterable(t *testing.T) {
	t.Parallel()

	s := &ts{}

	_ = WithGRPC[c](s)
	_ = WithHTTP[c](s)
	_ = WithNatsServicer[c](s)
	_ = WithKafkaServicer[c](s)
	_ = WithMonitorServicer[c](s)
}

type (
	c  struct{} // test config
	ts struct{} // test service
)

func (*ts) SetupGRPC(context.Context, *grpc.Server, c, *Resources) error {
	return nil
}
func (*ts) WithServerOptions() []grpc.ServerOption { return nil }
func (*ts) WithUnaryServerInterceptors(c, *Resources) []grpc.UnaryServerInterceptor {
	return nil
}

func (*ts) WithStreamServerInterceptors(c, *Resources) []grpc.StreamServerInterceptor {
	return nil
}
func (*ts) SetupHTTP(c, *gin.Engine, *Resources) error                   { return nil }
func (*ts) RegisterNatsSubscribers(context.Context, c, *Resources) error { return nil }
func (*ts) InitKafkaConsumers(context.Context, c, *Resources) error      { return nil }
func (*ts) WithPrometheusCollectors(*Resources) []prometheus.Collector   { return nil }
func (*ts) WithOpencensusViews() []*view.View                            { return nil }
func (*ts) InitMetricsValue()                                            {}
