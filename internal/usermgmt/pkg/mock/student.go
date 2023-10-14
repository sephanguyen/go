package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Student struct {
	RandomStudent
}

func NewStudent(id, externalID string) *Student {
	userID := field.NewString(id)
	externalUserID := field.NewString(externalID)
	if externalID == "" {
		externalUserID.SetNull()
	}

	return &Student{
		RandomStudent{
			UserID:         userID,
			ExternalUserID: externalUserID,
		},
	}
}

type StudentWithAssignedParent struct {
	Student
	Parents []Parent
}

type RandomStudent struct {
	entity.NullDomainStudent
	UserID            field.String
	GradeID           field.String
	Email             field.String
	Gender            field.String
	UserName          field.String
	FirstName         field.String
	LastName          field.String
	CurrentGrade      field.Int16
	EnrollmentStatus  field.String
	ContactPreference field.String
	ExternalUserID    field.String
	LoginEmail        field.String
}

func (m *Student) UserID() field.String {
	return m.RandomStudent.UserID
}

func (m *Student) GradeID() field.String {
	return m.RandomStudent.GradeID
}

func (m *Student) Email() field.String {
	return m.RandomStudent.Email
}

func (m *Student) Gender() field.String {
	return m.RandomStudent.Gender
}

func (m *Student) UserName() field.String {
	return m.RandomStudent.UserName
}

func (m *Student) FirstName() field.String {
	return m.RandomStudent.FirstName
}

func (m *Student) LastName() field.String {
	return m.RandomStudent.LastName
}

func (m *Student) CurrentGrade() field.Int16 {
	return m.RandomStudent.CurrentGrade
}

func (m *Student) EnrollmentStatus() field.String {
	return m.RandomStudent.EnrollmentStatus
}

func (m *Student) ContactPreference() field.String {
	return m.RandomStudent.ContactPreference
}

func (m *Student) ExternalUserID() field.String {
	return m.RandomStudent.ExternalUserID
}

func (m *Student) LoginEmail() field.String {
	return m.RandomStudent.LoginEmail
}
