package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type SchoolCourseAttribute interface {
	SchoolCourseID() field.String
	Name() field.String
	NamePhonetic() field.String
	SchoolID() field.String
	IsArchived() field.Boolean
}

type DomainSchoolCourse interface {
	SchoolCourseAttribute
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type SchoolCourseWillBeDelegated struct {
	SchoolCourseAttribute
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type DefaultDomainSchoolCourse struct{}

func (e DefaultDomainSchoolCourse) SchoolCourseID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolCourse) SchoolID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolCourse) Name() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolCourse) NamePhonetic() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolCourse) IsArchived() field.Boolean {
	return field.NewNullBoolean()
}
func (e DefaultDomainSchoolCourse) PartnerInternalID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolCourse) OrganizationID() field.String {
	return field.NewNullString()
}

type DomainSchoolCourses []DomainSchoolCourse

func (schoolCourses DomainSchoolCourses) SchoolCourseIDs() []string {
	schoolCourseIDs := []string{}
	for _, schoolCourse := range schoolCourses {
		schoolCourseIDs = append(schoolCourseIDs, schoolCourse.SchoolCourseID().String())
	}
	return schoolCourseIDs
}

func (schoolCourses DomainSchoolCourses) PartnerInternalIDs() []string {
	partnerInternalIDs := []string{}
	for _, schoolCourse := range schoolCourses {
		partnerInternalIDs = append(partnerInternalIDs, schoolCourse.PartnerInternalID().String())
	}
	return partnerInternalIDs
}
