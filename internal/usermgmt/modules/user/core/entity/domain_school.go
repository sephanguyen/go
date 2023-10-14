package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type SchoolAttribute interface {
	SchoolID() field.String
	Name() field.String
	NamePhonetic() field.String
	SchoolLevelID() field.String
	Address() field.String
	IsArchived() field.Boolean
}

type DomainSchool interface {
	SchoolAttribute
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type SchoolWillBeDelegated struct {
	SchoolAttribute
	valueobj.HasPartnerInternalID
	valueobj.HasOrganizationID
}

type DefaultDomainSchool struct{}

func (e DefaultDomainSchool) SchoolID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchool) SchoolLevelID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchool) Name() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchool) NamePhonetic() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchool) IsArchived() field.Boolean {
	return field.NewNullBoolean()
}
func (e DefaultDomainSchool) PartnerInternalID() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchool) Address() field.String {
	return field.NewNullString()
}
func (e DefaultDomainSchool) OrganizationID() field.String {
	return field.NewNullString()
}

type DomainSchools []DomainSchool

func (schools DomainSchools) SchoolIDs() []string {
	schoolIDs := []string{}
	for _, school := range schools {
		schoolIDs = append(schoolIDs, school.SchoolID().String())
	}
	return schoolIDs
}

func (schools DomainSchools) PartnerInternalIDs() []string {
	partnerInternalIDs := []string{}
	for _, school := range schools {
		partnerInternalIDs = append(partnerInternalIDs, school.PartnerInternalID().String())
	}
	return partnerInternalIDs
}
