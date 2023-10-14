package field

import (
	"testing"
	"time"

	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestDate_toPGDate(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		input          Date
		expectedOutput pgtype.Date
	}{
		{
			name:  "parse valid Date to pgtype.Date successfully",
			input: NewDate(now),
			expectedOutput: pgtype.Date{
				Time:   now,
				Status: pgtype.Present,
			},
		},
		{
			name:  "parse null Date to pgtype.Date successfully",
			input: NewNullDate(),
			expectedOutput: pgtype.Date{
				Status: pgtype.Null,
			},
		},
		{
			name:  "parse null Date to pgtype.Date successfully",
			input: NewUndefinedDate(),
			expectedOutput: pgtype.Date{
				Status: pgtype.Undefined,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, testCase.expectedOutput, *testCase.input.toPGDatePtr())
		})
	}
}

func TestDate_DecodeText(t *testing.T) {
	now := time.Now()
	value := NewDate(time.Now())

	err := value.DecodeText(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeText(nil, []byte(now.Format("2006-01-02")))
	assert.Nil(t, err)
}

func TestDate_DecodeBinary(t *testing.T) {
	now := time.Now()
	value := NewDate(now)

	err := value.DecodeBinary(nil, []byte{})
	assert.NotNil(t, err)

	err = value.DecodeBinary(nil, []byte("0000"))
	assert.Nil(t, err)
}
