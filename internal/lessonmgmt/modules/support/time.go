package support

import "time"

// The number of seconds elapsed end of date January 1, 1970 UTC
const UnixToEnd = 86400

func StartOfDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDate(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, t.Nanosecond(), t.Location())
}

func ConvertToStringRFCFormat(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
