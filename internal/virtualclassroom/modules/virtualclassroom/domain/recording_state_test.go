package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRecordingState_isValid(t *testing.T) {
	t.Parallel()
	mockName := "test-name"
	tcs := []struct {
		name           string
		recordingState *RecordingState
		isValid        bool
	}{
		{
			name: "full fields",
			recordingState: &RecordingState{
				IsRecording: true,
				Creator:     &mockName,
			},
			isValid: true,
		},
		{
			name: "nil creator",
			recordingState: &RecordingState{
				IsRecording: true,
				Creator:     nil,
			},
			isValid: false,
		},
		{
			name: "false status but creator not null",
			recordingState: &RecordingState{
				IsRecording: false,
				Creator:     &mockName,
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.recordingState.isValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
