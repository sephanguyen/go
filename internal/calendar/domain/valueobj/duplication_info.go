package valueobj

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type (
	DuplicationFrequency string
)

const (
	Daily  DuplicationFrequency = "daily"
	Weekly DuplicationFrequency = "weekly"
)

type DuplicationInfo struct {
	StartDate time.Time
	EndDate   time.Time
	Frequency DuplicationFrequency
}

func NewDuplicationInfo(startDate, endDate time.Time, frequency string) (*DuplicationInfo, error) {
	freq, err := GetDuplicationFrequency(frequency)
	if err != nil {
		return nil, err
	}

	duplicationInfo := &DuplicationInfo{
		StartDate: startDate,
		EndDate:   endDate,
		Frequency: freq,
	}

	return duplicationInfo, nil
}

func GetDuplicationFrequency(frequency string) (DuplicationFrequency, error) {
	switch strings.ToLower(frequency) {
	case "daily":
		return Daily, nil
	case "weekly":
		return Weekly, nil
	}

	return "", errors.New("unsupported duplication frequency")
}

func (d *DuplicationInfo) Validate() error {
	if d.StartDate.IsZero() {
		return fmt.Errorf("start date cannot be empty")
	}

	if d.EndDate.IsZero() {
		return fmt.Errorf("end date cannot be empty")
	}

	if len(d.Frequency) == 0 {
		return fmt.Errorf("frequency cannot be empty, should be daily or weekly")
	}

	if d.EndDate.Before(d.StartDate) {
		return fmt.Errorf("end date could not before start date")
	}

	return nil
}

func (d *DuplicationInfo) RetrieveDateOccurrences() []time.Time {
	dates := []time.Time{}
	days := d.EndDate.Sub(d.StartDate) / (24 * time.Hour)
	numOfDays := int(days)

	switch d.Frequency {
	case Daily:
		for i := 0; i <= numOfDays; i++ {
			nextDay := d.StartDate.AddDate(0, 0, i)
			dates = append(dates, nextDay)
		}
	case Weekly:
		for i := 0; i <= numOfDays; i += 7 {
			nextDay := d.StartDate.AddDate(0, 0, i)
			dates = append(dates, nextDay)
		}
	}

	return dates
}
