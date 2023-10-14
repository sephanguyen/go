package nats

import (
	"context"
	"testing"

	"github.com/golang/protobuf/proto"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
	"github.com/stretchr/testify/assert"
)

func Test_MarshalUnmarshal(t *testing.T) {

	t.Run("marshal with metadata, unmarshal ignore metadata", func(t *testing.T) {

		// borrow a protobuf type used as service internal msg
		internalB3 := &npb.B3TraceInfo{
			Header: &npb.B3TraceInfo_Single{
				Single: "test-single-content",
			},
		}

		raw, err := MarshalWithContext(context.Background(), internalB3)
		assert.NoError(t, err)

		internalB3v2 := &npb.B3TraceInfo{}
		err = UnmarshalIgnoreMetadata(raw, internalB3v2)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(internalB3v2, internalB3))
	})
	t.Run("marshal without metadata, unmarshal ignore metadata", func(t *testing.T) {
		// borrow a protobuf type used as service internal msg
		internalB3 := &npb.B3TraceInfo{
			Header: &npb.B3TraceInfo_Single{
				Single: "test-single-content",
			},
		}
		raw, err := proto.Marshal(internalB3)
		assert.NoError(t, err)

		internalB3v2 := &npb.B3TraceInfo{}
		err = UnmarshalIgnoreMetadata(raw, internalB3v2)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(internalB3v2, internalB3))
	})
	t.Run("marshal with metadata, unmarshal with metadata", func(t *testing.T) {
		// borrow a protobuf type used as service internal msg
		msgv1 := &npb.B3TraceInfo{
			Header: &npb.B3TraceInfo_Single{
				Single: "test-single-content",
			},
		}
		raw, err := MarshalWithContext(context.Background(), msgv1)
		assert.NoError(t, err)

		msgv2 := &npb.B3TraceInfo{}
		_, rawMsgBytes, err := UnmarshalWithContext(context.Background(), raw)
		assert.NoError(t, err)
		err = proto.Unmarshal(rawMsgBytes, msgv2)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(msgv2, msgv1), "internal messages are not equal")
	})
}
