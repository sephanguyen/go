package domain

import (
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/course_location_schedule/domain"
)

type (
	LessonAllocationStatus string
	CourseTeachingMethod   string
)

const (
	All               LessonAllocationStatus = "ALL"
	NoneAssigned      LessonAllocationStatus = "NONE_ASSIGNED"
	PartiallyAssigned LessonAllocationStatus = "PARTIALLY_ASSIGNED"
	FullyAssigned     LessonAllocationStatus = "FULLY_ASSIGNED"
	OverAssigned      LessonAllocationStatus = "OVER_ASSIGNED"

	None       CourseTeachingMethod = "COURSE_TEACHING_METHOD_NONE"
	Individual CourseTeachingMethod = "COURSE_TEACHING_METHOD_INDIVIDUAL"
	Group      CourseTeachingMethod = "COURSE_TEACHING_METHOD_GROUP"
)

type AllocatedStudent struct {
	StudentSubscriptionID string
	StudentID             string
	CourseID              string
	LocationID            string
	StartTime             time.Time
	EndTime               time.Time
	PurchasedSlot         int32
	AssignedSlot          int32
	ProductTypeSchedule   string
}

func (a *AllocatedStudent) IsWeeklySchedule() bool {
	return a.ProductTypeSchedule == string(domain.Frequency) || a.ProductTypeSchedule == string(domain.Scheduled)
}

func (a *AllocatedStudent) AllocationStatus() string {
	switch {
	case a.AssignedSlot == 0:
		return string(NoneAssigned)
	case a.AssignedSlot < a.PurchasedSlot:
		return string(PartiallyAssigned)
	case a.AssignedSlot > a.PurchasedSlot:
		return string(OverAssigned)
	default:
		return string(FullyAssigned)
	}
}

type LessonAllocationFilter struct {
	CourseID               []string
	CourseTypeID           []string
	TeachingMethod         []CourseTeachingMethod
	LocationID             []string
	StartDate              time.Time
	EndDate                time.Time
	LessonAllocationStatus LessonAllocationStatus
	IsClassUnassigned      bool
	IsOnlyReallocation     bool
	KeySearch              string
	ProductID              []string
	TimeZone               string
	Limit                  int
	Offset                 int
}
