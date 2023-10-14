package payloads

import (
	"time"

	"github.com/manabie-com/backend/internal/lessonmgmt/modules/assigned_student/domain"
)

type GetAssignedStudentListArg struct {
	PurchaseMethod string
	SchoolID       string
	StudentIDs     []string
	CourseIDs      []string
	LocationIDs    []string
	FromDate       time.Time
	ToDate         time.Time

	Limit                 uint32
	KeyWord               string
	StudentSubscriptionID string

	Timezone                  string
	AssignedStudentStatuses   []domain.AssignedStudentStatus
}
