package aggregate

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type DomainParent struct {
	// aggregate
	entity.DomainParent
	IndexAttr        int // additional info
	UserAccessPaths  entity.DomainUserAccessPaths
	UserGroupMembers entity.DomainUserGroupMembers
	LegacyUserGroups entity.LegacyUserGroups
	TaggedUsers      entity.DomainTaggedUsers
	UserPhoneNumbers entity.DomainUserPhoneNumbers
}

func (parent DomainParent) Index() int {
	return parent.IndexAttr
}

type DomainParents []DomainParent

func (parents DomainParents) ParentIDs() []string {
	parentIDs := make([]string, 0, len(parents))
	for _, parent := range parents {
		parentIDs = append(parentIDs, parent.UserID().String())
	}
	return parentIDs
}

func (parents DomainParents) ParentExternalUserIDs() []string {
	parentExternalUserIDs := make([]string, 0, len(parents))
	for _, parent := range parents {
		parentExternalUserIDs = append(parentExternalUserIDs, parent.ExternalUserID().String())
	}
	return parentExternalUserIDs
}

func (parents DomainParents) Users() entity.Users {
	users := make(entity.Users, len(parents))
	for i := range parents {
		users[i] = parents[i]
	}
	return users
}

type DomainParentWithChildren struct {
	DomainParent
	Children entity.DomainStudentParentRelationships
}

type DomainParentWithChildrens []DomainParentWithChildren
