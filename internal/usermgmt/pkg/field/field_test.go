package field

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsUndefined(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Field
		expectedOutput bool
	}{
		{
			name: "check Boolean is undefined",
			input: Boolean{
				value:  false,
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check Boolean is undefined",
			input: Boolean{
				value:  true,
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check Date is undefined",
			input: Date{
				value:  time.Time{},
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check Date is undefined",
			input: Date{
				value:  time.Now(),
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check Int32 is undefined",
			input: Int32{
				value:  0,
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check Int32 is undefined",
			input: Int32{
				value:  24,
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check String is undefined",
			input: String{
				value:  "",
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check String is undefined",
			input: String{
				value:  "example",
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check Time is undefined",
			input: Time{
				value:  time.Time{},
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check Time is undefined",
			input: Time{
				value:  time.Now(),
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check TimeWithoutTz is undefined",
			input: TimeWithoutTz{
				value:  time.Time{},
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
		{
			name: "check TimeWithoutTz is undefined",
			input: TimeWithoutTz{
				value:  time.Now(),
				status: StatusUndefined,
			},
			expectedOutput: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, IsUndefined(testCase.input), testCase.expectedOutput)
		})
	}
}

func TestIsNull(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Field
		expectedOutput bool
	}{
		{
			name: "check Boolean is undefined",
			input: Boolean{
				value:  false,
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check Boolean is undefined",
			input: Boolean{
				value:  true,
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check Date is undefined",
			input: Date{
				value:  time.Time{},
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check Date is undefined",
			input: Date{
				value:  time.Now(),
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check Int32 is undefined",
			input: Int32{
				value:  0,
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check Int32 is undefined",
			input: Int32{
				value:  24,
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check String is undefined",
			input: String{
				value:  "",
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check String is undefined",
			input: String{
				value:  "example",
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check Time is undefined",
			input: Time{
				value:  time.Time{},
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check Time is undefined",
			input: Time{
				value:  time.Now(),
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check TimeWithoutTz is undefined",
			input: TimeWithoutTz{
				value:  time.Time{},
				status: StatusNull,
			},
			expectedOutput: true,
		},
		{
			name: "check TimeWithoutTz is undefined",
			input: TimeWithoutTz{
				value:  time.Now(),
				status: StatusNull,
			},
			expectedOutput: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, IsNull(testCase.input), testCase.expectedOutput)
		})
	}
}

func TestIsNil(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          func() Field
		expectedOutput bool
	}{
		{
			name: "check nil for valued Boolean ptr",
			input: func() Field {
				boolVar := NewBoolean(true)
				return &boolVar
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil Boolean ptr",
			input: func() Field {
				var boolPtr *Boolean
				return boolPtr
			},
			expectedOutput: true,
		},
		{
			name: "check nil for valued String ptr",
			input: func() Field {
				stringVar := NewString("string")
				return &stringVar
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil String ptr",
			input: func() Field {
				var stringPtr *String
				return stringPtr
			},
			expectedOutput: true,
		},
		{
			name: "check nil for valued Int16 ptr",
			input: func() Field {
				int16Var := NewInt16(16)
				return &int16Var
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil Int16 ptr",
			input: func() Field {
				var int16Ptr *Int16
				return int16Ptr
			},
			expectedOutput: true,
		},
		{
			name: "check nil for valued Int32 ptr",
			input: func() Field {
				int32Var := NewInt16(32)
				return &int32Var
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil Int32 ptr",
			input: func() Field {
				var int32Ptr *Int32
				return int32Ptr
			},
			expectedOutput: true,
		},
		{
			name: "check nil for valued Int64 ptr",
			input: func() Field {
				int64Var := NewInt16(64)
				return &int64Var
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil Int64 ptr",
			input: func() Field {
				var int64Ptr *Int64
				return int64Ptr
			},
			expectedOutput: true,
		},
		{
			name: "check nil for valued Date ptr",
			input: func() Field {
				dateVar := NewDate(time.Now())
				return &dateVar
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil Date ptr",
			input: func() Field {
				var datePtr *Date
				return datePtr
			},
			expectedOutput: true,
		},
		{
			name: "check nil for valued Time ptr",
			input: func() Field {
				timeVar := NewTime(time.Now())
				return &timeVar
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil Time ptr",
			input: func() Field {
				var timePtr *Time
				return timePtr
			},
			expectedOutput: true,
		},
		{
			name: "check nil for valued TimeWithoutTz ptr",
			input: func() Field {
				timeWithoutTzVar := NewTimeWithoutTz(time.Now())
				return &timeWithoutTzVar
			},
			expectedOutput: false,
		},
		{
			name: "check nil for nil TimeWithoutTz ptr",
			input: func() Field {
				var timeWithoutTzPtr *TimeWithoutTz
				return timeWithoutTzPtr
			},
			expectedOutput: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, IsNil(testCase.input()), testCase.expectedOutput)
		})
	}
}

func TestIsPresent(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		input          Field
		expectedOutput bool
	}{
		{
			name: "check Boolean is undefined",
			input: Boolean{
				value:  false,
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check Boolean is undefined",
			input: Boolean{
				value:  true,
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check Date is undefined",
			input: Date{
				value:  time.Time{},
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check Date is undefined",
			input: Date{
				value:  time.Now(),
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check Int32 is undefined",
			input: Int32{
				value:  0,
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check Int32 is undefined",
			input: Int32{
				value:  24,
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check String is undefined",
			input: String{
				value:  "",
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check String is undefined",
			input: String{
				value:  "example",
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check Time is undefined",
			input: Time{
				value:  time.Time{},
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check Time is undefined",
			input: Time{
				value:  time.Now(),
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check TimeWithoutTz is undefined",
			input: TimeWithoutTz{
				value:  time.Time{},
				status: StatusPresent,
			},
			expectedOutput: true,
		},
		{
			name: "check TimeWithoutTz is undefined",
			input: TimeWithoutTz{
				value:  time.Now(),
				status: StatusPresent,
			},
			expectedOutput: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(t, IsPresent(testCase.input), testCase.expectedOutput)
		})
	}
}
