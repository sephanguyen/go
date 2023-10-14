package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestZegoSDKError(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name     string
		code     string
		message  string
		hasError bool
	}{
		{
			name:    "zego sdk error code and message not empty",
			code:    JSONUnmarshalErrorCode,
			message: "test message",
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			err := NewZegoSDKError(tc.code, tc.message)
			zcSDKErr := &ZegoSDKError{
				Code:    tc.code,
				Message: tc.message,
			}

			require.NotNil(t, err)
			require.NotNil(t, zcSDKErr.GetCode())
			require.NotNil(t, zcSDKErr.GetMessage())
		})
	}
}
