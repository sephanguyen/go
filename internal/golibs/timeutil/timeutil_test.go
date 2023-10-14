package timeutil_test

import (
	"os"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
	"gotest.tools/assert"
)

func TestDefaultPSPStartDate_EnvIsSet(t *testing.T) {
	t.Parallel()
	old := os.Getenv("PSP_START_DATE_IN_VN")
	os.Setenv("PSP_START_DATE_IN_VN", "August 05")
	defer func() {
		os.Setenv("PSP_START_DATE_IN_VN", old)
	}()

	startDate := timeutil.DefaultPSPStartDate(pb.COUNTRY_VN)
	expected := time.Date(time.Now().Year(), time.August, 05, 0, 0, 0, 0, time.UTC)
	if !startDate.Equal(expected) {
		t.Fatalf("timeutil.DefaultPSPStartDate: got: %v, want: %v", startDate, expected)
	}
}

func TestDefaultPSPStartDate_EnvIsNotSet(t *testing.T) {
	t.Parallel()
	startDate := timeutil.DefaultPSPStartDate(pb.COUNTRY_NONE)
	expected := time.Date(time.Now().Year(), time.August, 01, 0, 0, 0, 0, time.UTC)
	if !startDate.Equal(expected) {
		t.Fatalf("timeutil.DefaultPSPStartDate: got: %v, want: %v", startDate, expected)
	}
}

func TestPSPStartDate(t *testing.T) {
	t.Parallel()
	startDate := timeutil.PSPStartDate(pb.COUNTRY_VN, "August 12")
	expected := time.Date(time.Now().Year(), time.August, 12, 0, 0, 0, 0, time.UTC)
	if !startDate.Equal(expected) {
		t.Fatalf("timeutil.DefaultPSPStartDate: got: %v, want: %v", startDate, expected)
	}
}

func TestPSPStartDateError(t *testing.T) {
	t.Parallel()
	startDate := timeutil.PSPStartDate(pb.COUNTRY_VN, "Aug 12")
	expected := time.Date(time.Now().Year(), time.August, 01, 0, 0, 0, 0, time.UTC)
	if !startDate.Equal(expected) {
		t.Fatalf("timeutil.DefaultPSPStartDate: got: %v, want: %v", startDate, expected)
	}
}

func TestMidnightIn(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		country  pb.Country
		t        string
		expected string
	}{
		{
			country:  pb.COUNTRY_VN,
			t:        "2019-12-10T10:11:12Z",
			expected: "2019-12-10T00:00:00+07:00",
		},
		{
			country:  pb.COUNTRY_VN,
			t:        "2019-12-10T19:11:12Z",
			expected: "2019-12-11T00:00:00+07:00",
		},
		{
			country:  pb.COUNTRY_VN,
			t:        "2019-12-10T17:00:00Z",
			expected: "2019-12-11T00:00:00+07:00",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()
			ti, _ := time.Parse(time.RFC3339, tc.t)
			got := timeutil.MidnightIn(tc.country, ti)
			expected, _ := time.Parse(time.RFC3339, tc.expected)
			if !got.Equal(expected) {
				t.Errorf("MidnightIn(%v, %v) = %v, expected: %v", tc.country, ti, got, expected)
			}
		})
	}
}

