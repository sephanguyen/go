package field

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestTimeWithoutTz_toPGTimestamp(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		input          TimeWithoutTz
		expectedOutput pgtype.Timestamp
	}{
		{
			name:  "parse valid TimeWithoutTz to pgtype.Timestamp successfully",
			input: NewTimeWithoutTz(now.UTC()),
			expectedOutput: pgtype.Timestamp{
				Time:   now.UTC(),
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null TimeWithoutTz to pgtype.Timestamp successfully",
			input: NewNullTimeWithoutTz(),
			expectedOutput: pgtype.Timestamp{
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null TimeWithoutTz to pgtype.Timestamp successfully",
			input: NewUndefinedTimeWithoutTz(),
			expectedOutput: pgtype.Timestamp{
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, testCase.input.toPGTimestamp())
		})
	}
}

func TestTimeWithoutTz_DecodeText(t *testing.T) {
	now := time.Now()
	value := NewTimeWithoutTz(time.Now())

	err := value.DecodeText(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeText(nil, []byte(now.Format("2006-01-02 15:04:05.999999999")))
	assert.Nil(t, err)
}

func TestTimeWithoutTz_DecodeBinary(t *testing.T) {
	now := time.Now()
	value := NewTimeWithoutTz(now)

	err := value.DecodeBinary(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeBinary(nil, []byte("00000000"))
	assert.Nil(t, err)
}
