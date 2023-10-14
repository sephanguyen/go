package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStudentSubscriptionAccessPaths_IsValid(t *testing.T) {
	t.Parallel()
	now := time.Now()
	tcs := []struct {
		name    string
		input   StudentSubscriptionAccessPaths
		isValid bool
	}{
		{
			name: "full fields",
			input: StudentSubscriptionAccessPaths{
				{
					SubscriptionID: "id-1",
					LocationID:     "location-id-1",
					CreatedAt:      now,
					UpdatedAt:      now,
					DeletedAt:      now,
				},
				{
					SubscriptionID: "id-2",
					LocationID:     "location-id-2",
					CreatedAt:      now,
					UpdatedAt:      now,
					DeletedAt:      now,
				},
			},
			isValid: true,
		},
		{
			name: "missing location id",
			input: StudentSubscriptionAccessPaths{
				{
					SubscriptionID: "id-1",
					LocationID:     "location-id-1",
					CreatedAt:      now,
					UpdatedAt:      now,
					DeletedAt:      now,
				},
				{
					SubscriptionID: "id-2",
					CreatedAt:      now,
					UpdatedAt:      now,
					DeletedAt:      now,
				},
			},
			isValid: false,
		},
		{
			name: "missing subscription id",
			input: StudentSubscriptionAccessPaths{
				{
					LocationID: "location-id-1",
					CreatedAt:  now,
					UpdatedAt:  now,
					DeletedAt:  now,
				},
				{
					SubscriptionID: "id-2",
					LocationID:     "location-id-2",
					CreatedAt:      now,
					UpdatedAt:      now,
					DeletedAt:      now,
				},
			},
			isValid: false,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.IsValid()
			if tc.isValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
