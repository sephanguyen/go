package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCurrentPolling_isValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name           string
		currentPolling *CurrentPolling
		isValid        bool
	}{
		{
			name: "full fields",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				EndedAt:   &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusStarted,
			},
			isValid: true,
		},
		{
			name: "only have required fields",
			currentPolling: &CurrentPolling{
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: true,
					},
				},
				Status: CurrentPollingStatusStarted,
			},
			isValid: true,
		},
		{
			name: "have duplicated right option",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				EndedAt:   &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusStarted,
			},
			isValid: true,
		},
		{
			name: "have duplicated wrong option",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				EndedAt:   &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusStarted,
			},
			isValid: true,
		},
		{
			name: "have duplicated option but different value",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				EndedAt:   &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusStarted,
			},
			isValid: false,
		},
		{
			name: "these are no any right option",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				EndedAt:   &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: false,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusStarted,
			},
			isValid: false,
		},
		{
			name: "miss option fields",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				EndedAt:   &now,
				Status:    CurrentPollingStatusStarted,
			},
			isValid: false,
		},
		{
			name: "miss status field",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				EndedAt:   &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
			},
			isValid: false,
		},
		{
			name: "have status is stopped but have no stopped at field",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				EndedAt:   &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusStopped,
			},
			isValid: false,
		},
		{
			name: "have status is ended but have no ended at field",
			currentPolling: &CurrentPolling{
				CreatedAt: now,
				UpdatedAt: now,
				StoppedAt: &now,
				Options: CurrentPollingOptions{
					{
						Answer:    "A",
						IsCorrect: false,
					},
					{
						Answer:    "B",
						IsCorrect: true,
					},
					{
						Answer:    "C",
						IsCorrect: false,
					},
				},
				Status: CurrentPollingStatusEnded,
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.currentPolling.isValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
