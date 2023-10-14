package valueobj

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
)

func TestDuplicationInfo_GetDuplicationFrequency(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		freq         string
		expectedFreq DuplicationFrequency
		hasError     bool
	}{
		{
			name:         "valid duplication frequency - daily",
			freq:         "daily",
			expectedFreq: Daily,
			hasError:     false,
		},
		{
			name:         "valid duplication frequency - weekly",
			freq:         "weekly",
			expectedFreq: Weekly,
			hasError:     false,
		},
		{
			name:         "valid duplication frequency but with caps",
			freq:         "DaIlY",
			expectedFreq: Daily,
			hasError:     false,
		},
		{
			name:         "valid duplication frequency but with space",
			freq:         "weekly ",
			expectedFreq: "",
			hasError:     true,
		},
		{
			name:         "invalid duplication frequency",
			freq:         "hello",
			expectedFreq: "",
			hasError:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			duplicationFrequency, err := GetDuplicationFrequency(tc.freq)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedFreq, duplicationFrequency)
		})
	}
}

func TestDuplicationInfo_Validate(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name     string
		dupInfo  *DuplicationInfo
		hasError bool
	}{
		{
			name: "duplication info is valid",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
				Frequency: Daily,
			},
			hasError: false,
		},
		{
			name: "start date is empty",
			dupInfo: &DuplicationInfo{
				EndDate:   time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
				Frequency: Daily,
			},
			hasError: true,
		},
		{
			name: "end date is empty",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				Frequency: Daily,
			},
			hasError: true,
		},
		{
			name: "frequency is empty",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
			},
			hasError: true,
		},
		{
			name: "start date greater than end date",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 10, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				Frequency: Daily,
			},
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dupInfo.Validate()
			if err != nil {
				require.True(t, tc.hasError)
			} else {
				require.False(t, tc.hasError)
			}
		})
	}
}

func TestDuplicationInfo_RetrieveDateOccurrences(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name          string
		dupInfo       *DuplicationInfo
		expectedDates []time.Time
	}{
		{
			name: "duplication info generates expected daily dates",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 9, 23, 0, 0, 0, 0, time.UTC),
				Frequency: Daily,
			},
			expectedDates: []time.Time{
				time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 9, 20, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 9, 21, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 9, 22, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 9, 23, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "duplication info generates daily dates with same start and end date",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				Frequency: Daily,
			},
			expectedDates: []time.Time{
				time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "duplication info generates weekly dates with same start and end date",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				Frequency: Weekly,
			},
			expectedDates: []time.Time{
				time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "duplication info generates expected weekly dates with last date same as end date",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
				Frequency: Weekly,
			},
			expectedDates: []time.Time{
				time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 9, 26, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 03, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "duplication info generates expected weekly dates with end date 6 days after last date",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 10, 16, 0, 0, 0, 0, time.UTC),
				Frequency: Weekly,
			},
			expectedDates: []time.Time{
				time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 9, 26, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 03, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "duplication info generates expected weekly dates with end date 1 day after last date",
			dupInfo: &DuplicationInfo{
				StartDate: time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2022, 10, 18, 0, 0, 0, 0, time.UTC),
				Frequency: Weekly,
			},
			expectedDates: []time.Time{
				time.Date(2022, 9, 19, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 9, 26, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 03, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC),
				time.Date(2022, 10, 17, 0, 0, 0, 0, time.UTC),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			isEqual := slices.Equal(tc.dupInfo.RetrieveDateOccurrences(), tc.expectedDates)
			require.True(t, isEqual)
		})
	}
}
