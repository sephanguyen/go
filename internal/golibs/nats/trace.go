package nats

import (
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

const (
	retryPublishEventTimes = 3
)

// This is used to inject, extract metadata from context to protobuf, and backward
type B3Carrier struct {
	*npb.B3TraceInfo
}

//    b3: {TraceId}-{SpanId}-{SamplingState}-{ParentSpanId}
//  2. Multiple Headers:
//    x-b3-traceid: {TraceId}
//    x-b3-parentspanid: {ParentSpanId}
//    x-b3-spanid: {SpanId}
//    x-b3-sampled: {SamplingState}
//    x-b3-flags: {DebugFlag}
func (w *B3Carrier) Get(key string) string {
	if key != "b3" && w.GetMultiples() == nil {
		return ""
	}
	switch key {
	case "b3":
		return w.GetSingle()
	case "x-b3-traceid":
		return w.GetMultiples().GetTraceId()
	case "x-b3-parentspanid":
		return w.GetMultiples().GetParentSpanId()
	case "x-b3-spanid":
		return w.GetMultiples().GetSpanId()
	case "x-b3-sampled":
		return w.GetMultiples().GetSampled()
	case "x-b3-flags":
		return w.GetMultiples().GetFlags()
	}
	return ""
}
func (w *B3Carrier) Set(key string, val string) {
	if key != "b3" && w.GetMultiples() == nil {
		return
	}
	switch key {
	case "b3":
		w.B3TraceInfo.Header = &npb.B3TraceInfo_Single{Single: val}
	case "x-b3-traceid":
		w.B3TraceInfo.GetMultiples().TraceId = val
	case "x-b3-parentspanid":
		w.B3TraceInfo.GetMultiples().ParentSpanId = val
	case "x-b3-spanid":
		w.B3TraceInfo.GetMultiples().SpanId = val
	case "x-b3-sampled":
		w.B3TraceInfo.GetMultiples().Sampled = val
	case "x-b3-flags":
		w.B3TraceInfo.GetMultiples().Flags = val
	}
}
func (w *B3Carrier) Keys() []string {
	return []string{"b3", "x-b3-traceid", "x-b3-parentspanid", "x-b3-spanid", "x-b3-sampled", "x-b3-flags"}
}
