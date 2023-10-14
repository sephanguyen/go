package field

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/stretchr/testify/assert"
)

func TestNewDate(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		input          time.Time
		expectedOutput Date
	}{
		{
			name:  "init Date with zero value",
			input: time.Time{},
			expectedOutput: Date{
				value:  time.Time{},
				status: StatusPresent,
			},
		},
		{
			name:  "init Date with valid value",
			input: now,
			expectedOutput: Date{
				value:  now,
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewDate(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullDate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput Date
	}{
		{
			name: "init null Date",
			expectedOutput: Date{
				value:  time.Time{},
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullDate() == testCase.expectedOutput)
		})
	}
}

func TestDate_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput error
		input          interface{}
	}{
		{
			name:           "happy case",
			expectedOutput: nil,
			input:          `"1999/01/12"`,
		},
		{
			name:           "bad case: invalid format",
			expectedOutput: fmt.Errorf(`field Date invalid: parsing time "1999-01-12" as "2006/01/02": cannot parse "-01-12" as "/", invalid value: 1999-01-12`),
			input:          `"1999-01-12"`,
		},
		{
			name:           "bad case: invalid string",
			expectedOutput: fmt.Errorf(`field Date invalid: json: cannot unmarshal number into Go value of type string, invalid value: 1`),
			input:          `1`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			err := NewNullDate().Ptr().UnmarshalJSON([]byte(testCase.input.(string)))
			if err != nil {
				assert.Equal(t, testCase.expectedOutput.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedOutput, nil)
			}

		})
	}
}

func TestDate_SetNull(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
	}{
		{
			name: "happy case",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			now := time.Now()
			value := NewDate(now)
			assert.Equal(t, value.Date(), now)

			value.SetNull()
			assert.Equal(t, value, NewNullDate())
			assert.NotEqual(t, value.Date(), now)
		})
	}
}

func TestDate_UnmarshalCSV(t *testing.T) {
	wrongDate := "."
	_, parseErr := time.Parse(constant.DateLayout, wrongDate)
	parseErr = InvalidDataError{
		Method:       "time.Parse",
		FieldName:    "Date",
		Reason:       parseErr.Error(),
		InvalidValue: wrongDate,
	}

	type args struct {
		data string
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "existed value is 2006/01/02",
			args: args{
				data: "2006/01/02",
			},
			wantErr: nil,
		},
		{
			name: "null value",
			args: args{
				data: "",
			},
			wantErr: nil,
		},
		{
			name: "existed invalid value",
			args: args{
				data: wrongDate,
			},
			wantErr: parseErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewNullDate()
			assert.Equal(t, tt.wantErr, field.UnmarshalCSV(tt.args.data))

			if tt.wantErr == nil {
				date, _ := time.Parse(constant.DateLayout, tt.args.data)
				assert.Equal(t, date, field.Date())
			}

			if tt.args.data == "" {
				assert.Equal(t, StatusNull, int(field.status))
			}
		})
	}
}

func TestDates_Dates(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		input          Dates
		expectedOutput []time.Time
	}

	emptyTime := time.Time{}
	now := time.Now()

	testCases := []testCase{
		{
			name:           "nil slice",
			input:          nil,
			expectedOutput: []time.Time{},
		},
		{
			name:           "empty slice",
			input:          Dates{},
			expectedOutput: []time.Time{},
		},
		{
			name: "there are values in slice",
			input: Dates{
				NewDate(emptyTime),
				NewDate(now),
				NewNullDate(),
			},
			expectedOutput: []time.Time{emptyTime, now, emptyTime},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, len(testCase.expectedOutput), len(testCase.input.Dates()))
		assert.Equal(t, testCase.expectedOutput, testCase.input.Dates())
	}
}
