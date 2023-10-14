package interceptors

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"

	ocprom "contrib.go.opencensus.io/exporter/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

var MillisecondsDistribution = view.Distribution(5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000, 10000)

var serverLatencyView = &view.View{
	Name:        "grpc.io/server/server_latency",
	Description: "Distribution of server latency in milliseconds, by method.",
	TagKeys:     []tag.Key{ocgrpc.KeyServerMethod},
	Measure:     ocgrpc.ServerLatency,
	Aggregation: MillisecondsDistribution,
}

// GrpcServerViews default grpc service view
var GrpcServerViews = []*view.View{
	ocgrpc.ServerReceivedBytesPerRPCView,
	ocgrpc.ServerSentBytesPerRPCView,
	serverLatencyView,
	ocgrpc.ServerCompletedRPCsView,
	ocgrpc.ServerReceivedMessagesPerRPCView,
	ocgrpc.ServerSentMessagesPerRPCView,
}

type ignoredTraceCtxKey struct{}

func IgnoredTraceCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, ignoredTraceCtxKey{}, struct{}{})
}

func isCtxIgnoringTracing(ctx context.Context) bool {
	return ctx.Value(ignoredTraceCtxKey{}) != nil
}

type TimedSpan struct {
	span      trace.Span
	startTime time.Time
}

const MaxTime = 60000 // in ms

var globalTracer trace.Tracer

type initTelemetryOpt struct {
	collectors []prometheus.Collector
}

func WithCollectors(collectors []prometheus.Collector) InitTelemetryOpt {
	return func(opt *initTelemetryOpt) {
		opt.collectors = append(opt.collectors, collectors...)
	}
}

type InitTelemetryOpt func(*initTelemetryOpt)

// Deprecated: Use NewTelemetry instead.
// InitTelemetry starts prometheus and jaeger exporter
func InitTelemetry(c *configs.CommonConfig, serviceName string, sampleRate float64, opts ...InitTelemetryOpt) (pe *ocprom.Exporter, tp *tracesdk.TracerProvider, err error) {
	initopt := &initTelemetryOpt{}
	for _, f := range opts {
		f(initopt)
	}

	if c.StatsEnabled {
		registry := prometheus.NewRegistry()
		registry.MustRegister(collectors.NewGoCollector())
		for _, item := range initopt.collectors {
			registry.MustRegister(item)
		}
		pe, err = ocprom.NewExporter(ocprom.Options{
			Registry: registry,
		})
		if err != nil {
			err = fmt.Errorf("failed to create Prometheus exporter: %v", err)
			return
		}

		view.RegisterExporter(pe)
		view.SetReportingPeriod(30 * time.Second)
	}

	if c.RemoteTrace.Enabled && c.RemoteTrace.OtelCollectorReceiver != "" {
		exp, jerr := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(c.RemoteTrace.OtelCollectorReceiver)))
		if jerr != nil {
			err = fmt.Errorf("failed to create Jaeger exporter: %v", jerr)
			return
		}
		tp = tracesdk.NewTracerProvider(
			tracesdk.WithBatcher(exp),
			tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(sampleRate))),
			tracesdk.WithResource(resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				attribute.String("exporter", "jaeger"),
			)),
		)

		otel.SetTracerProvider(tp)
		globalTracer = otel.Tracer(serviceName)
	}

	otel.SetTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3SingleHeader | b3.B3MultipleHeader)))

	return
}

// StartMetricHandler start prometheus HTTP handler
func StartMetricHandler(path, port string, pe *ocprom.Exporter) {
	if pe == nil {
		return
	}
	mux := http.NewServeMux()
	mux.Handle(path, pe)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Println("error starting prometheus handler", err)
	}
}

// StartMetricHandlerWithProvider start prometheus HTTP handler
func StartMetricHandlerWithProvider(path, port string, pp *PrometheusProvider) error {
	if pp == nil {
		return nil
	}
	mux := http.NewServeMux()
	ep, err := pp.Exporter()
	if err != nil {
		return fmt.Errorf("could not get prometheus exporter: %w", err)
	}
	mux.Handle(path, ep)
	if err = http.ListenAndServe(port, mux); err != nil {
		return fmt.Errorf("error starting prometheus handler: %w", err)
	}

	return nil
}

func StartSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, TimedSpan) {
	if isCtxIgnoringTracing(ctx) {
		tracer := trace.NewNoopTracerProvider().Tracer("")
		ctx2, span := tracer.Start(ctx, "")
		return ctx2, TimedSpan{span: span}
	}
	if globalTracer == nil {
		globalTracer = otel.Tracer("")
	}
	ctx, span := globalTracer.Start(ctx, spanName, opts...)
	return ctx, TimedSpan{
		span:      span,
		startTime: time.Now(),
	}
}
func (t *TimedSpan) SetAttributes(attrs ...attribute.KeyValue) {
	t.span.SetAttributes(attrs...)
}
func (t *TimedSpan) Span() trace.Span {
	return t.span
}
func (t *TimedSpan) RecordError(err error) {
	t.span.RecordError(err)
}

func (t *TimedSpan) End() {
	duration := time.Since(t.startTime).Milliseconds()
	// attrExecutionTime := attribute.Int64("x-execution-time", duration)
	// t.span.SetAttributes(attrExecutionTime)
	if duration > MaxTime {
		t.span.SetAttributes(attribute.String("x-timed-out", "true"))
	}
	t.span.End()
}

type IgnoredTraceInterceptor struct {
	ignoredMethods map[string]struct{}
}

func NewIgnoredTraceInterceptor(methods map[string]struct{}) *IgnoredTraceInterceptor {
	return &IgnoredTraceInterceptor{ignoredMethods: methods}
}

func (i *IgnoredTraceInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if _, ignored := i.ignoredMethods[info.FullMethod]; ignored {
		ctx = IgnoredTraceCtx(ctx)
		return handler(ctx, req)
	}
	return handler(ctx, req)
}

// NewTelemetry will return prometheus and jaeger provider
func NewTelemetry(otelCollectorReceiver, serviceName string, sampleRate float64) (*PrometheusProvider, *tracesdk.TracerProvider, error) {
	metrics := NewPrometheusProvider()

	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(otelCollectorReceiver)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Jaeger exporter: %+v", err)
	}
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithSampler(tracesdk.ParentBased(tracesdk.TraceIDRatioBased(sampleRate))),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			attribute.String("exporter", "jaeger"),
		)),
	)

	globalTracer = otel.Tracer(serviceName)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(b3.New(b3.WithInjectEncoding(b3.B3SingleHeader | b3.B3MultipleHeader)))

	return metrics, tp, nil
}

type PrometheusProvider struct {
	Registry *prometheus.Registry
}

func NewPrometheusProvider() *PrometheusProvider {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	return &PrometheusProvider{
		Registry: registry,
	}
}

func (p *PrometheusProvider) AddCollectors(collectors ...prometheus.Collector) error {
	for _, item := range collectors {
		if err := p.Registry.Register(item); err != nil {
			return err
		}
	}

	return nil
}

func (p *PrometheusProvider) Exporter() (*ocprom.Exporter, error) {
	pe, err := ocprom.NewExporter(ocprom.Options{
		Registry: p.Registry,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create the Prometheus stats exporter: %v", err)
	}

	view.RegisterExporter(pe)
	view.SetReportingPeriod(30 * time.Second)

	return pe, nil
}

func (p *PrometheusProvider) Register(c prometheus.Collector) error {
	return p.Registry.Register(c)
}
