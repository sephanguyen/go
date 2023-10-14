package dto

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	pbc "github.com/manabie-com/backend/pkg/genproto/bob"
)

type TimesheetQueryArgs struct {
	StaffIDs      []string
	LocationID    string
	TimesheetDate time.Time
}

func (t *TimesheetQueryArgs) Validate() error {
	if len(t.StaffIDs) == 0 || t.LocationID == "" || t.TimesheetDate.IsZero() {
		return fmt.Errorf("validate TimesheetQueryArgs Failed: staffIDs: %v, locationID: %s, timesheetDate: %v", t.StaffIDs, t.LocationID, t.TimesheetDate)
	}
	return nil
}

func (t *TimesheetQueryArgs) Normalize() {
	t.TimesheetDate = t.TimesheetDate.In(timeutil.Timezone(pbc.COUNTRY_JP))
	t.TimesheetDate = time.Date(t.TimesheetDate.Year(), t.TimesheetDate.Month(), t.TimesheetDate.Day(), 0, 0, 0, 0, t.TimesheetDate.Location())
}
