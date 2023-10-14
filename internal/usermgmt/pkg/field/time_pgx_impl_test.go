package field

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestTime_toPGTimestamptz(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		input          Time
		expectedOutput pgtype.Timestamptz
	}{
		{
			name:  "parse valid Time to pgtype.Timestamptz successfully",
			input: NewTime(now),
			expectedOutput: pgtype.Timestamptz{
				Time:   now,
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null Time to pgtype.Timestamptz successfully",
			input: NewNullTime(),
			expectedOutput: pgtype.Timestamptz{
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null Time to pgtype.Timestamptz successfully",
			input: NewUndefinedTime(),
			expectedOutput: pgtype.Timestamptz{
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, testCase.input.toPGTimestamptz())
		})
	}
}

func TestTime_DecodeText(t *testing.T) {
	now := time.Now()
	value := NewTime(time.Now())

	err := value.DecodeText(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeText(nil, []byte(now.Format("2006-01-02 15:04:05.999999999Z07")))
	assert.Nil(t, err)
}

func TestTime_DecodeBinary(t *testing.T) {
	now := time.Now()
	value := NewTime(now)

	err := value.DecodeBinary(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeBinary(nil, []byte("00000000"))
	assert.Nil(t, err)
}
