package field

import (
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestInt32_toPGInt32(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Int32
		expectedOutput pgtype.Int4
	}{
		{
			name:  "parse valid Int32 to pgtype.Int4 successfully",
			input: NewInt32(1),
			expectedOutput: pgtype.Int4{
				Int:    1,
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null Int32 to pgtype.Int4 successfully",
			input: NewNullInt32(),
			expectedOutput: pgtype.Int4{
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null Int32 to pgtype.Int4 successfully",
			input: NewUndefinedInt32(),
			expectedOutput: pgtype.Int4{
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, *testCase.input.toPGInt4Ptr())
		})
	}
}

func TestInt32_DecodeText(t *testing.T) {
	value := NewInt32(0)

	err := value.DecodeText(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeText(nil, []byte("1"))
	assert.Nil(t, err)
}

func TestInt32_DecodeBinary(t *testing.T) {
	value := NewInt32(0)

	err := value.DecodeBinary(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeBinary(nil, []byte("0000"))
	assert.Nil(t, err)
}
