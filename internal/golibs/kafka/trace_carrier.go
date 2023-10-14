package kafka

import (
	"k8s.io/utils/strings/slices"
)

// Lib: https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/propagators/b3
// Ref: https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main

const (
	// Default B3 Header names.
	b3ContextHeader      = "b3"
	b3DebugFlagHeader    = "x-b3-flags"
	b3TraceIDHeader      = "x-b3-traceid"
	b3SpanIDHeader       = "x-b3-spanid"
	b3SampledHeader      = "x-b3-sampled"
	b3ParentSpanIDHeader = "x-b3-parentspanid"
)

type B3TraceCarrierImpl struct {
	tracesInfo map[string]string
}

func (c *B3TraceCarrierImpl) Keys() []string {
	return []string{b3ContextHeader, b3DebugFlagHeader, b3TraceIDHeader, b3SpanIDHeader, b3SampledHeader, b3ParentSpanIDHeader}
}

func (c *B3TraceCarrierImpl) Get(key string) string {
	if !slices.Contains(c.Keys(), key) {
		return ""
	}

	if val, ok := c.tracesInfo[key]; ok {
		return val
	}

	return ""
}

func (c *B3TraceCarrierImpl) Set(key string, val string) {
	if !slices.Contains(c.Keys(), key) {
		return
	}

	if c.tracesInfo == nil {
		c.tracesInfo = make(map[string]string)
	}

	if val != "" {
		c.tracesInfo[key] = val
	}
}

func (c *B3TraceCarrierImpl) GetAllValues() map[string]string {
	return c.tracesInfo
}
