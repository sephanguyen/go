package utils

import (
	"time"
)

const (
	TimeFormat = "2006-01-02 15:04:00 -0700"
	TimeOut    = 15 * time.Second
)

func StartOfDate(t time.Time, timezone int32) time.Time {
	if timezone == 0 {
		return t.Truncate(24 * time.Hour)
	}
	timeAfterTruncated := t.Truncate(time.Duration(24) * time.Hour)

	if t.Hour() >= int(24-timezone) {
		return timeAfterTruncated.Add(time.Duration(24-timezone) * time.Hour)
	}

	return timeAfterTruncated.Add(time.Duration(-timezone) * time.Hour)
}

func EndOfDate(t time.Time, timezone int32) time.Time {
	nextDate := StartOfDate(t, timezone).AddDate(0, 0, 1)
	endOfDate := nextDate.Add(-1 * time.Second)
	return endOfDate
}

func TimeNow(timezone int32) time.Time {
	return StartOfDate(time.Now().UTC(), timezone)
}

// DaysIn return number of days of a month, a year
func DaysIn(month time.Month, year int) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func ConvertToLocalTime(utcTime time.Time, timezone int32) time.Time {
	loc := time.FixedZone("", int(timezone)*60*60)
	localTime := utcTime.In(loc)
	return time.Date(localTime.Year(), localTime.Month(), localTime.Day(), localTime.Hour(), localTime.Minute(), localTime.Second(), localTime.Nanosecond(), loc)
}

func StartOfLocalDate(utcTime time.Time, timezone int32) time.Time {
	localTime := ConvertToLocalTime(utcTime, timezone)
	return StartOfDate(localTime, timezone)
}

// CheckTimeRangeOverlap reports whether the time range of t1 and t2 is overlap.
func CheckTimeRangeOverlap(t1 TimeRange, t2 TimeRange) bool {
	if !t1.FromTime.After(t2.ToTime) && !t2.FromTime.After(t1.ToTime) {
		return true
	}
	return false
}

// BeforeRangeTime reports whether the time t is before time range timeRange.
func BeforeRangeTime(time time.Time, timeRange TimeRange) bool {
	return time.Before(timeRange.FromTime)
}

// AfterRangeTime reports whether the time t is after time range timeRange.
func AfterRangeTime(time time.Time, timeRange TimeRange) bool {
	return time.After(timeRange.ToTime)
}

// WithinRangeTime reports whether the time t is within time range timeRange.
func WithinRangeTime(time time.Time, timeRange TimeRange) bool {
	if !time.Before(timeRange.FromTime) && !time.After(timeRange.ToTime) {
		return true
	}
	return false
}
