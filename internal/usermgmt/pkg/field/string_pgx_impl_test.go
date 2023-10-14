package field

import (
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestString_toPGText(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          String
		expectedOutput pgtype.Text
	}{
		{
			name:  "parse valid String to pgtype.Text successfully",
			input: NewString("example"),
			expectedOutput: pgtype.Text{
				String: "example",
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null String to pgtype.Text successfully",
			input: NewNullString(),
			expectedOutput: pgtype.Text{
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null String to pgtype.Text successfully",
			input: NewUndefinedString(),
			expectedOutput: pgtype.Text{
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, testCase.input.toPGText())
		})
	}
}
