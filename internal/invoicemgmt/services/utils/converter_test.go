package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCurrency(t *testing.T) {
	testCase := []struct {
		name        string
		given       float64
		expected    string
		expectedErr error
	}{
		{
			name:     "Test Positive 3-digit value",
			given:    100,
			expected: "100",
		},
		{
			name:     "Test Positive 3-digit value with decimal",
			given:    100.53,
			expected: "100.53",
		},
		{
			name:     "Test Positive 4-digit value",
			given:    1020,
			expected: "1,020",
		},
		{
			name:     "Test Positive 4-digit value with decimal",
			given:    5392.69,
			expected: "5,392.69",
		},
		{
			name:     "Test large value",
			given:    123456789,
			expected: "123,456,789",
		},
		{
			name:     "Test large value with decimal",
			given:    123456789.69,
			expected: "123,456,789.69",
		},
		{
			name:     "Test Negative 3-digit value",
			given:    -100,
			expected: "-100",
		},
		{
			name:     "Test Negative 3-digit value with decimal",
			given:    -100.32,
			expected: "-100.32",
		},
		{
			name:     "Test Negative 2-digit value",
			given:    -10,
			expected: "-10",
		},
		{
			name:     "Test zero value",
			given:    0,
			expected: "0",
		},
		{
			name:     "Test Negative 4-digit value",
			given:    -1000,
			expected: "-1,000",
		},
		{
			name:     "Test Negative 4-digit value with decimal",
			given:    -1000.55,
			expected: "-1,000.55",
		},
		{
			name:     "Test Negative large value",
			given:    -1121231232,
			expected: "-1,121,231,232",
		},
		{
			name:     "Test Negative large value with decimal",
			given:    -1121231232.11,
			expected: "-1,121,231,232.11",
		},
	}

	for _, tc := range testCase {
		actual := FormatCurrency(tc.given)
		assert.Equal(t, tc.expected, actual)
	}
}
