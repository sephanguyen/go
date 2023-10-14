package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainUserGroupMember interface {
	valueobj.HasUserGroupID
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

type DomainUserGroupMembers []DomainUserGroupMember

type UserGroupMemberWillBeDelegated struct {
	valueobj.HasUserGroupID
	valueobj.HasUserID
	valueobj.HasOrganizationID
}

type EmptyUserGroupMember struct{}

func (ugm EmptyUserGroupMember) UserGroupID() field.String {
	return field.NewNullString()
}

func (ugm EmptyUserGroupMember) UserID() field.String {
	return field.NewNullString()
}

func (ugm EmptyUserGroupMember) OrganizationID() field.String {
	return field.NewNullString()
}
