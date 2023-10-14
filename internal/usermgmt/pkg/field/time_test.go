package field

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"

	"github.com/stretchr/testify/assert"
)

func TestNewTime(t *testing.T) {
	t.Parallel()

	now := time.Now()

	testCases := []struct {
		name           string
		input          time.Time
		expectedOutput Time
	}{
		{
			name:  "init Time with zero value",
			input: time.Time{},
			expectedOutput: Time{
				value:  time.Time{},
				status: StatusPresent,
			},
		},
		{
			name:  "init Time with valid value",
			input: now,
			expectedOutput: Time{
				value:  now,
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewTime(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullTime(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput Time
	}{
		{
			name: "init null Time",
			expectedOutput: Time{
				value:  time.Time{},
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullTime() == testCase.expectedOutput)
		})
	}
}

func TestTime_UnmarshalJSON(t *testing.T) {
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
			expectedOutput: fmt.Errorf(`field Time invalid: parsing time "1999-01-12" as "2006/01/02": cannot parse "-01-12" as "/", invalid value: 1999-01-12`),
			input:          `"1999-01-12"`,
		},
		{
			name:           "bad case: invalid string",
			expectedOutput: fmt.Errorf(`field Time invalid: json: cannot unmarshal number into Go value of type string, invalid value: 1`),
			input:          `1`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Log(testCase.name)

			err := NewNullTime().Ptr().UnmarshalJSON([]byte(testCase.input.(string)))
			if err != nil {
				assert.Equal(t, testCase.expectedOutput.Error(), err.Error())
			} else {
				assert.Equal(t, testCase.expectedOutput, nil)
			}

		})
	}
}

func TestTime_SetNull(t *testing.T) {
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
			value := NewTime(now)
			assert.Equal(t, value.Time(), now)

			value.SetNull()
			assert.Equal(t, value, NewNullTime())
			assert.NotEqual(t, value.Time(), now)
		})
	}
}

func TestTime_UnmarshalCSV(t *testing.T) {
	wrongTime := "."
	_, parseErr := time.Parse(constant.DateLayout, wrongTime)
	parseErr = InvalidDataError{
		Method:       "time.Parse",
		FieldName:    "Time",
		Reason:       parseErr.Error(),
		InvalidValue: wrongTime,
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
				data: wrongTime,
			},
			wantErr: parseErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewNullTime()
			assert.Equal(t, tt.wantErr, field.UnmarshalCSV(tt.args.data))

			if tt.wantErr == nil {
				date, _ := time.Parse(constant.DateLayout, tt.args.data)
				assert.Equal(t, date, field.Time())
			}

			if tt.args.data == "" {
				assert.Equal(t, StatusNull, int(field.status))
			}
		})
	}
}

func TestTimes_Times(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		input          Times
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
			input:          Times{},
			expectedOutput: []time.Time{},
		},
		{
			name: "there are values in slice",
			input: Times{
				NewTime(emptyTime),
				NewTime(now),
				NewNullTime(),
			},
			expectedOutput: []time.Time{emptyTime, now, emptyTime},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, len(testCase.expectedOutput), len(testCase.input.Times()))
		assert.Equal(t, testCase.expectedOutput, testCase.input.Times())
	}
}
