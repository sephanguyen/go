package field

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          int64
		expectedOutput Int64
	}{
		{
			name:  "init Int64 with zero value",
			input: 0,
			expectedOutput: Int64{
				value:  0,
				status: StatusPresent,
			},
		},
		{
			name:  "init Int64 with zero value",
			input: 24,
			expectedOutput: Int64{
				value:  24,
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewInt64(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullInt64(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput Int64
	}{
		{
			name: "init null Int64",
			expectedOutput: Int64{
				value:  0,
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullInt64() == testCase.expectedOutput)
		})
	}
}

func TestInt64_UnmarshalJSON(t *testing.T) {
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
			assert.Equal(t, NewNullInt64().Ptr().UnmarshalJSON([]byte(`19990112`)), testCase.expectedOutput)
		})
	}
}

func TestInt64_SetNull(t *testing.T) {
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

			data := int64(1234)
			value := NewInt64(data)
			assert.Equal(t, value.Int64(), data)

			value.SetNull()
			assert.Equal(t, value, NewNullInt64())
			assert.NotEqual(t, value.Int64(), data)
		})
	}
}

func TestInt64_UnmarshalCSV(t *testing.T) {
	wrongInt64 := "."
	_, parseErr := strconv.ParseInt(wrongInt64, 10, 64)
	parseErr = InvalidDataError{
		Method:       "strconv.ParseInt",
		FieldName:    "Int64",
		Reason:       parseErr.Error(),
		InvalidValue: wrongInt64,
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
				data: wrongInt64,
			},
			wantErr: parseErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewNullInt64()
			assert.Equal(t, tt.wantErr, field.UnmarshalCSV(tt.args.data))

			if tt.wantErr == nil {
				int64Data, _ := strconv.Atoi(tt.args.data)
				assert.Equal(t, int64(int64Data), field.Int64())
			}

			if tt.args.data == "" {
				assert.Equal(t, StatusNull, int(field.status))
			}
		})
	}
}

func TestInt64s_Int64s(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		input          Int64s
		expectedOutput []int64
	}

	var zeroInt64 int64
	positiveInt64 := int64(12)

	testCases := []testCase{
		{
			name:           "nil slice",
			input:          nil,
			expectedOutput: []int64{},
		},
		{
			name:           "empty slice",
			input:          Int64s{},
			expectedOutput: []int64{},
		},
		{
			name: "there are values in slice",
			input: Int64s{
				NewInt64(zeroInt64),
				NewInt64(positiveInt64),
				NewNullInt64(),
			},
			expectedOutput: []int64{zeroInt64, positiveInt64, 0},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, len(testCase.expectedOutput), len(testCase.input.Int64s()))
		assert.Equal(t, testCase.expectedOutput, testCase.input.Int64s())
	}
}
