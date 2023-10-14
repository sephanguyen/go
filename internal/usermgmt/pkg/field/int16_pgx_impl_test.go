package field

import (
	"testing"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestInt16_toPGInt16(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Int16
		expectedOutput pgtype.Int2
	}{
		{
			name:  "parse valid Int16 to pgtype.Int2 successfully",
			input: NewInt16(1),
			expectedOutput: pgtype.Int2{
				Int:    1,
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null Int16 to pgtype.Int2 successfully",
			input: NewNullInt16(),
			expectedOutput: pgtype.Int2{
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null Int16 to pgtype.Int2 successfully",
			input: NewUndefinedInt16(),
			expectedOutput: pgtype.Int2{
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, *testCase.input.toPGInt2Ptr())
		})
	}
}

func TestInt16_DecodeText(t *testing.T) {
	value := NewInt16(0)

	err := value.DecodeText(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeText(nil, []byte("1"))
	assert.Nil(t, err)
}

func TestInt16_DecodeBinary(t *testing.T) {
	value := NewInt16(0)

	err := value.DecodeBinary(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeBinary(nil, []byte("00"))
	assert.Nil(t, err)
}
