package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/constant"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

const (
	ParentTagsField = "parent_tag"
)

type DomainParentProfile interface {
	UserProfile
}

type DomainParent interface {
	DomainParentProfile
	valueobj.HasSchoolID
	valueobj.HasOrganizationID
	valueobj.HasUserID
	valueobj.HasCountry
	valueobj.HasLoginEmail
}

type ParentWillBeDelegated struct {
	DomainParentProfile
	valueobj.HasSchoolID
	valueobj.HasOrganizationID
	valueobj.HasUserID
	valueobj.HasCountry
	valueobj.HasLoginEmail
}

type NullDomainParent struct {
	EmptyUser
}

func (parent NullDomainParent) SchoolID() field.Int32 {
	return field.NewNullInt32()
}

func (parent NullDomainParent) Group() field.String {
	return field.NewString(constant.UserGroupParent)
}

func (parent NullDomainParent) UserRole() field.String {
	return field.NewString(string(constant.UserRoleStudent))
}

func ValidParent(parent DomainParent, isEnableUsername bool) error {
	return ValidUser(isEnableUsername, parent)
}
