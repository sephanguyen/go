package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type SchoolCourse struct {
	RandomSchoolCourse
}

type RandomSchoolCourse struct {
	entity.DefaultDomainSchoolCourse
	SchoolCourseID    field.String
	PartnerInternalID field.String
	SchoolID          field.String
	IsArchived        field.Boolean
}

func (s SchoolCourse) SchoolCourseID() field.String {
	return s.RandomSchoolCourse.SchoolCourseID
}
func (s SchoolCourse) PartnerInternalID() field.String {
	return s.RandomSchoolCourse.PartnerInternalID
}
func (s SchoolCourse) SchoolID() field.String {
	return s.RandomSchoolCourse.SchoolID
}
func (s SchoolCourse) IsArchived() field.Boolean {
	return s.RandomSchoolCourse.IsArchived
}
