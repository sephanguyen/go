package tracer

import (
	"context"
	"encoding/hex"

	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
	"go.opencensus.io/trace/propagation"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"
)

const traceContextKey = "grpc-trace-bin"

// B3Handler wraps B3 header to out going context
type B3Handler struct {
	*ocgrpc.ClientHandler
}

// TagRPC calls ClientHandler.TagRPC after inject b3 header
func (h *B3Handler) TagRPC(ctx context.Context, rti *stats.RPCTagInfo) context.Context {
	ctx = h.ClientHandler.TagRPC(ctx, rti)
	span := trace.FromContext(ctx)
	spanCtx := span.SpanContext()

	traceContextBinary := propagation.Binary(spanCtx)
	sampled := "0"
	if spanCtx.IsSampled() {
		sampled = "1"
	}
	return metadata.AppendToOutgoingContext(ctx,
		traceContextKey, string(traceContextBinary),
		b3.TraceIDHeader, hex.EncodeToString(spanCtx.TraceID[:]),
		b3.SpanIDHeader, hex.EncodeToString(spanCtx.SpanID[:]),
		b3.SampledHeader, sampled,
	)
}
