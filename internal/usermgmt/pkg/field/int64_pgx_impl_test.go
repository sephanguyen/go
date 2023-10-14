package field

import (
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestInt64_toPGInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Int64
		expectedOutput pgtype.Int8
	}{
		{
			name:  "parse valid Int64 to pgtype.Int8 successfully",
			input: NewInt64(1),
			expectedOutput: pgtype.Int8{
				Int:    1,
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null Int64 to pgtype.Int8 successfully",
			input: NewNullInt64(),
			expectedOutput: pgtype.Int8{
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null Int64 to pgtype.Int8 successfully",
			input: NewUndefinedInt64(),
			expectedOutput: pgtype.Int8{
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, *testCase.input.toPGInt8Ptr())
		})
	}
}

func TestInt64_DecodeText(t *testing.T) {
	value := NewInt64(0)

	err := value.DecodeText(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeText(nil, []byte("1"))
	assert.Nil(t, err)
}

func TestInt64_DecodeBinary(t *testing.T) {
	value := NewInt64(0)

	err := value.DecodeBinary(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeBinary(nil, []byte("00000000"))
	assert.Nil(t, err)
}
