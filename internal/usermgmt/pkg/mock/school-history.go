package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type SchoolHistory struct {
	RandomSchoolHistory
}

type RandomSchoolHistory struct {
	entity.DefaultDomainSchoolHistory
	SchoolID       field.String
	SchoolCourseID field.String
	StartDate      field.Time
	EndDate        field.Time
}

func (s SchoolHistory) SchoolID() field.String {
	return s.RandomSchoolHistory.SchoolID
}
func (s SchoolHistory) SchoolCourseID() field.String {
	return s.RandomSchoolHistory.SchoolCourseID
}
func (s SchoolHistory) StartDate() field.Time {
	return s.RandomSchoolHistory.StartDate
}
func (s SchoolHistory) EndDate() field.Time {
	return s.RandomSchoolHistory.EndDate
}
