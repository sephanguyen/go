package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStreamingProvider_isValid(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		name              string
		streamingProvider *StreamingProvider
		isValid           bool
	}{
		{
			name: "full fields",
			streamingProvider: &StreamingProvider{
				StreamingRoomID:     "test-id-1",
				TotalStreamingSlots: 1,
			},
			isValid: true,
		},
		{
			name: "nil id",
			streamingProvider: &StreamingProvider{
				StreamingRoomID:     "",
				TotalStreamingSlots: 1,
			},
			isValid: false,
		},
		{
			name: "0 slot",
			streamingProvider: &StreamingProvider{
				StreamingRoomID:     "test-id-1",
				TotalStreamingSlots: 0,
			},
			isValid: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.streamingProvider.isValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
