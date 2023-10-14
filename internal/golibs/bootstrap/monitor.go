package bootstrap

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/golibs/nats"

	"github.com/prometheus/client_golang/prometheus"
	"go.opencensus.io/stats/view"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/b3"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// MonitorServicer Implement this interface to expose
// a metrics api to monitor your service (usually exposed at port 8888)
type MonitorServicer[T any] interface {
	// Provide us with your own metric collector, if any
	WithPrometheusCollectors(*Resources) []prometheus.Collector
	// Provide us with your own opencensus, if any
	WithOpencensusViews() []*view.View

	// Provide us init some metrics on our code, if any
	InitMetricsValue()
}

// DefaultMonitorService
type DefaultMonitorService[T any] struct{}

func (DefaultMonitorService[T]) WithPrometheusCollectors(*Resources) []prometheus.Collector {
	return nil
}

func (DefaultMonitorService[T]) WithOpencensusViews() []*view.View {
	return nil
}

func (DefaultMonitorService[T]) InitMetricsValue() {
}

func (b *bootstrapper[T]) setupMonitor(monitorServicer MonitorServicer[T], c *T, rsc *Resources) error {
	_, ok := interface{}(monitorServicer).(NatsServicer[T])
	var ocViews []*view.View
	ocViews = append(ocViews, interceptors.GrpcServerViews...)

	if rsc.natsjs != nil && ok {
		ocViews = append(ocViews, DefaultNatsOpencensusViews()...)
	}

	ocViews = append(ocViews, monitorServicer.WithOpencensusViews()...)
	if err := view.Register(ocViews...); err != nil {
		return fmt.Errorf("setupMonitor: Failed to register ocgrpc server views: %w", err)
	}

	cc, err := extract[configs.CommonConfig](c, commonFieldName)
	if err != nil {
		return fmt.Errorf("setupMonitor: failed to extract configs.CommonConfig, error: %w", err)
	}

	pp, tp, err := newTelemetry(cc)
	if err != nil {
		return fmt.Errorf("setupMonitor: newTelemetry failed: %w", err)
	}

	if pp != nil {
		if err = pp.AddCollectors(monitorServicer.WithPrometheusCollectors(rsc)...); err != nil {
			return fmt.Errorf("could not add collectors for apis: %v", err)
		}
		// Register Prometheus metrics handler.
		go func() {
			if err := interceptors.StartMetricHandlerWithProvider("/metrics", ":8888", pp); err != nil {
				rsc.Logger().Info("setupMonitor: interceptors.StartMetricHandlerWithProvider failed", zap.Error(err))
			}
		}()

		monitorServicer.InitMetricsValue()
	}

	if tp != nil {
		// grpc telemetry handler
		_, ok := interface{}(monitorServicer).(GRPCServicer[T])
		if ok {
			grpcUnary := []grpc.UnaryServerInterceptor{otelgrpc.UnaryServerInterceptor(
				otelgrpc.WithPropagators(b3.New(b3.WithInjectEncoding(b3.B3SingleHeader|b3.B3MultipleHeader))),
				otelgrpc.WithTracerProvider(tp),
			)}

			b.unaryInterceptors = append(b.unaryInterceptors, grpcUnary...)
		}

		b.tp = tp
	}

	return nil
}

func newTelemetry(cc *configs.CommonConfig) (*interceptors.PrometheusProvider, *tracesdk.TracerProvider, error) {
	pp, tp, err := interceptors.NewTelemetry(cc.RemoteTrace.OtelCollectorReceiver, cc.Name, 1)
	if !cc.StatsEnabled {
		pp = nil
	}

	if !cc.RemoteTrace.Enabled || len(cc.RemoteTrace.OtelCollectorReceiver) == 0 {
		tp = nil
		err = nil
	}

	return pp, tp, err
}

func DefaultNatsOpencensusViews() []*view.View {
	return []*view.View{
		nats.JetstreamProcessedMessagesView,
		nats.JetstreamProcessedMessagesLatencyView,
	}
}
