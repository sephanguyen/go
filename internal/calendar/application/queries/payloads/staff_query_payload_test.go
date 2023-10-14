package payloads

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetStaffRequest_Validate(t *testing.T) {
	t.Parallel()
	locationID := "location-id1"

	testCases := []struct {
		name     string
		req      *GetStaffRequest
		hasError bool
	}{
		{
			name: "request full fields",
			req: &GetStaffRequest{
				LocationID: locationID,
			},
			hasError: false,
		},
		{
			name:     "request location id empty",
			req:      &GetStaffRequest{},
			hasError: true,
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
func TestGetStaffByLocationIDsAndNameOrEmailRequest_Validate(t *testing.T) {
	t.Parallel()
	locationID := []string{"location-id1", "location-id2"}
	keyword := "name"

	testCases := []struct {
		name     string
		req      *GetStaffByLocationIDsAndNameOrEmailRequest
		hasError bool
	}{
		{
			name: "request full fields",
			req: &GetStaffByLocationIDsAndNameOrEmailRequest{
				LocationIDs: locationID,
				Keyword:     keyword,
			},
			hasError: false,
		},
		{
			name: "request keyword is empty",
			req: &GetStaffByLocationIDsAndNameOrEmailRequest{
				LocationIDs: locationID,
			},
			hasError: false,
		},
		{
			name: "request location ids is empty",
			req: &GetStaffByLocationIDsAndNameOrEmailRequest{
				Keyword: keyword,
			},
			hasError: true,
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
