package domain

import (
	"fmt"
)

type (
	AssignedStudentStatus string
	PurchaseMethod        string
)

const (
	AssignedStudentStatusUnderAssigned AssignedStudentStatus = "STUDENT_STATUS_UNDER_ASSIGNED"
	AssignedStudentStatusJustAssigned  AssignedStudentStatus = "STUDENT_STATUS_JUST_ASSIGNED"
	AssignedStudentStatusOverAssigned  AssignedStudentStatus = "STUDENT_STATUS_OVER_ASSIGNED"
	PurchaseMethodRecurring            PurchaseMethod        = "PURCHASE_METHOD_RECURRING"
	PurchaseMethodSlot                 PurchaseMethod        = "PURCHASE_METHOD_SLOT"
)

type AssignedStudent struct {
	StudentID             string
	CourseID              string
	LocationID            string
	Duration              string
	PurchasedSlot         int32
	AssignedSlot          int32
	SlotGap               int32
	AssignedStatus        AssignedStudentStatus
	StudentSubscriptionID string
}

type AssignedStudentBuilder struct {
	asgStudent *AssignedStudent
}

func NewAssignedStudent() *AssignedStudentBuilder {
	return &AssignedStudentBuilder{
		asgStudent: &AssignedStudent{},
	}
}

// BuildDraft will return a asgStudent but not valid data
func (at *AssignedStudentBuilder) BuildDraft() *AssignedStudent {
	return at.asgStudent
}

func (at *AssignedStudentBuilder) WithID(id string) *AssignedStudentBuilder {
	at.asgStudent.StudentID = id
	return at
}

func (at *AssignedStudentBuilder) WithCourseID(courseID string) *AssignedStudentBuilder {
	at.asgStudent.CourseID = courseID
	return at
}

func (at *AssignedStudentBuilder) WithDuration(duration string) *AssignedStudentBuilder {
	at.asgStudent.Duration = duration
	return at
}

func (at *AssignedStudentBuilder) WithPurchaseSlot(purchaseSlot int32) *AssignedStudentBuilder {
	at.asgStudent.PurchasedSlot = purchaseSlot
	return at
}

func (at *AssignedStudentBuilder) WithAssignedSlot(assignedSlot int32) *AssignedStudentBuilder {
	at.asgStudent.AssignedSlot = assignedSlot
	return at
}

func (at *AssignedStudentBuilder) WithSlotGap(slotGap int32) *AssignedStudentBuilder {
	at.asgStudent.SlotGap = slotGap
	return at
}

func (at *AssignedStudentBuilder) WithLocationID(locationID string) *AssignedStudentBuilder {
	at.asgStudent.LocationID = locationID
	return at
}

func (at *AssignedStudentBuilder) WithAssignedStatus(status AssignedStudentStatus) *AssignedStudentBuilder {
	at.asgStudent.AssignedStatus = status
	return at
}

func (at *AssignedStudentBuilder) WithStudentSubscriptionID(studentSubID string) *AssignedStudentBuilder {
	at.asgStudent.StudentSubscriptionID = studentSubID
	return at
}

func (a AssignedStudent) IsValid() error {
	if len(a.StudentID) == 0 {
		return fmt.Errorf("AssignedStudent.StudentID cannot be empty")
	}

	if len(a.CourseID) == 0 {
		return fmt.Errorf("AssignedStudent.CourseID of student %s cannot be empty", a.StudentID)
	}

	if len(a.Duration) == 0 {
		return fmt.Errorf("AssignedStudent.Duration of student %s cannot be empty", a.StudentID)
	}

	if len(a.LocationID) == 0 {
		return fmt.Errorf("AssignedStudent.LocationID of student %s cannot be empty", a.StudentID)
	}

	return nil
}
