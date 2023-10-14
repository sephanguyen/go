package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainStudentCourseAttribute interface {
	StudentPackageID() field.String
	CourseID() field.String
	StartAt() field.Time
	EndAt() field.Time
}

type DomainStudentCourse interface {
	DomainStudentCourseAttribute
	valueobj.HasUserID
	valueobj.HasLocationID
}

type DomainStudentCourses []DomainStudentCourse

type StudentCourseWillBeDelegated struct {
	DomainStudentCourseAttribute
	valueobj.HasUserID
	valueobj.HasLocationID
}

type DefaultDomainStudentCourse struct{}

func (e DefaultDomainStudentCourse) CourseID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainStudentCourse) StartAt() field.Time {
	return field.NewNullTime()
}

func (e DefaultDomainStudentCourse) EndAt() field.Time {
	return field.NewNullTime()
}

func (e DefaultDomainStudentCourse) StudentID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainStudentCourse) LocationID() field.String {
	return field.NewNullString()
}

func (e DefaultDomainStudentCourse) StudentPackageID() field.String {
	return field.NewNullString()
}
