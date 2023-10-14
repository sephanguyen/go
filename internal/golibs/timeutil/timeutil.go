package timeutil

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/bob"

	"go.uber.org/multierr"
)

// StartWeek returns Monday at 00:00:00 AM UTC time.
func StartWeek() time.Time {
	now := time.Now().UTC()
	if wd := now.Weekday(); wd != time.Monday {
		if wd == time.Sunday {
			// Go weekday is calculated from Sunday to Saturday,
			// but student learning week is counted from Monday to Sunday,
			// so if today is Sunday, that means Monday is 6 days ago.
			now = now.Add(-6 * 24 * time.Hour)
		} else {
			now = now.Add(-time.Duration(int(wd-time.Monday)) * 24 * time.Hour)
		}
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UTC()
}

// StartWeekIn returns Monday at 00:00:00 AM in provided country time.
// If country is invalid, it returns the time in UTC.
func StartWeekIn(c pb.Country) time.Time {
	loc := Timezone(c)
	now := time.Now().In(loc)
	if wd := now.Weekday(); wd != time.Monday {
		if wd == time.Sunday {
			// Go weekday is calculated from Sunday to Saturday,
			// but student learning week is counted from Monday to Sunday,
			// so if today is Sunday, that means Monday is 6 days ago.
			now = now.Add(-6 * 24 * time.Hour)
		} else {
			now = now.Add(-time.Duration(int(wd-time.Monday)) * 24 * time.Hour)
		}
	}
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// EndWeek returns Sunday at 23:59:59.999 PM UTC time.
func EndWeek() time.Time {
	end := StartWeek().Add(6 * 24 * time.Hour)
	return time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999, end.Location()).UTC()
}

// EndWeek returns Sunday at 23:59:59.999 PM in provided country time.
func EndWeekIn(c pb.Country) time.Time {
	end := StartWeekIn(c).Add(6 * 24 * time.Hour)
	return time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 999, end.Location())
}

func StartDateNextWeek() time.Time {
	return StartWeek().Add(7 * 24 * time.Hour).UTC()
}

// EndOfToday returns today at 23:59:59.999 PM UTC time.
func EndOfToday() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999, now.Location()).UTC()
}

// Now can be use in code that need to be test with time comparing
var Now = func() time.Time {
	return time.Now()
}

// DefaultPSPStartDate returns the start date of all preset study plans
// of the current year by a country.
func DefaultPSPStartDate(country pb.Country) time.Time {
	var env string
	switch country {
	// The start date of a preset study plan counts at week 1 in VN.
	// Start date must be on Monday.
	case pb.COUNTRY_VN:
		env = os.Getenv("PSP_START_DATE_IN_VN")
	}

	startDate, err := time.Parse("January 02", env)
	if err != nil {
		// default start date is August 01.
		startDate = time.Date(0001, time.August, 01, 0, 0, 0, 0, time.UTC)
	}
	return time.Date(time.Now().Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
}

func PSPStartDate(country pb.Country, startDate string) time.Time {
	if startDate == "" {
		return DefaultPSPStartDate(country)
	}
	startTime, err := time.Parse("January 02", startDate)
	if err != nil {
		// default start date is August 01.
		startTime = time.Date(0001, time.August, 01, 0, 0, 0, 0, time.UTC)
	}

	return time.Date(time.Now().Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.UTC)
}

func MidnightIn(country pb.Country, t time.Time) time.Time {
	loc := Timezone(country)
	d := t.In(loc)
	return time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, loc)
}

var (
	vnTimeLoc = time.FixedZone("UTC+7", 7*60*60)
	sgTimeLoc = time.FixedZone("UTC+8", 8*60*60)
	jpTimeLoc = time.FixedZone("UTC+9", 9*60*60)
)

func Timezone(country pb.Country) *time.Location {
	switch country {
	case pb.COUNTRY_VN, pb.COUNTRY_ID:
		return vnTimeLoc
	case pb.COUNTRY_SG:
		return sgTimeLoc
	case pb.COUNTRY_JP:
		return jpTimeLoc
	default:
		return time.UTC
	}
}

func ResetTimeZone(t time.Time, timezone string) (time.Time, error) {
	local, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), local), nil
}

func Location(timezone string) *time.Location {
	if len(timezone) == 0 {
		return vnTimeLoc
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = vnTimeLoc
	}
	return loc
}

func CalculateRunTime(ctx context.Context, name string) func() {
	start := time.Now()
	return func() {
		fmt.Println(fmt.Sprintf("%s took", name), time.Since(start))
	}
}

func EqualDate(date1, date2 time.Time) bool {
	if date1.Location() != date2.Location() {
		date2 = date2.In(date1.Location())
	}

	yearDate1, monthDate1, dayDate1 := date1.Date()
	yearDate2, monthDate2, dayDate2 := date2.Date()

	if yearDate1 != yearDate2 || monthDate1 != monthDate2 || dayDate1 != dayDate2 {
		return false
	}

	return true
}

func BeforeDate(date1, date2 time.Time) bool {
	if date1.Year() < date2.Year() {
		return true
	}
	if date1.Year() > date2.Year() {
		return false
	}
	return date1.Year() == date2.Year() && date1.YearDay() < date2.YearDay()
}

func ParsingTimeFromYYYYMMDDStr(timeStr, tz string) (time.Time, error) {
	// format YYYY-MM-DD HH:MM:SS
	dateTime := strings.Split(timeStr, " ")
	d := strings.Split(dateTime[0], "-")
	t := strings.Split(dateTime[1], ":")
	var result time.Time

	year, err1 := strconv.ParseInt(d[0], 10, 32)
	month, err2 := strconv.ParseInt(d[1], 10, 32)
	day, err3 := strconv.ParseInt(d[2], 10, 32)
	hour, err4 := strconv.ParseInt(t[0], 10, 32)
	min, err5 := strconv.ParseInt(t[1], 10, 32)
	sec, err6 := strconv.ParseInt(t[2], 10, 32)

	err := multierr.Combine(err1, err2, err3, err4, err5, err6)
	if err != nil {
		return result, err
	}

	location, err := time.LoadLocation(tz)
	if err != nil {
		return result, err
	}
	result = time.Date(int(year), time.Month(int(month)), int(day), int(hour), int(min), int(sec), 0, location)
	return result, nil
}

func NormalizeToStartOfDay(date time.Time, country pb.Country) time.Time {
	d := date.In(Timezone(country))
	d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
	return d
}
