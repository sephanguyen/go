package field

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewTimeWithoutTz(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		input          time.Time
		expectedOutput TimeWithoutTz
	}{
		{
			name:  "init TimeWithoutTz with zero value",
			input: time.Time{},
			expectedOutput: TimeWithoutTz{
				value:  time.Time{},
				status: StatusPresent,
			},
		},
		{
			name:  "init TimeWithoutTz with valid value",
			input: now,
			expectedOutput: TimeWithoutTz{
				value:  now,
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewTimeWithoutTz(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullTimeWithoutTz(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput TimeWithoutTz
	}{
		{
			name: "init null TimeWithoutTz",
			expectedOutput: TimeWithoutTz{
				value:  time.Time{},
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullTimeWithoutTz() == testCase.expectedOutput)
		})
	}
}
