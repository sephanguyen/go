package kafka

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestB3TraceCarrierImpl_Keys(t *testing.T) {
	t.Run("happy case", func(t *testing.T) {
		carrier := B3TraceCarrierImpl{}
		assert.Equal(t, []string{b3ContextHeader, b3DebugFlagHeader, b3TraceIDHeader, b3SpanIDHeader, b3SampledHeader, b3ParentSpanIDHeader}, carrier.Keys())
	})
}

func TestB3TraceCarrierImpl_GetSet(t *testing.T) {
	t.Run("happy case: single", func(t *testing.T) {
		carrier := B3TraceCarrierImpl{}

		carrier.Set(b3ContextHeader, "b3-test")
		b3HeaderVal := carrier.Get(b3ContextHeader)

		assert.Equal(t, carrier.tracesInfo[b3ContextHeader], b3HeaderVal)
	})

	t.Run("happy case: multiples", func(t *testing.T) {
		carrier := B3TraceCarrierImpl{}
		carrier.Set(b3DebugFlagHeader, "b3-debug-flag")
		carrier.Set(b3TraceIDHeader, "b3-trace-id")
		carrier.Set(b3SpanIDHeader, "b3-span-id")
		carrier.Set(b3SampledHeader, "b3-sampled")
		carrier.Set(b3ParentSpanIDHeader, "b3-parent-span-id")

		b3DebugFlagVal := carrier.Get(b3DebugFlagHeader)
		b3TraceIDVal := carrier.Get(b3TraceIDHeader)
		b3SpanIDVal := carrier.Get(b3SpanIDHeader)
		b3SampledVal := carrier.Get(b3SampledHeader)
		b3ParentSpanIDVal := carrier.Get(b3ParentSpanIDHeader)

		assert.Equal(t, carrier.tracesInfo[b3DebugFlagHeader], b3DebugFlagVal)
		assert.Equal(t, carrier.tracesInfo[b3TraceIDHeader], b3TraceIDVal)
		assert.Equal(t, carrier.tracesInfo[b3SpanIDHeader], b3SpanIDVal)
		assert.Equal(t, carrier.tracesInfo[b3SampledHeader], b3SampledVal)
		assert.Equal(t, carrier.tracesInfo[b3ParentSpanIDHeader], b3ParentSpanIDVal)
	})

	t.Run("case: get not-exist value", func(t *testing.T) {
		carrier := B3TraceCarrierImpl{}
		emptyVal := carrier.Get(b3ContextHeader)
		assert.Equal(t, "", emptyVal)
	})

	t.Run("case: get out of key", func(t *testing.T) {
		carrier := B3TraceCarrierImpl{}
		emptyVal := carrier.Get("wrong-key")
		assert.Equal(t, "", emptyVal)
	})

	t.Run("case: set/get out of key", func(t *testing.T) {
		carrier := B3TraceCarrierImpl{}
		carrier.Set("wrong-key", "wrong-value")
		wrongVal := carrier.Get("wrong-key")
		assert.Equal(t, "", wrongVal)
	})
}
