package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type StudentPackage interface {
	valueobj.HasStudentPackageID
	StartDate() field.Time
	EndDate() field.Time
	IsActive() field.Boolean
	LocationIDs() []field.String
}

type DomainStudentPackage interface {
	StudentPackage

	valueobj.HasPackageID
	valueobj.HasStudentID
	valueobj.HasOrganizationID
}

type StudentPackageWillBeDelegated struct {
	StudentPackage

	valueobj.HasPackageID
	valueobj.HasStudentID
	valueobj.HasOrganizationID
}

var _ DomainStudentPackage = (*DefaultDomainStudentPackage)(nil)

type DefaultDomainStudentPackage struct{}

func (domainStudentPackage DefaultDomainStudentPackage) StudentPackageID() field.String {
	return field.NewNullString()
}

func (domainStudentPackage DefaultDomainStudentPackage) StartDate() field.Time {
	return field.NewNullTime()
}

func (domainStudentPackage DefaultDomainStudentPackage) EndDate() field.Time {
	return field.NewNullTime()
}

func (domainStudentPackage DefaultDomainStudentPackage) IsActive() field.Boolean {
	return field.NewNullBoolean()
}

func (domainStudentPackage DefaultDomainStudentPackage) LocationIDs() []field.String {
	return make([]field.String, 0)
}

func (domainStudentPackage DefaultDomainStudentPackage) PackageID() field.String {
	return field.NewNullString()
}

func (domainStudentPackage DefaultDomainStudentPackage) StudentID() field.String {
	return field.NewNullString()
}

func (domainStudentPackage DefaultDomainStudentPackage) OrganizationID() field.String {
	return field.NewNullString()
}

type DomainStudentPackages []DomainStudentPackage
