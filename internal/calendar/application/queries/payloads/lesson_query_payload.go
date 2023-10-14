package payloads

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/calendar/domain/dto"
	lesson_domain "github.com/manabie-com/backend/internal/lessonmgmt/modules/lesson/domain"
)

type GetLessonDetailRequest struct {
	LessonID string
}

type GetLessonDetailResponse struct {
	Lesson    *lesson_domain.Lesson
	Scheduler *dto.Scheduler
}

func (r *GetLessonDetailRequest) Validate() error {
	if len(r.LessonID) == 0 {
		return fmt.Errorf("lesson id cannot be empty")
	}

	return nil
}

type GetLessonIDsForBulkStatusUpdateRequest struct {
	LocationID string
	Action     lesson_domain.LessonBulkAction
	StartDate  time.Time
	EndDate    time.Time
	StartTime  time.Time
	EndTime    time.Time
	Timezone   string
}

type GetLessonIDsForBulkStatusUpdateResponse struct {
	LessonStatus           lesson_domain.LessonSchedulingStatus
	ModifiableLessonsCount uint32
	LessonsCount           uint32
	LessonIDs              []string
}

func (r *GetLessonIDsForBulkStatusUpdateRequest) Validate() error {
	if len(r.LocationID) == 0 {
		return fmt.Errorf("location id cannot be empty")
	}

	if r.StartDate.IsZero() {
		return fmt.Errorf("start date could not be empty")
	}

	if r.EndDate.IsZero() {
		return fmt.Errorf("end date could not be empty")
	}

	if r.EndDate.Before(r.StartDate) {
		return fmt.Errorf("end date could not before start date")
	}

	if !r.StartTime.IsZero() || !r.EndTime.IsZero() {
		if r.StartTime.IsZero() {
			return fmt.Errorf("start time could not be empty")
		}

		if r.EndTime.IsZero() {
			return fmt.Errorf("end time could not be empty")
		}

		if r.EndTime.Before(r.StartTime) {
			return fmt.Errorf("end time could not before start time")
		}
	}

	return nil
}
