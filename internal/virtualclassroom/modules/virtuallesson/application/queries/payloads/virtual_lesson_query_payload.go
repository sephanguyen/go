package payloads

import (
	"time"

	"github.com/manabie-com/backend/internal/virtualclassroom/modules/virtualclassroom/domain"
)

type (
	TimeLookup        string
	LiveLessonStatus  string
	LessonTimeCompare string
)

const (
	TimeLookupStartTime                  TimeLookup = "TIME_LOOKUP_START_TIME"
	TimeLookupEndTime                    TimeLookup = "TIME_LOOKUP_END_TIME"
	TimeLookupEndTimeIncludeWithoutEndAt TimeLookup = "TIME_LOOKUP_END_TIME_INCLUDE_WITHOUT_END_AT"
	TimeLookupEndTimeIncludeWithEndAt    TimeLookup = "TIME_LOOKUP_END_TIME_INCLUDE_WITH_END_AT"

	LiveLessonStatusNone     LiveLessonStatus = "LIVE_LESSON_STATUS_NONE"
	LiveLessonStatusNotEnded LiveLessonStatus = "LIVE_LESSON_STATUS_NOT_ENDED"
	LiveLessonStatusEnded    LiveLessonStatus = "LIVE_LESSON_STATUS_ENDED"

	LessonTimeCompareFuture         LessonTimeCompare = ">"
	LessonTimeCompareFutureAndEqual LessonTimeCompare = ">="
	LessonTimeComparePast           LessonTimeCompare = "<"
	LessonTimeComparePastAndEqual   LessonTimeCompare = "<="
)

type GetLiveLessonsByLocationsRequest struct {
	LocationIDs              []string
	CourseIDs                []string
	StartDate                time.Time
	EndDate                  time.Time
	LessonSchedulingStatuses []domain.LessonSchedulingStatus

	GetWhitelistCourseIDs bool

	Limit int32
	Page  int32
}

type GetLiveLessonsByLocationsResponse struct {
	Lessons []*domain.VirtualLesson
	Total   int32
}

type GetVirtualLessonsArgs struct {
	StudentIDs               []string
	CourseIDs                []string
	LocationIDs              []string
	StartDate                time.Time
	EndDate                  time.Time
	LessonSchedulingStatuses []domain.LessonSchedulingStatus

	ReplaceCourseIDColumn bool

	Limit int32
	Page  int32
}

type GetLearnersByLessonIDArgs struct {
	LessonID   string
	StudentIDs []string
	Limit      int32

	// used for pagination
	LessonCourseID string
	UserID         string
}

type GetLessonMemberUsersByLessonIDArgs struct {
	LessonID   string
	StudentIDs []string

	// use lessonmgmt db
	UseLessonmgmtDB bool
}

type GetLearnersByLessonIDResponse struct {
	StudentIDs  []string
	StudentInfo map[string][]*domain.StudentEnrollmentStatusHistory
	Limit       int32

	// used for pagination
	LastLessonCourseID string
	LastUserID         string
}

type GetLessonsArgs struct {
	CurrentTime       time.Time
	TimeLookup        TimeLookup
	LessonTimeCompare LessonTimeCompare
	SortAscending     bool
	SchoolID          string

	// paging
	Limit          uint32
	OffsetLessonID string

	// filters
	LocationIDs              []string
	TeacherIDs               []string
	StudentIDs               []string
	CourseIDs                []string
	LessonSchedulingStatuses []domain.LessonSchedulingStatus
	LiveLessonStatus         LiveLessonStatus
	FromDate                 time.Time
	ToDate                   time.Time
}
