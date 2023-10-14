package common

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/timeutil"
	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	pb "github.com/manabie-com/backend/pkg/genproto/bob"
)

func GetTimesheetUniqueKey(teacherID string, locationID string, timesheetDate time.Time) string {
	return fmt.Sprintf("%s_%s_%s", teacherID, locationID, timesheetDate.Format("2006-01-02"))
}

// MergeListTimesheet will loop two list and merge which elements which have same timesheet general info
func MergeListTimesheet(timesheets1, timesheets2 []*dto.Timesheet) ([]*dto.Timesheet, error) {
	if len(timesheets1) == 0 {
		return timesheets2, nil
	}
	if len(timesheets2) == 0 {
		return timesheets1, nil
	}

	var (
		listTimesheet = make([]*dto.Timesheet, len(timesheets1))
		mapTimesheet  = make(map[string]*dto.Timesheet, 0)
	)

	copy(listTimesheet, timesheets1)

	listTimesheet = append(listTimesheet, timesheets2...)
	if len(listTimesheet) == 0 {
		return nil, nil
	}

	for _, e := range listTimesheet {
		normalizedDate := timeutil.NormalizeToStartOfDay(e.TimesheetDate, pb.COUNTRY_JP)
		key := GetTimesheetUniqueKey(e.StaffID, e.LocationID, normalizedDate)
		value, found := mapTimesheet[key]
		if !found {
			mapTimesheet[key] = e
		} else {
			merged, err := e.Merge(value)
			if err != nil {
				return nil, err
			}
			mapTimesheet[key] = merged
		}
	}
	mergedListTimesheet := make([]*dto.Timesheet, 0, len(mapTimesheet))
	for _, value := range mapTimesheet {
		mergedListTimesheet = append(mergedListTimesheet, value)
	}
	return mergedListTimesheet, nil
}

// CompareListTimesheet return true if each timesheet in timesheets1 equal with other timesheet in timesheets2 two
// not compare the order of timesheet in timesheets1 and timesheets2
func CompareListTimesheet(timesheets1, timesheets2 []*dto.Timesheet) bool {
	if len(timesheets1) == 0 && len(timesheets2) == 0 {
		return true
	}

	if len(timesheets1) != len(timesheets2) {
		return false
	}

	mapTimesheets1 := map[string]*dto.Timesheet{}
	for _, e := range timesheets1 {
		mapTimesheets1[e.ID] = e
	}

	for _, e := range timesheets2 {
		if !e.IsEqual(mapTimesheets1[e.ID]) {
			return false
		}
	}
	return true
}
