package domain

import (
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
	npb "github.com/manabie-com/backend/pkg/manabuf/nats/v1"
)

type TemporaryLocationAssignmentAttribute struct {
	StudentID      string
	LocationID     string
	OrganizationID string
	StartDate      time.Time
	EndDate        time.Time
}

type TemporaryLocationAssign struct {
	TemporaryLocationAssignmentAttribute
}

type TemporaryLocationAssignment struct {
	*npb.LessonReallocateStudentEnrollmentStatusEvent_StudentEnrollmentStatusInfo
}

func (t *TemporaryLocationAssignment) Validate() error {
	if t.StartDate.AsTime().After(t.EndDate.AsTime()) {
		return fmt.Errorf("end_date must be greater than or equal to start_date")
	}
	return nil
}

func NewTemporaryLocationAssignment(attribute TemporaryLocationAssignmentAttribute) *TemporaryLocationAssign {
	return &TemporaryLocationAssign{attribute}
}

func (t *TemporaryLocationAssign) UserID() field.String {
	return field.NewString(t.StudentID)
}

func (t *TemporaryLocationAssign) LocationID() field.String {
	return field.NewString(t.TemporaryLocationAssignmentAttribute.LocationID)
}

func (t *TemporaryLocationAssign) EnrollmentStatus() field.String {
	return field.NewString(entity.StudentEnrollmentStatusTemporary)
}

func (t *TemporaryLocationAssign) StartDate() field.Time {
	return field.NewTime(t.TemporaryLocationAssignmentAttribute.StartDate)
}

func (t *TemporaryLocationAssign) EndDate() field.Time {
	return field.NewTime(t.TemporaryLocationAssignmentAttribute.EndDate)
}

func (t *TemporaryLocationAssign) Comment() field.String {
	return field.NewNullString()
}

func (t *TemporaryLocationAssign) OrderID() field.String {
	return field.NewNullString()
}

func (t *TemporaryLocationAssign) OrderSequenceNumber() field.Int32 {
	return field.NewNullInt32()
}

func (t *TemporaryLocationAssign) OrganizationID() field.String {
	return field.NewString(t.TemporaryLocationAssignmentAttribute.OrganizationID)
}

func (t *TemporaryLocationAssign) CreatedAt() field.Time {
	return field.NewNullTime()
}
