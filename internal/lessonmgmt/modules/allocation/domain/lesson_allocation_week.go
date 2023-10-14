package domain

import (
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type LessonAllocationInfo struct {
	LessonID         string
	StartTime        time.Time
	EndTime          time.Time
	LocationID       string
	AttendanceStatus domain.StudentAttendStatus
	Status           domain.LessonSchedulingStatus
	TeachingMethod   domain.LessonTeachingMethod
	LessonReportID   string
	IsLocked         bool
}
