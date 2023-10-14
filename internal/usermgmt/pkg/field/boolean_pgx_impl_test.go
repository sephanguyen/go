package field

import (
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestBoolean_toPGBoolean(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Boolean
		expectedOutput pgtype.Bool
	}{
		{
			name:  "parse valid Boolean to pgtype.Bool successfully",
			input: NewBoolean(true),
			expectedOutput: pgtype.Bool{
				Bool:   true,
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null Boolean to pgtype.Bool successfully",
			input: NewNullBoolean(),
			expectedOutput: pgtype.Bool{
				Bool:   false,
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null Boolean to pgtype.Bool successfully",
			input: NewUndefinedBoolean(),
			expectedOutput: pgtype.Bool{
				Bool:   false,
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, *testCase.input.toPGBooleanPtr())
		})
	}
}

func TestBoolean_DecodeText(t *testing.T) {
	value := NewBoolean(true)

	err := value.DecodeText(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeText(nil, []byte("1"))
	assert.Nil(t, err)
}

func TestBoolean_DecodeBinary(t *testing.T) {
	value := NewBoolean(true)

	err := value.DecodeBinary(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeBinary(nil, []byte("1"))
	assert.Nil(t, err)
}
