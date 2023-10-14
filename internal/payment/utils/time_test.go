package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEndOfDate(t *testing.T) {
	for i := 0; i < 24; i++ {
		timeTest := time.Unix(1673283600, 0).Add(time.Duration(i) * time.Hour).UTC()
		timeConvert := EndOfDate(timeTest, 7)
		yearOfEndDate, monthOfEndDate, dateOfEndDate := timeConvert.Date()
		require.Equal(t, 2023, yearOfEndDate)
		require.Equal(t, time.Month(1), monthOfEndDate)
		require.Equal(t, 10, dateOfEndDate)
		require.Equal(t, 16, timeConvert.Hour())
		require.Equal(t, 59, timeConvert.Minute())
		require.Equal(t, 59, timeConvert.Second())
	}
}

func TestStartOfDate(t *testing.T) {
	for i := 0; i < 24; i++ {
		timeTest := time.Unix(1673283600, 0).Add(time.Duration(i) * time.Hour).UTC()
		timeConvert := StartOfDate(timeTest, 7)
		yearOfEndDate, monthOfEndDate, dateOfEndDate := timeConvert.Date()
		require.Equal(t, 2023, yearOfEndDate)
		require.Equal(t, time.Month(1), monthOfEndDate)
		require.Equal(t, 9, dateOfEndDate)
		require.Equal(t, 17, timeConvert.Hour())
		require.Equal(t, 0, timeConvert.Minute())
		require.Equal(t, 0, timeConvert.Second())
	}
}
