package commands

import (
	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
	lpb "github.com/manabie-com/backend/pkg/manabuf/lessonmgmt/v1"
)

type UpdateLessonOneTimeCommandRequest struct {
	Lesson        *domain.Lesson
	CurrentLesson *domain.Lesson
	TimeZone      string
}

type UpdateRecurringLessonCommandRequest struct {
	SelectedLesson *domain.Lesson
	CurrentLesson  *domain.Lesson
	RRuleCmd       RecurrenceRuleCommand
	TimeZone       string
	ZoomInfo       *ZoomInfo
}

type UpdateLessonStatusCommandRequest struct {
	LessonID         string
	SchedulingStatus string
	SavingType       lpb.SavingType
}

type UpdateLessonStatusCommandResponse struct {
	UpdatedLesson []*domain.Lesson
}

type BulkUpdateLessonSchedulingStatusCommandRequest struct {
	Lessons []*domain.Lesson
	Action  lpb.LessonBulkAction
}

type BulkUpdateLessonSchedulingStatusCommandResponse struct {
	UpdatedLessons []*domain.Lesson
}

type MarkStudentAsReallocateRequest struct {
	Member        *domain.LessonMember
	ReAllocations *domain.Reallocation
}
