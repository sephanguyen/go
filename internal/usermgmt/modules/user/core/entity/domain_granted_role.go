package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainGrantedRole interface {
	ID() field.String
	UserGroupID() field.String
	RoleID() field.String
	valueobj.HasOrganizationID
}

type GrantedRoleWillBeDelegated struct {
	DomainGrantedRole
}
