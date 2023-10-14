package dto

import (
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"
)

type ListTimesheetLessonHours []*TimesheetLessonHours

type TimesheetLessonHours struct {
	TimesheetID string
	LessonID    string
	FlagOn      bool
	IsCreated   bool
	IsDeleted   bool
}

func (t *TimesheetLessonHours) ToEntity() *entity.TimesheetLessonHours {
	return &entity.TimesheetLessonHours{
		TimesheetID: database.Text(t.TimesheetID),
		LessonID:    database.Text(t.LessonID),
		FlagOn:      database.Bool(t.FlagOn),
	}
}

func (t *TimesheetLessonHours) IsEqual(timesheetLessonHours2 *TimesheetLessonHours) bool {
	if t == nil && timesheetLessonHours2 == nil {
		return true
	}
	if t == nil || timesheetLessonHours2 == nil {
		return false
	}
	if t.TimesheetID == timesheetLessonHours2.TimesheetID &&
		t.LessonID == timesheetLessonHours2.LessonID &&
		t.IsCreated == timesheetLessonHours2.IsCreated &&
		t.FlagOn == timesheetLessonHours2.FlagOn {
		return true
	}
	return false
}

func NewTimesheetLessonHoursFromEntity(entity *entity.TimesheetLessonHours) *TimesheetLessonHours {
	isCreated := false

	if !entity.CreatedAt.Time.IsZero() {
		isCreated = true
	}

	return &TimesheetLessonHours{
		TimesheetID: entity.TimesheetID.String,
		LessonID:    entity.LessonID.String,
		FlagOn:      entity.FlagOn.Bool,
		IsCreated:   isCreated,
	}
}

func (l ListTimesheetLessonHours) UpdateTimesheetID(timesheetID string) {
	for i := range l {
		l[i].TimesheetID = timesheetID
	}
}

func (l ListTimesheetLessonHours) ToEntities() []*entity.TimesheetLessonHours {
	timesheetLessonHoursEntities := make([]*entity.TimesheetLessonHours, 0, len(l))
	for _, e := range l {
		timesheetLessonHoursEntities = append(timesheetLessonHoursEntities, e.ToEntity())
	}
	return timesheetLessonHoursEntities
}

func getTimesheetLessonHoursUniqueKey(timesheetLessonHours *TimesheetLessonHours) string {
	return fmt.Sprintf("%s_%s", timesheetLessonHours.TimesheetID, timesheetLessonHours.LessonID)
}

func (l ListTimesheetLessonHours) IsEqual(listTimesheetHours2 ListTimesheetLessonHours) bool {
	if len(l) != len(listTimesheetHours2) {
		return false
	}
	if len(l) == 0 && len(listTimesheetHours2) == 0 {
		return true
	}
	mapTimesheetLessonHours := make(map[string]*TimesheetLessonHours, len(l))
	for _, e := range l {
		mapTimesheetLessonHours[getTimesheetLessonHoursUniqueKey(e)] = e
	}

	for _, e := range listTimesheetHours2 {
		if !e.IsEqual(mapTimesheetLessonHours[getTimesheetLessonHoursUniqueKey(e)]) {
			return false
		}
	}
	return true
}

func (l ListTimesheetLessonHours) Merge(listTimesheetLessonHours2 ListTimesheetLessonHours) ListTimesheetLessonHours {
	mapTimesheetLessonHours := make(map[string]*TimesheetLessonHours, len(l))
	formatter := "%s_%s"

	for _, e := range l {
		key := fmt.Sprintf(formatter, e.TimesheetID, e.LessonID)
		mapTimesheetLessonHours[key] = e
	}

	for _, e := range listTimesheetLessonHours2 {
		key := fmt.Sprintf(formatter, e.TimesheetID, e.LessonID)
		_, found := mapTimesheetLessonHours[key]
		if !found {
			mapTimesheetLessonHours[key] = e
		}
	}

	mergedListTimesheetLessonHours := make(ListTimesheetLessonHours, 0)
	for _, value := range mapTimesheetLessonHours {
		mergedListTimesheetLessonHours = append(mergedListTimesheetLessonHours, value)
	}
	return mergedListTimesheetLessonHours
}
