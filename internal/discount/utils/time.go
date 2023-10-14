package utils

import (
	"time"

	"github.com/manabie-com/backend/internal/discount/constant"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

const TimeOut = 15 * time.Second

func IsZeroTime(t time.Time) bool {
	return t.Unix() == 0 || t.IsZero()
}

func IsSameDate(t1, t2 time.Time) bool {
	dateFormat := "2006-01-02 00:00:00"
	return t1.Format(dateFormat) == t2.Format(dateFormat)
}

// get date format and truncate time component
func DateFormatWithNoTimestamp(timeStr string) (time.Time, error) {
	dateFormattedParse, err := time.Parse(constant.DateFormatYYYYMMDD, timeStr)
	if err != nil {
		return time.Time{}, status.Error(codes.Internal, err.Error())
	}
	// Truncate the time component to get the date only
	return dateFormattedParse.UTC().Truncate(constant.DayDuration), nil
}
