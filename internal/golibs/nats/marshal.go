package nats

import (
	"context"
	"fmt"

	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	"go.opentelemetry.io/otel"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// UnmarshalIgnoreMetadata mostly used by tests, that only care about unwrapped message
func UnmarshalIgnoreMetadata(raw []byte, dst proto.Message) error {
	wrapped := &npb.WrapperMsg{}
	err := proto.Unmarshal(raw, wrapped)
	if err != nil {
		// try unmarshaling normally
		return proto.Unmarshal(raw, dst)
	}
	return anypb.UnmarshalTo(wrapped.GetMessage(), dst, proto.UnmarshalOptions{})
}

// MarshalWithContrext inject metadata such as resource_path, tracing from context to protobuf msg
func MarshalWithContext(ctx context.Context, msg proto.Message) ([]byte, error) {
	anymsg, err := anypb.New(msg)
	if err != nil {
		return nil, err
	}
	wrappedMsg := &npb.WrapperMsg{
		Message: anymsg,
	}

	// mandatory context injection goes here (RLS)
	carrier := &B3Carrier{&npb.B3TraceInfo{}}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	wrappedMsg.TraceInfo = carrier.B3TraceInfo

	return proto.Marshal(wrappedMsg)
}

// UnmarshalWithContext extract metadata (resource_path,tracing...) from protobuf to context
// and return raw byte of core messages
func UnmarshalWithContext(parentCtx context.Context, raw []byte) (context.Context, []byte, error) {
	wrapped := &npb.WrapperMsg{}
	err := proto.Unmarshal(raw, wrapped)
	if err != nil {
		return parentCtx, nil, err
	}
	if wrapped.GetMessage() == nil {
		return parentCtx, nil, fmt.Errorf("nil core message")
	}

	// trace info
	carrier := &B3Carrier{wrapped.GetTraceInfo()}
	newCtx := otel.GetTextMapPropagator().Extract(parentCtx, carrier)

	// resource_path ... TODO

	return newCtx, wrapped.GetMessage().GetValue(), nil
}