func TestEqualDate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	testCases := []struct {
		arg1     time.Time
		arg2     time.Time
		expected bool
	}{
		{
			arg1:     now,
			arg2:     now,
			expected: true,
		},
		{
			arg1:     now,
			arg2:     now.Add(time.Hour * 30),
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			result := timeutil.EqualDate(tc.arg1, tc.arg2)
			if result != tc.expected {
				t.Errorf("EqualDate(%v, %v) = %v, expected: %v", tc.arg1, tc.arg2, result, tc.expected)
			}
		})
	}
}
func TestParsingTimeFromYYYYMMDDStr(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		timezone string
		timeStr  string
		expected string
	}{
		{
			timezone: "Asia/Ho_Chi_Minh",
			timeStr:  "2019-01-01 07:30:00",
			expected: "2019-01-01T07:30:00+07:00",
		},
		{
			timezone: "UTC",
			timeStr:  "2019-12-11 19:11:12",
			expected: "2019-12-11T19:11:12+00:00",
		},
		{
			timezone: "Asia/Tokyo",
			timeStr:  "2019-12-11 17:00:00",
			expected: "2019-12-11T17:00:00+09:00",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			got, _ := timeutil.ParsingTimeFromYYYYMMDDStr(tc.timeStr, tc.timezone)
			expected, _ := time.Parse(time.RFC3339, tc.expected)
			if !got.Equal(expected) {
				t.Errorf("ParsingTimeFromYYYYMMDDStr(%v, %v) = %v, expected: %v", tc.timeStr, tc.timezone, got, expected)
			}
		})
	}
}

func TestBeforeDate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		time1    time.Time
		time2    time.Time
		expected bool
	}{
		{
			time1:    time.Now().AddDate(-1, 0, 0),
			time2:    time.Now(),
			expected: true,
		},
		{
			time1:    time.Now(),
			time2:    time.Now().AddDate(-1, 0, 0),
			expected: false,
		},
		{
			time1:    time.Now().AddDate(0, -1, 0),
			time2:    time.Now(),
			expected: true,
		},
		{
			time1:    time.Now(),
			time2:    time.Now().AddDate(0, -1, 0),
			expected: false,
		},
		{
			time1:    time.Now().AddDate(0, 0, -1),
			time2:    time.Now(),
			expected: true,
		},
		{
			time1:    time.Now(),
			time2:    time.Now().AddDate(0, 0, -1),
			expected: false,
		},
		{
			time1:    time.Now(),
			time2:    time.Now(),
			expected: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run("", func(t *testing.T) {
			t.Parallel()

			result := timeutil.BeforeDate(tc.time1, tc.time2)
			assert.Equal(t, tc.expected, result)

		})
	}
}

func TestNormalizeToStartOfDay(t *testing.T) {
	t.Parallel()

	t.Run("should convert correct when time is 00:00 in JP timezone", func(t *testing.T) {
		date, _ := time.Parse(time.DateTime, "2023-08-31 15:00:00")
		normalizedDate := timeutil.NormalizeToStartOfDay(date, pb.COUNTRY_JP)
		assert.Equal(t, normalizedDate.Day(), 1)
		assert.Equal(t, normalizedDate.Month().String(), "September")
		assert.Equal(t, normalizedDate.Year(), 2023)
		assert.Equal(t, normalizedDate.Hour(), 0)
		assert.Equal(t, normalizedDate.Location().String(), "UTC+9")
	})
	t.Run("should convert correct when time is 09:00 in JP timezone", func(t *testing.T) {
		date, _ := time.Parse(time.DateTime, "2023-09-01 00:00:00")
		normalizedDate := timeutil.NormalizeToStartOfDay(date, pb.COUNTRY_JP)
		assert.Equal(t, normalizedDate.Day(), 1)
		assert.Equal(t, normalizedDate.Month().String(), "September")
		assert.Equal(t, normalizedDate.Year(), 2023)
		assert.Equal(t, normalizedDate.Hour(), 0)
		assert.Equal(t, normalizedDate.Location().String(), "UTC+9")
	})
	t.Run("should convert correct when timezone is invalid", func(t *testing.T) {
		date, _ := time.Parse(time.DateTime, "2023-09-01 00:00:00")
		normalizedDate := timeutil.NormalizeToStartOfDay(date, 10)
		assert.Equal(t, normalizedDate.Day(), 1)
		assert.Equal(t, normalizedDate.Month().String(), "September")
		assert.Equal(t, normalizedDate.Year(), 2023)
		assert.Equal(t, normalizedDate.Hour(), 0)
		assert.Equal(t, normalizedDate.Location().String(), "UTC")
	})
}
