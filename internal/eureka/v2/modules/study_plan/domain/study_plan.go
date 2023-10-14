package domain

import (
	"time"
)

// StudyPlan represents a plan that students need to learn.
type StudyPlan struct {
	ID           string
	Name         string
	CourseID     string
	AcademicYear string
	Status       StudyPlanStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    time.Time
}

type StudyPlanStatus string

const (
	StudyPlanStatusNone     StudyPlanStatus = "STUDY_PLAN_STATUS_NONE"
	StudyPlanStatusActive   StudyPlanStatus = "STUDY_PLAN_STATUS_ACTIVE"
	StudyPlanStatusArchived StudyPlanStatus = "STUDY_PLAN_STATUS_ARCHIVED"
)

func NewStudyPlan(studyPlan StudyPlan) (StudyPlan, error) {
	if studyPlan.ID == "" {
		return StudyPlan{}, ErrIDRequired
	}

	if studyPlan.CourseID == "" {
		return StudyPlan{}, ErrCourseIDRequired
	}

	return studyPlan, nil
}
