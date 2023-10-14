package payloads

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"

	"github.com/stretchr/testify/require"
)

func TestFetchDateInfoByDateRangeRequest_Validate(t *testing.T) {
	t.Parallel()
	startDate := time.Date(2022, 10, 01, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2022, 10, 10, 0, 0, 0, 0, time.UTC)
	locationID := idutil.ULIDNow()
	timezone := "sample-timezone"

	testCases := []struct {
		name     string
		req      *FetchDateInfoByDateRangeRequest
		hasError bool
	}{
		{
			name:     "request empty",
			req:      &FetchDateInfoByDateRangeRequest{},
			hasError: true,
		},
		{
			name: "request start date empty",
			req: &FetchDateInfoByDateRangeRequest{
				EndDate:    endDate,
				LocationID: locationID,
				Timezone:   timezone,
			},
			hasError: true,
		},
		{
			name: "request end date empty",
			req: &FetchDateInfoByDateRangeRequest{
				StartDate:  startDate,
				LocationID: locationID,
				Timezone:   timezone,
			},
			hasError: true,
		},
		{
			name: "request location id empty",
			req: &FetchDateInfoByDateRangeRequest{
				StartDate: startDate,
				EndDate:   endDate,
				Timezone:  timezone,
			},
			hasError: true,
		},
		{
			name: "start date is greater than the end date",
			req: &FetchDateInfoByDateRangeRequest{
				StartDate:  endDate,
				EndDate:    startDate,
				LocationID: locationID,
				Timezone:   timezone,
			},
			hasError: true,
		},
		{
			name: "valid request without timezone",
			req: &FetchDateInfoByDateRangeRequest{
				StartDate:  startDate,
				EndDate:    endDate,
				LocationID: locationID,
			},
			hasError: false,
		},
		{
			name: "valid request",
			req: &FetchDateInfoByDateRangeRequest{
				StartDate:  startDate,
				EndDate:    endDate,
				LocationID: locationID,
				Timezone:   timezone,
			},
			hasError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.req.Validate()

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
