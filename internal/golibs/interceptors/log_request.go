package interceptors

import (
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

type WrappedJSONMarshaler struct {
	*jsonpb.Marshaler
}

func (w *WrappedJSONMarshaler) Marshal(out io.Writer, pb proto.Message) error {
	return w.Marshaler.Marshal(out, pb)
}
