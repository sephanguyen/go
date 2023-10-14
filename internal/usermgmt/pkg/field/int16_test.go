package field

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestNewInt16(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          int16
		expectedOutput Int16
	}{
		{
			name:  "init Int16 with zero value",
			input: 0,
			expectedOutput: Int16{
				value:  0,
				status: StatusPresent,
			},
		},
		{
			name:  "init Int16 with zero value",
			input: 24,
			expectedOutput: Int16{
				value:  24,
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewInt16(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullInt16(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput Int16
	}{
		{
			name: "init null Int16",
			expectedOutput: Int16{
				value:  0,
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullInt16() == testCase.expectedOutput)
		})
	}
}

func TestInt16_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput error
	}{
		{
			name:           "happy case",
			expectedOutput: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, NewNullInt16().Ptr().UnmarshalJSON([]byte(`19990`)), testCase.expectedOutput)
		})
	}
}

func TestInt16_SetNull(t *testing.T) {
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

			data := int16(1234)
			value := NewInt16(data)
			assert.Equal(t, value.Int16(), data)

			value.SetNull()
			assert.Equal(t, value, NewNullInt16())
			assert.NotEqual(t, value.Int16(), data)
		})
	}
}

func TestInt16_UnmarshalCSV(t *testing.T) {
	wrongInt16 := "."
	_, parseErr := strconv.ParseInt(wrongInt16, 10, 16)
	parseErr = InvalidDataError{
		Method:       "strconv.ParseInt",
		FieldName:    "Int16",
		Reason:       parseErr.Error(),
		InvalidValue: wrongInt16,
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
			name: "null value",
			args: args{
				data: "",
			},
			wantErr: nil,
		},
		{
			name: "existed value is 0",
			args: args{
				data: "0",
			},
			wantErr: nil,
		},
		{
			name: "existed value is -1",
			args: args{
				data: "-1",
			},
			wantErr: nil,
		},
		{
			name: "existed invalid value",
			args: args{
				data: wrongInt16,
			},
			wantErr: parseErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewNullInt16()
			assert.Equal(t, tt.wantErr, field.UnmarshalCSV(tt.args.data))

			if tt.wantErr == nil {
				int16Data, _ := strconv.Atoi(tt.args.data)
				assert.Equal(t, int16(int16Data), field.Int16())
			}

			if tt.args.data == "" {
				assert.Equal(t, StatusNull, int(field.status))
			}
		})
	}
}

func TestInt16s_Int16s(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		input          Int16s
		expectedOutput []int16
	}

	var zeroInt16 int16
	positiveInt16 := int16(12)

	testCases := []testCase{
		{
			name:           "nil slice",
			input:          nil,
			expectedOutput: []int16{},
		},
		{
			name:           "empty slice",
			input:          Int16s{},
			expectedOutput: []int16{},
		},
		{
			name: "there are values in slice",
			input: Int16s{
				NewInt16(zeroInt16),
				NewInt16(positiveInt16),
				NewNullInt16(),
			},
			expectedOutput: []int16{zeroInt16, positiveInt16, 0},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, len(testCase.expectedOutput), len(testCase.input.Int16s()))
		assert.Equal(t, testCase.expectedOutput, testCase.input.Int16s())
	}
}
