package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Course interface {
	CourseID() field.String
	CoursePartnerID() field.String
}

func (courses DomainCourses) CourseIDs() []string {
	courseIDs := []string{}
	for _, course := range courses {
		courseIDs = append(courseIDs, course.CourseID().String())
	}
	return courseIDs
}

type DomainCourses []DomainCourse

type DomainCourse interface {
	Course
	valueobj.HasOrganizationID
}

type DefaultDomainCourse struct{}

func (e DefaultDomainCourse) CourseID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainCourse) CoursePartnerID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainCourse) OrganizationID() field.String {
	return field.NewNullString()
}
