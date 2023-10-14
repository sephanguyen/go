package utils

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"

	"github.com/stretchr/testify/assert"
)

func TestSplitWithCapacity(t *testing.T) {
	testCases := []struct {
		name, s, sep string
		cap          int
		want         []string
	}{
		{
			name: "normal input with larger capacity than length",
			s:    "a,b,c",
			sep:  ",",
			cap:  5,
			want: []string{"a", "b", "c", "", ""},
		},
		{
			name: "normal input with smaller capacity than length",
			s:    "a b c",
			sep:  " ",
			cap:  0,
			want: []string{"a", "b", "c"},
		},
		{
			name: "empty input should return slice of empty strings",
			s:    "",
			sep:  " ",
			cap:  2,
			want: []string{"", ""},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := SplitWithCapacity(tc.s, tc.sep, tc.cap)
			if !equalSlices(got, tc.want) {
				t.Errorf("got %v; want %v", got, tc.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTruncateTimeToStartOfDay(t *testing.T) {
	t.Run("return correct date time value", func(t *testing.T) {
		randomDate, _ := time.Parse("2006/01/02", "2023/04/02")
		date := TruncateTimeToStartOfDay(randomDate)
		assert.Equal(t, 2023, date.Year())
		assert.Equal(t, "April", date.Month().String())
		assert.Equal(t, 02, date.Day())
		assert.Equal(t, 0, date.Hour())
		assert.Equal(t, 0, date.Minute())
		assert.Equal(t, 0, date.Second())
	})
	t.Run("return correct date time value", func(t *testing.T) {
		randomDate, _ := time.Parse("2006 Jan 02 15:04:05", "2023 Apr 02 12:15:30.918273645")
		date := TruncateTimeToStartOfDay(randomDate)
		assert.Equal(t, 2023, date.Year())
		assert.Equal(t, "April", date.Month().String())
		assert.Equal(t, 02, date.Day())
		assert.Equal(t, 0, date.Hour())
		assert.Equal(t, 0, date.Minute())
		assert.Equal(t, 0, date.Second())
	})
}

func TestCompareStringsRegardlessOrder(t *testing.T) {
	type testCase struct {
		name                 string
		userAccessPathsInput []string
		locationIDs          []string
		expectedErr          error
	}

	var (
		location1 = "location-1"
		location2 = "location-2"
		location3 = "location-3"
		location4 = "location-4"
		location5 = "location-5"
	)

	testCases := []testCase{
		{
			name:                 "both slices are empty",
			userAccessPathsInput: nil,
			locationIDs:          nil,
			expectedErr:          nil,
		},
		{
			name:                 "values of two slices are equal regardless order, length of them are 1",
			userAccessPathsInput: []string{location1},
			locationIDs:          []string{location1},
			expectedErr:          nil,
		},
		{
			name: "values of two slices are equal, same order, length of them are 2",
			userAccessPathsInput: []string{
				location1,
				location2,
			},
			locationIDs: []string{
				location1,
				location2,
			},
			expectedErr: nil,
		},
		{
			name: "values of two slices are equal, different order, length of them are 2",
			userAccessPathsInput: []string{
				location1,
				location2,
			},
			locationIDs: []string{
				location2,
				location1,
			},
			expectedErr: nil,
		},
		{
			name: "values of two slices are equal, different order, length of them are 4",
			userAccessPathsInput: []string{
				location1,
				location2,
				location3,
				location4,
			},
			locationIDs: []string{
				location4,
				location2,
				location1,
				location3,
			},
			expectedErr: nil,
		},
		{
			name: "values of two slices are not equal equal, length of them are 2",
			userAccessPathsInput: []string{
				location1,
				location2,
			},
			locationIDs: []string{
				location1,
				location3,
			},
			expectedErr: fmt.Errorf(`can not find "%s" of first slice in second slice: %s`, location2, []string{location1, location3}),
		},
		{
			name: "values of two slices are not equal equal, length of them are different",
			userAccessPathsInput: []string{
				location1,
				location2,
				location3,
				location4,
			},
			locationIDs: []string{
				location2,
				location1,
				location5,
				location3,
				location4,
			},
			expectedErr: fmt.Errorf("length of two string slices are not equal, length of first slice is %v but second slice is %v", 4, 5),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualErr := CompareStringsRegardlessOrder(testCase.userAccessPathsInput, testCase.locationIDs)

			assert.Equal(t, testCase.expectedErr, actualErr)
		})
	}
}

func TestInArrayInt(t *testing.T) {
	t.Parallel()
	intArr := []int{1, 2, 3}
	t.Run("int in array", func(t *testing.T) {
		t.Parallel()
		result := InArrayInt(1, intArr)
		assert.Exactly(t, true, result)
	})

	t.Run("int not in array", func(t *testing.T) {
		t.Parallel()
		result := InArrayInt(4, intArr)
		assert.Exactly(t, false, result)
	})
}

func TestMaxTime(t *testing.T) {
	t.Parallel()

	type args struct {
		times []time.Time
	}
	tests := []struct {
		name string
		args args
		want time.Time
	}{
		{
			name: "empty slice",
			args: args{times: []time.Time{}},
			want: time.Time{},
		},
		{
			name: "single element",
			args: args{times: []time.Time{time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}},
			want: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			name: "multiple elements",
			args: args{times: []time.Time{
				time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
				time.Date(2019, 1, 1, 0, 0, 0, 0, time.UTC),
			}},
			want: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, MaxTime(tt.args.times), "MaxTime(%v)", tt.args.times)
		})
	}
}

func TestIsFutureDate(t *testing.T) {
	t.Parallel()

	type args struct {
		startTime field.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "past date",
			args: args{startTime: field.NewTime(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))},
			want: false,
		},
		{
			name: "future date",
			args: args{startTime: field.NewTime(time.Date(3024, 1, 1, 0, 0, 0, 0, time.UTC))},
			want: true,
		},
		{
			name: "current date",
			args: args{startTime: field.NewTime(time.Now())},
			want: false,
		},
		{
			name: "different location",
			args: args{startTime: field.NewTime(time.Date(3024, 1, 1, 0, 0, 0, 0, time.FixedZone("UTC+7", 7*60*60)))},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsFutureDate(tt.args.startTime), "IsFutureDate(%v)", tt.args.startTime)
		})
	}
}
