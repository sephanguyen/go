package common

import (
	"fmt"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/timesheet/domain/dto"
	tpb "github.com/manabie-com/backend/pkg/manabuf/timesheet/v1"

	"github.com/stretchr/testify/assert"
)

func TestTimesheet_MergeListTimesheet(t *testing.T) {
	t.Parallel()
	now := time.Now()
	timesheets1 := []*dto.Timesheet{
		{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson1_1",
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson1_2",
				},
			},
			IsCreated: true,
		},
		{
			ID:                       "timesheet1_2",
			StaffID:                  "staff1_1",
			LocationID:               "location1_2",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet1_3",
			StaffID:                  "staff1_2",
			LocationID:               "location1_3",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet1_4",
			StaffID:                  "staff1_4",
			LocationID:               "location1_4",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
	}
	timesheets2 := []*dto.Timesheet{
		{
			ID:                       "timesheet2_1",
			StaffID:                  "staff2_1",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet2_2",
			StaffID:                  "staff2_2",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					LessonID: "lesson2_1",
				},
				{
					LessonID: "lesson2_2",
				},
			},
			IsCreated: true,
		},
	}
	timesheets3 := []*dto.Timesheet{}
	timesheets4 := []*dto.Timesheet{
		{
			ID:                       "timesheet2_1",
			StaffID:                  "staff2_1",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet2_2",
			StaffID:                  "staff2_2",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:              "",
			StaffID:         "staff1_4",
			LocationID:      "location1_4",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					LessonID: "lesson2_4",
				},
				{
					LessonID: "lesson2_5",
				},
			},
			IsCreated: false,
		},
	}
	timesheets5 := []*dto.Timesheet{
		{
			ID:                       "timesheet2_1",
			StaffID:                  "staff2_1",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_SUBMITTED.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
	}
	expectTimesheets1 := []*dto.Timesheet{
		{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson1_1",
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson1_2",
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson2_1",
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson2_2",
				},
			},
			IsCreated: true,
		},
		{
			ID:                       "timesheet1_2",
			StaffID:                  "staff1_1",
			LocationID:               "location1_2",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet1_3",
			StaffID:                  "staff1_2",
			LocationID:               "location1_3",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet2_1",
			StaffID:                  "staff2_1",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet2_2",
			StaffID:                  "staff2_2",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet1_4",
			StaffID:                  "staff1_4",
			LocationID:               "location1_4",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
	}
	expectTimesheets2 := []*dto.Timesheet{
		{
			ID:              "timesheet1_1",
			StaffID:         "staff1_1",
			LocationID:      "location1_1",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson1_1",
				},
				{
					TimesheetID: "timesheet1_1",
					LessonID:    "lesson1_2",
				},
			},
			IsCreated: true,
		},
		{
			ID:                       "timesheet1_2",
			StaffID:                  "staff1_1",
			LocationID:               "location1_2",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet1_3",
			StaffID:                  "staff1_2",
			LocationID:               "location1_3",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet2_1",
			StaffID:                  "staff2_1",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:                       "timesheet2_2",
			StaffID:                  "staff2_2",
			LocationID:               "location1",
			TimesheetStatus:          tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:            now,
			ListTimesheetLessonHours: nil,
			IsCreated:                true,
		},
		{
			ID:              "timesheet1_4",
			StaffID:         "staff1_4",
			LocationID:      "location1_4",
			TimesheetStatus: tpb.TimesheetStatus_TIMESHEET_STATUS_DRAFT.String(),
			TimesheetDate:   now,
			ListTimesheetLessonHours: []*dto.TimesheetLessonHours{
				{
					TimesheetID: "timesheet1_4",
					LessonID:    "lesson2_4",
				},
				{
					TimesheetID: "timesheet1_4",
					LessonID:    "lesson2_5",
				},
			},
			IsCreated: true,
		},
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
		expectedErr  interface{}
	}{
		{
			name:         "merge success",
			request:      [][]*dto.Timesheet{timesheets1, timesheets2},
			expectedResp: expectTimesheets1,
			expectedErr:  nil,
		},
		{
			name:         "merge success with one empty timesheetID",
			request:      [][]*dto.Timesheet{timesheets1, timesheets4},
			expectedResp: expectTimesheets2,
			expectedErr:  nil,
		},
		{
			name:         "merge success with one slice is empty",
			request:      [][]*dto.Timesheet{timesheets1, timesheets3},
			expectedResp: timesheets1,
			expectedErr:  nil,
		},
		{
			name:         "merge failed with when different timesheet info",
			request:      [][]*dto.Timesheet{timesheets2, timesheets5},
			expectedResp: ([]*dto.Timesheet)(nil),
			expectedErr:  fmt.Errorf("validateMerge failed"),
		},
	}
	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			resp, err := MergeListTimesheet(testcase.request.([][]*dto.Timesheet)[0], testcase.request.([][]*dto.Timesheet)[1])
			assert.Equal(t, testcase.expectedErr, err)
			assert.True(t, CompareListTimesheet(testcase.expectedResp.([]*dto.Timesheet), resp))
		})
	}
}
