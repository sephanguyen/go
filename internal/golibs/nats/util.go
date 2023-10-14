package nats

import (
	"context"

	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"

	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type resourcePath int

const (
	resourcePathKey resourcePath = iota
)

func TraceInfoFromContext(ctx context.Context) *npb.B3TraceInfo {
	carrier := &B3Carrier{&npb.B3TraceInfo{}}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	return carrier.B3TraceInfo
}

func ContextWithTraceInfo(ctx context.Context, traceInfo *npb.B3TraceInfo) context.Context {
	carrier := &B3Carrier{traceInfo}
	newCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)
	return newCtx
}

func WarningIfError(l *zap.Logger, err error, msg string) {
	if err == nil {
		return
	}
	l.Warn(msg, zap.Error(err))
}
