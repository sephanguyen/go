package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainSchoolAdminProfile interface {
	UserProfile
}

// DomainSchoolAdmin represents a school admin entity in business
type DomainSchoolAdmin interface {
	DomainSchoolAdminProfile
	valueobj.HasOrganizationID
	valueobj.HasSchoolID
	valueobj.HasCountry
	valueobj.HasUserID
	valueobj.HasLoginEmail
}

type SchoolAdminToDelegate struct {
	DomainSchoolAdminProfile
	valueobj.HasSchoolID
	valueobj.HasOrganizationID
	valueobj.HasCountry
	valueobj.HasUserID
	valueobj.HasLoginEmail
}

type NullDomainSchoolAdmin struct {
	EmptyUser
}

func (schoolAdmin NullDomainSchoolAdmin) SchoolID() field.Int32 {
	return field.NewNullInt32()
}

// ValidSchoolAdmin func following business logic to validate a school admin
// Ref for product specs: <https://product-specs.com>
// Returns domain error if there are any violations
func ValidSchoolAdmin(schoolAdmin DomainSchoolAdmin) error {
	switch {
	case field.IsNull(schoolAdmin.UserID()):
		break
	}
	return nil
}
