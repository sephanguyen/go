package utils

import (
	"fmt"
	"time"
)

const (
	CountryJp = "COUNTRY_JP"
	CountryVn = "COUNTRY_VN"
)

var countryTZMap = map[string]string{
	CountryJp: "Asia/Tokyo",
	CountryVn: "Asia/Ho_Chi_Minh",
}

func GetTimeLocationByCountry(country string) (*time.Location, error) {
	if country == "" {
		country = CountryVn
	}

	timezone, ok := countryTZMap[country]
	if !ok {
		return nil, fmt.Errorf("invalid organization country")
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, err
	}

	return location, nil
}

func ResetTimeComponent(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func GetTimeInLocation(t time.Time, country string) (time.Time, error) {
	location, err := GetTimeLocationByCountry(country)
	if err != nil {
		return time.Time{}, err
	}

	return t.In(location), nil
}
