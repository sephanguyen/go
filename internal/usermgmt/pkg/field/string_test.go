package field

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          string
		expectedOutput String
	}{
		{
			name:  "init String with zero value",
			input: "",
			expectedOutput: String{
				value:  "",
				status: StatusPresent,
			},
		},
		{
			name:  "init String with valid value",
			input: "example",
			expectedOutput: String{
				value:  "example",
				status: StatusPresent,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewString(testCase.input) == testCase.expectedOutput)
		})
	}
}

func TestNewNullString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		expectedOutput String
	}{
		{
			name: "init null String",
			expectedOutput: String{
				value:  "",
				status: StatusNull,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.True(t, NewNullString() == testCase.expectedOutput)
		})
	}
}

func TestString_UnmarshalJSON(t *testing.T) {
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
			assert.Equal(t, NewNullString().Ptr().UnmarshalJSON([]byte(`"data"`)), testCase.expectedOutput)
		})
	}
}

func TestString_SetNull(t *testing.T) {
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

			data := "custom string"
			value := NewString(data)
			assert.Equal(t, value.String(), data)

			value.SetNull()
			assert.Equal(t, value, NewNullString())
			assert.NotEqual(t, value.String(), data)
		})
	}
}

func TestString_UnmarshalCSV(t *testing.T) {
	type args struct {
		data string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "empty value",
			args: args{
				data: "",
			},
			wantErr: nil,
		},
		// {
		// 	name: "value includes space",
		// 	args: args{
		// 		data: " ",
		// 	},
		// 	wantErr: nil,
		// },
		{
			name: "existed value",
			args: args{
				data: "data",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := NewNullString()
			assert.Equal(t, tt.wantErr, field.UnmarshalCSV(tt.args.data))

			if tt.wantErr == nil {
				assert.Equal(t, tt.args.data, field.String())
			}

			if tt.args.data == "" {
				assert.Equal(t, StatusNull, int(field.status))
			}
		})
	}
}

func TestStrings_Strings(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name           string
		input          Strings
		expectedOutput []string
	}

	testCases := []testCase{
		{
			name:           "nil slice",
			input:          nil,
			expectedOutput: []string{},
		},
		{
			name:           "empty slice",
			input:          Strings{},
			expectedOutput: []string{},
		},
		{
			name: "there are values in slice",
			input: Strings{
				NewString("string1"),
				NewString("string2"),
				NewString(""),
				NewNullString(),
			},
			expectedOutput: []string{"string1", "string2", "", ""},
		},
	}

	for _, testCase := range testCases {
		assert.Equal(t, len(testCase.expectedOutput), len(testCase.input.Strings()))
		assert.Equal(t, testCase.expectedOutput, testCase.input.Strings())
	}
}

func TestString_TrimSpace(t *testing.T) {
	tests := []struct {
		name string
		str  String
		want String
	}{
		{
			name: "trim space for undefined string",
			str:  NewUndefinedString(),
			want: NewUndefinedString(),
		},
		{
			name: "trim space for null string",
			str:  NewNullString(),
			want: NewNullString(),
		},
		{
			name: "with spaces for head and tail",
			str:  NewString(" New.String "),
			want: NewString("New.String"),
		},
		{
			name: "with a space",
			str:  NewString(" "),
			want: NewString(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.str.TrimSpace(), "ToLower()")
		})
	}

}

func TestString_ToLower(t *testing.T) {
	tests := []struct {
		name string
		str  String
		want String
	}{
		{
			name: "to lower case for undefined string",
			str:  NewUndefinedString(),
			want: NewUndefinedString(),
		},
		{
			name: "to lower case for null string",
			str:  NewNullString(),
			want: NewNullString(),
		},
		{
			name: "to lower case for string 1",
			str:  NewString("New.String"),
			want: NewString("new.string"),
		},
		{
			name: "to lower case for string 2",
			str:  NewString("new.string"),
			want: NewString("new.string"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.str.ToLower(), "ToLower()")
		})
	}

}
