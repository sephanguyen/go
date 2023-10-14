package entities

import (
	"testing"

	"github.com/manabie-com/backend/internal/calendar/domain/constants"

	"github.com/stretchr/testify/require"
)

func TestDateType_GetDateTypeID(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name               string
		dateTypeID         string
		expectedDateTypeID constants.DateTypeID
		hasError           bool
	}{
		{
			name:               "valid date type ID - regular",
			dateTypeID:         "regular",
			expectedDateTypeID: constants.RegularDay,
			hasError:           false,
		},
		{
			name:               "valid date type ID - seasonal",
			dateTypeID:         "seasonal",
			expectedDateTypeID: constants.SeasonalDay,
			hasError:           false,
		},
		{
			name:               "valid date type ID - spare",
			dateTypeID:         "spare",
			expectedDateTypeID: constants.SpareDay,
			hasError:           false,
		},
		{
			name:               "valid date type ID - closed",
			dateTypeID:         "closed",
			expectedDateTypeID: constants.ClosedDay,
			hasError:           false,
		},
		{
			name:               "valid date type ID but with caps",
			dateTypeID:         "reGuLaR",
			expectedDateTypeID: constants.RegularDay,
			hasError:           false,
		},
		{
			name:               "valid date type ID but with space",
			dateTypeID:         "closed ",
			expectedDateTypeID: "",
			hasError:           true,
		},
		{
			name:               "invalid date type ID",
			dateTypeID:         "hello",
			expectedDateTypeID: "",
			hasError:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dateTypeID, err := GetDateTypeID(tc.dateTypeID)

			if tc.hasError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tc.expectedDateTypeID, dateTypeID)
		})
	}
}
