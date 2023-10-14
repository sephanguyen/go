package payloads

import (
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lrd "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson_report/domain"
)

type (
	CalendarView string
)

const (
	Daily   CalendarView = "DAILY"
	Weekly  CalendarView = "WEEKLY"
	Monthly CalendarView = "MONTHLY"
)

type GetLessonByID struct {
	LessonID string
}

type GetLessonListArg struct {
	TeacherIDs               []string
	StudentIDs               []string
	CourseIDs                []string
	LocationIDs              []string
	ClassIDs                 []string
	Grades                   []int32
	FromDate                 time.Time
	ToDate                   time.Time
	TimeZone                 string
	Dow                      []domain.DateOfWeek // date_of_weeks
	FromTime                 string
	ToTime                   string
	GradesV2                 []string
	Limit                    uint32
	KeyWord                  string
	LessonTime               string
	CurrentTime              time.Time
	Compare                  string
	SchoolID                 string
	LessonID                 string
	LessonSchedulingStatuses []domain.LessonSchedulingStatus
	CourseTypesIDs           []string
	LessonReportStatus       []lrd.LessonReportStatus

	// to look also in given_name column in table if not using user basic info
	SearchInGivenNameColumn bool
}

type GetLessonListOnCalendarArgs struct {
	View                                CalendarView
	FromDate                            time.Time
	ToDate                              time.Time
	LocationID                          string
	Timezone                            string
	StudentIDs                          []string
	CourseIDs                           []string
	TeacherIDs                          []string
	ClassIDs                            []string
	IsIncludeNoneAssignedTeacherLessons bool
}

type GetLessonsByLocationStatusAndDateTimeRangeArgs struct {
	LocationID   string
	LessonStatus domain.LessonSchedulingStatus
	StartDate    time.Time
	EndDate      time.Time
	StartTime    time.Time
	EndTime      time.Time
	Timezone     string
}
