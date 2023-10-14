package dto

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"

	"github.com/stretchr/testify/assert"
)

func TestTimesheetQueryArgs_Validate(t *testing.T) {
	t.Parallel()

	var (
		tsQueryArgWithEmptyStaffIds = &TimesheetQueryArgs{
			LocationID:    "ts_1",
			TimesheetDate: time.Now(),
		}

		tsQueryArgWithEmptyLocationId = &TimesheetQueryArgs{
			StaffIDs:      []string{"staff-1", "staff-2", "staff-3"},
			LocationID:    "",
			TimesheetDate: time.Now(),
		}

		tsQueryArgWithEmptyTimesheetDate = &TimesheetQueryArgs{
			StaffIDs:   []string{"staff-1", "staff-2", "staff-3"},
			LocationID: "",
		}

		tsQueryArgValid = &TimesheetQueryArgs{
			StaffIDs:      []string{"staff-1", "staff-2", "staff-3"},
			LocationID:    "location-1",
			TimesheetDate: time.Now(),
		}
	)

	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{

		{
			name:         "validate success",
			request:      tsQueryArgValid,
			expectedResp: nil,
		},
		{
			name:         "fail with empty staff ids",
			request:      tsQueryArgWithEmptyStaffIds,
			expectedResp: fmt.Errorf("validate TimesheetQueryArgs Failed: staffIDs: %v, locationID: %s, timesheetDate: %v", tsQueryArgWithEmptyStaffIds.StaffIDs, tsQueryArgWithEmptyStaffIds.LocationID, tsQueryArgWithEmptyStaffIds.TimesheetDate),
		},
		{
			name:         "fail with empty location id",
			request:      tsQueryArgWithEmptyLocationId,
			expectedResp: fmt.Errorf("validate TimesheetQueryArgs Failed: staffIDs: %v, locationID: %s, timesheetDate: %v", tsQueryArgWithEmptyLocationId.StaffIDs, tsQueryArgWithEmptyLocationId.LocationID, tsQueryArgWithEmptyLocationId.TimesheetDate),
		},
		{
			name:         "fail with empty timesheet date",
			request:      tsQueryArgWithEmptyTimesheetDate,
			expectedResp: fmt.Errorf("validate TimesheetQueryArgs Failed: staffIDs: %v, locationID: %s, timesheetDate: %v", tsQueryArgWithEmptyTimesheetDate.StaffIDs, tsQueryArgWithEmptyTimesheetDate.LocationID, tsQueryArgWithEmptyTimesheetDate.TimesheetDate),
		},
	}
	for _, testcase := range testCases {
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*TimesheetQueryArgs).Validate()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
func TestTimesheetQueryArgs_Normalize(t *testing.T) {
	t.Parallel()

	var (
		tsQueryArgs = &TimesheetQueryArgs{
			LocationID:    "ts_1",
			TimesheetDate: time.Now(),
		}
	)

	t.Run("normalize dates success", func(t *testing.T) {
		tsQueryArgs.Normalize()
		assert.Equal(t, tsQueryArgs.TimesheetDate.Location(), timeutil.Timezone(pbc.COUNTRY_JP))
		assert.Equal(t, tsQueryArgs.TimesheetDate.Hour(), 0)
		assert.Equal(t, tsQueryArgs.TimesheetDate.Minute(), 0)
		assert.Equal(t, tsQueryArgs.TimesheetDate.Second(), 0)
		assert.Equal(t, tsQueryArgs.TimesheetDate.Nanosecond(), 0)
	})
}
