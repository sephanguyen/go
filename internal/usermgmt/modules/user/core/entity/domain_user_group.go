package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type UserGroupAttribute interface {
	UserGroupID() field.String
	Name() field.String
	OrgLocationID() field.String
	IsSystem() field.Boolean
}

type DomainUserGroup interface {
	UserGroupAttribute
	valueobj.HasOrganizationID
}

type UserGroupWillBeDelegated struct {
	UserGroupAttribute
	valueobj.HasOrganizationID
}
