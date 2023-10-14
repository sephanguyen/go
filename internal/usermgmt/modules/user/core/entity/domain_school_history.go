package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type SchoolHistoryAttribute interface {
	StartDate() field.Time
	EndDate() field.Time
}
type SchoolHistoryAttributeIsCurrentSchool interface {
	IsCurrent() field.Boolean
}

type DomainSchoolHistory interface {
	SchoolHistoryAttribute
	valueobj.HasSchoolInfoID
	valueobj.HasSchoolCourseID
	valueobj.HasUserID
	valueobj.HasOrganizationID
	SchoolHistoryAttributeIsCurrentSchool
}

type SchoolHistoryWillBeDelegated struct {
	SchoolHistoryAttribute
	valueobj.HasSchoolInfoID
	valueobj.HasSchoolCourseID
	valueobj.HasUserID
	valueobj.HasOrganizationID
	SchoolHistoryAttributeIsCurrentSchool
}

type DefaultDomainSchoolHistory struct{}

func (e DefaultDomainSchoolHistory) SchoolID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolHistory) SchoolCourseID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolHistory) UserID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchoolHistory) IsCurrent() field.Boolean {
	return field.NewNullBoolean()
}
func (e DefaultDomainSchoolHistory) StartDate() field.Time {
	return field.NewNullTime()
}
func (e DefaultDomainSchoolHistory) EndDate() field.Time {
	return field.NewNullTime()
}
func (e DefaultDomainSchoolHistory) OrganizationID() field.String {
	return field.NewNullString()
}

type DomainSchoolHistories []DomainSchoolHistory

func (schoolHistories DomainSchoolHistories) UserIDs() []string {
	userIDs := []string{}
	for _, schoolHistory := range schoolHistories {
		userIDs = append(userIDs, schoolHistory.UserID().String())
	}

	return userIDs
}
func (schoolHistories DomainSchoolHistories) SchoolIDs() []string {
	schoolIDs := make([]string, 0, len(schoolHistories))
	for _, schoolHistory := range schoolHistories {
		schoolIDs = append(schoolIDs, schoolHistory.SchoolID().String())
	}

	return schoolIDs
}

type SchoolHistoryCurrentSchool struct {
	IsCurrentSchool bool
}

func (schoolHistoryCurrentSchool SchoolHistoryCurrentSchool) IsCurrent() field.Boolean {
	return field.NewBoolean(schoolHistoryCurrentSchool.IsCurrentSchool)
}
