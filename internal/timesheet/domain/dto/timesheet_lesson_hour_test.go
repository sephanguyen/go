package dto

import (
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/timesheet/domain/entity"

	"github.com/stretchr/testify/assert"
)

func TestNewTimesheetLessonHoursFromEntity(t *testing.T) {
	t.Parallel()

	var (
		timesheetLessonHours = &TimesheetLessonHours{
			TimesheetID: "timesheetID_1",
			LessonID:    "lessonID_1",
			FlagOn:      true,
		}
		timesheetLessonHoursEntity = &entity.TimesheetLessonHours{
			TimesheetID: database.Text("timesheetID_1"),
			LessonID:    database.Text("lessonID_1"),
			FlagOn:      database.Bool(true),
		}
		timesheetLessonHoursWithCreatedFlag = &TimesheetLessonHours{
			TimesheetID: "timesheetID_1",
			LessonID:    "lessonID_1",
			FlagOn:      true,
			IsCreated:   true,
		}
		timesheetLessonHoursWithCreatedDateEntity = &entity.TimesheetLessonHours{
			TimesheetID: database.Text("timesheetID_1"),
			LessonID:    database.Text("lessonID_1"),
			FlagOn:      database.Bool(true),
			CreatedAt:   database.Timestamptz(time.Now()),
		}
	)
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "new timesheet lesson hours from entity success",
			request:      timesheetLessonHoursEntity,
			expectedResp: timesheetLessonHours,
		},
		{
			name:         "new timesheet lesson hours from entity set is created",
			request:      timesheetLessonHoursWithCreatedDateEntity,
			expectedResp: timesheetLessonHoursWithCreatedFlag,
		},
	}
	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			resp := NewTimesheetLessonHoursFromEntity(testcase.request.(*entity.TimesheetLessonHours))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheetLessonHours_ToEntity(t *testing.T) {
	t.Parallel()
	timesheetLessonHours := &TimesheetLessonHours{
		TimesheetID: "timesheetID_1",
		LessonID:    "lessonID_1",
		FlagOn:      true,
	}
	timesheetLessonHoursEntity := &entity.TimesheetLessonHours{
		TimesheetID: database.Text("timesheetID_1"),
		LessonID:    database.Text("lessonID_1"),
		FlagOn:      database.Bool(true),
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "to entity success",
			request:      timesheetLessonHours,
			expectedResp: timesheetLessonHoursEntity,
		},
	}
	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(*TimesheetLessonHours).ToEntity()
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}

func TestTimesheetLessonHours_IsEqual(t *testing.T) {
	t.Parallel()
	timesheetLessonHours1 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_1",
		LessonID:    "lessonID_1",
		IsCreated:   true,
	}
	timesheetLessonHours2 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_2",
		LessonID:    "lessonID_2",
		IsCreated:   false,
	}
	timesheetLessonHours3 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_1",
		LessonID:    "lessonID_1",
		IsCreated:   true,
	}
	testCases := []struct {
		name         string
		request      interface{}
		expectedResp interface{}
	}{
		{
			name:         "compare success with true",
			request:      []*TimesheetLessonHours{timesheetLessonHours1, timesheetLessonHours3},
			expectedResp: true,
		},
		{
			name:         "compare success with false",
			request:      []*TimesheetLessonHours{timesheetLessonHours1, timesheetLessonHours2},
			expectedResp: false,
		},
	}
	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.([]*TimesheetLessonHours)[0].IsEqual(testcase.request.([]*TimesheetLessonHours)[1])
			assert.Equal(t, testcase.expectedResp.(bool), resp)
		})
	}
}

func TestListTimesheetLessonHours_IsEqual(t *testing.T) {
	t.Parallel()
	timesheetLessonHours1 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_1",
		LessonID:    "lessonID_1",
		IsCreated:   true,
	}
	timesheetLessonHours2 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_2",
		LessonID:    "lessonID_2",
		IsCreated:   false,
	}
	timesheetLessonHours3 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_1",
		LessonID:    "lessonID_1",
		IsCreated:   true,
	}
	testCases := []struct {
		name         string
		request      interface{}
		toCompare    interface{}
		expectedResp interface{}
	}{
		{
			name:         "compare success with true",
			request:      ListTimesheetLessonHours{timesheetLessonHours1, timesheetLessonHours3},
			toCompare:    ListTimesheetLessonHours{timesheetLessonHours3, timesheetLessonHours1},
			expectedResp: true,
		},
		{
			name:         "both lists are empty",
			request:      ListTimesheetLessonHours{},
			toCompare:    ListTimesheetLessonHours{},
			expectedResp: true,
		},
		{
			name:         "compare success with false",
			request:      ListTimesheetLessonHours{timesheetLessonHours1, timesheetLessonHours1},
			toCompare:    ListTimesheetLessonHours{timesheetLessonHours2, timesheetLessonHours3},
			expectedResp: false,
		},
		{
			name:         "length not equal",
			request:      ListTimesheetLessonHours{timesheetLessonHours1, timesheetLessonHours3},
			toCompare:    ListTimesheetLessonHours{timesheetLessonHours3},
			expectedResp: false,
		},
	}
	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(ListTimesheetLessonHours).IsEqual(testcase.toCompare.(ListTimesheetLessonHours))
			assert.Equal(t, testcase.expectedResp.(bool), resp)
		})
	}
}

func TestListTimesheetLessonHours_Merge(t *testing.T) {
	t.Parallel()
	timesheetLessonHours1 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_1",
		LessonID:    "lessonID_1",
		IsCreated:   true,
	}
	timesheetLessonHours2 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_2",
		LessonID:    "lessonID_2",
		IsCreated:   false,
	}
	timesheetLessonHours3 := &TimesheetLessonHours{
		TimesheetID: "timesheetID_3",
		LessonID:    "lessonID_3",
		IsCreated:   true,
	}
	testCases := []struct {
		name         string
		request      interface{}
		toMerge      interface{}
		toCompare    interface{}
		expectedResp interface{}
	}{
		{
			name:         "merge timesheet lesson hours success",
			request:      ListTimesheetLessonHours{timesheetLessonHours1, timesheetLessonHours3},
			toMerge:      ListTimesheetLessonHours{timesheetLessonHours2},
			toCompare:    ListTimesheetLessonHours{timesheetLessonHours1, timesheetLessonHours2, timesheetLessonHours3},
			expectedResp: true,
		},
	}
	for _, testCase := range testCases {
		testcase := testCase
		t.Run(testcase.name, func(t *testing.T) {
			resp := testcase.request.(ListTimesheetLessonHours).Merge(testcase.toMerge.(ListTimesheetLessonHours)).IsEqual(testcase.toCompare.(ListTimesheetLessonHours))
			assert.Equal(t, testcase.expectedResp, resp)
		})
	}
}
