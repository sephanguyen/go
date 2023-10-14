package field

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInt32(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          int32
		expectedOutput Int32
	}{
		{
			name:  "init Int32 with zero value",
			input: 0,
			expectedOutput: Int32{
				value:  0,
				status: StatusPresent,
			},
		},
		{
			name:  "init Int32 with zero value",
			input: 24,
			expectedOutput: Int32{
				value:  24,
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewInt32(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullInt32(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput Int32
	}{
		{
			name: "init null Int32",
			expectedOutput: Int32{
				value:  0,
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullInt32() == testCase.expectedOutput)
		})
	}
}

func TestInt32_UnmarshalJSON(t *testing.T) {
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
			assert.Equal(t, NewNullInt32().Ptr().UnmarshalJSON([]byte(`19990112`)), testCase.expectedOutput)
		})
	}
}

func TestInt32_SetNull(t *testing.T) {
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

			data := int32(1234)
			value := NewInt32(data)
			assert.Equal(t, value.Int32(), data)

			value.SetNull()
			assert.Equal(t, value, NewNullInt32())
			assert.NotEqual(t, value.Int32(), data)
		})
	}
}

func TestInt32_UnmarshalCSV(t *testing.T) {
	wrongInt32 := "."
	_, parseErr := strconv.ParseInt(wrongInt32, 10, 32)
	parseErr = InvalidDataError{
		Method:       "strconv.ParseInt",
		FieldName:    "Int32",
		Reason:       parseErr.Error(),
		InvalidValue: wrongInt32,
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
				data: wrongInt32,
			},
			wantErr: parseErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewNullInt32()
			assert.Equal(t, tt.wantErr, field.UnmarshalCSV(tt.args.data))

			if tt.wantErr == nil {
				int32Data, _ := strconv.Atoi(tt.args.data)
				assert.Equal(t, int32(int32Data), field.Int32())
			}

			if tt.args.data == "" {
				assert.Equal(t, StatusNull, int(field.status))
			}
		})
	}
}

func TestInt32s_Int32s(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		input          Int32s
		expectedOutput []int32
	}

	var zeroInt32 int32
	positiveInt32 := int32(12)

	testCases := []testCase{
		{
			name:           "nil slice",
			input:          nil,
			expectedOutput: []int32{},
		},
		{
			name:           "empty slice",
			input:          Int32s{},
			expectedOutput: []int32{},
		},
		{
			name: "there are values in slice",
			input: Int32s{
				NewInt32(zeroInt32),
				NewInt32(positiveInt32),
				NewNullInt32(),
			},
			expectedOutput: []int32{zeroInt32, positiveInt32, 0},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, len(testCase.expectedOutput), len(testCase.input.Int32s()))
		assert.Equal(t, testCase.expectedOutput, testCase.input.Int32s())
	}
}
