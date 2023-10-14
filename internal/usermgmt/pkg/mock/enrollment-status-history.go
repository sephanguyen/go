package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type EnrollmentStatusHistory struct {
	RandomEnrollmentStatusHistory
}

type RandomEnrollmentStatusHistory struct {
	entity.DefaultDomainEnrollmentStatusHistory
	UserID           field.String
	LocationID       field.String
	EnrollmentStatus field.String
	StartDate        field.Time
	EndDate          field.Time
	CreatedAt        field.Time
}

func (s EnrollmentStatusHistory) UserID() field.String {
	return s.RandomEnrollmentStatusHistory.UserID
}
func (s EnrollmentStatusHistory) LocationID() field.String {
	return s.RandomEnrollmentStatusHistory.LocationID
}
func (s EnrollmentStatusHistory) EnrollmentStatus() field.String {
	return s.RandomEnrollmentStatusHistory.EnrollmentStatus
}
func (s EnrollmentStatusHistory) StartDate() field.Time {
	return s.RandomEnrollmentStatusHistory.StartDate
}
func (s EnrollmentStatusHistory) EndDate() field.Time {
	return s.RandomEnrollmentStatusHistory.EndDate
}
func (s EnrollmentStatusHistory) CreatedAt() field.Time {
	return s.RandomEnrollmentStatusHistory.CreatedAt
}
