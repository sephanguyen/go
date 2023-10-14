package learnosity

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatUTCTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name           string
		Input          any
		ExpectedOutput any
	}{
		{
			Name:           "UTC time",
			Input:          time.Date(2023, time.December, 15, 12, 01, 01, 01, time.UTC),
			ExpectedOutput: "20231215-1201",
		},
		{
			Name:           "Non-UTC time",
			Input:          time.Date(2023, time.December, 15, 12, 01, 01, 01, time.FixedZone("UTC+7", 7*60*60)),
			ExpectedOutput: "20231215-0501",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			output := FormatUTCTime(tc.Input.(time.Time))
			assert.Equal(t, tc.ExpectedOutput, output)
		})
	}
}

func TestJSONMarshalToString(t *testing.T) {
	t.Parallel()
	_, err := json.Marshal(func() {})

	testCases := []struct {
		Name           string
		Input          any
		ExpectedOutput any
		ExpectedErr    error
	}{
		{
			Name: "Valid struct",
			Input: struct {
				Name  string `json:"name"`
				Age   int    `json:"age"`
				Email string `json:"email"`
			}{
				Name:  "John Doe",
				Age:   25,
				Email: "john.doe@example.com",
			},
			ExpectedOutput: `{"name":"John Doe","age":25,"email":"john.doe@example.com"}`,
			ExpectedErr:    nil,
		},
		{
			Name:           "Invalid struct",
			Input:          func() {},
			ExpectedOutput: "",
			ExpectedErr:    err,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			output, err := JSONMarshalToString(tc.Input)
			if err != nil {
				assert.Equal(t, err, tc.ExpectedErr, err)
			}
			assert.Equal(t, tc.ExpectedOutput, output)
		})
	}
}

func TestContainsKey(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name           string
		Map            map[string]any
		Key            string
		ExpectedOutput any
	}{
		{
			Name: "Map contains key",
			Map: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			Key:            "key1",
			ExpectedOutput: true,
		},
		{
			Name: "Map does not contain key",
			Map: map[string]any{
				"key1": "value1",
				"key2": "value2",
			},
			Key:            "key3",
			ExpectedOutput: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			output := ContainsKey(tc.Map, tc.Key)
			assert.Equal(t, tc.ExpectedOutput, output)
		})
	}
}
