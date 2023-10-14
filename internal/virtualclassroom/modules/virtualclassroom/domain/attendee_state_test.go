package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAttendeeState_isValid(t *testing.T) {
	now := time.Now()
	t.Parallel()
	tcs := []struct {
		name          string
		attendeeState *AttendeeState
		isValid       bool
	}{
		{
			name: "full fields",
			attendeeState: &AttendeeState{
				UserID: "user-id",
				RaisingHandState: &AttendeeRaisingHandState{
					IsRaisingHand: true,
					UpdatedAt:     now,
				},
				AnnotationState: &AttendeeAnnotationState{
					BeAllowed: true,
					UpdatedAt: now,
				},
				PollingAnswer: &AttendeePollingAnswerState{
					Answer:    []string{"a", "b"},
					UpdatedAt: now,
				},
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			attendeeState: &AttendeeState{
				UserID: "user-id",
				RaisingHandState: &AttendeeRaisingHandState{
					IsRaisingHand: true,
					UpdatedAt:     now,
				},
				AnnotationState: &AttendeeAnnotationState{
					BeAllowed: true,
					UpdatedAt: now,
				},
			},
			isValid: true,
		},
		{
			name: "miss user id field",
			attendeeState: &AttendeeState{
				RaisingHandState: &AttendeeRaisingHandState{
					IsRaisingHand: true,
					UpdatedAt:     now,
				},
				AnnotationState: &AttendeeAnnotationState{
					BeAllowed: true,
					UpdatedAt: now,
				},
				PollingAnswer: &AttendeePollingAnswerState{
					Answer:    []string{"a", "b"},
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
		{
			name: "miss raising hand state field",
			attendeeState: &AttendeeState{
				UserID: "user-id",
				AnnotationState: &AttendeeAnnotationState{
					BeAllowed: true,
					UpdatedAt: now,
				},
				PollingAnswer: &AttendeePollingAnswerState{
					Answer:    []string{"a", "b"},
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
		{
			name: "miss annotation state field",
			attendeeState: &AttendeeState{
				UserID: "user-id",
				RaisingHandState: &AttendeeRaisingHandState{
					IsRaisingHand: true,
					UpdatedAt:     now,
				},
				PollingAnswer: &AttendeePollingAnswerState{
					Answer:    []string{"a", "b"},
					UpdatedAt: now,
				},
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.attendeeState.isValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
