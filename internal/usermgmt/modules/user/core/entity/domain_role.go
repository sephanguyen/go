package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type RoleAttribute interface {
	RoleID() field.String
	RoleName() field.String
	IsSystem() field.Boolean
}

type DomainRole interface {
	RoleAttribute
	valueobj.HasOrganizationID
}

type RoleWillBeDelegated interface {
	RoleAttribute
	valueobj.HasOrganizationID
}

type NullDomainRole struct{}

func (e NullDomainRole) RoleID() field.String {
	return field.NewNullString()
}
func (e NullDomainRole) RoleName() field.String {
	return field.NewNullString()
}
func (e NullDomainRole) IsSystem() field.Boolean {
	return field.NewNullBoolean()
}
func (e NullDomainRole) OrganizationID() field.String {
	return field.NewNullString()
}

type DomainRoles []DomainRole

func (roles DomainRoles) RoleNames() []string {
	roleNames := []string{}
	for _, role := range roles {
		if !role.RoleName().IsEmpty() {
			roleNames = append(roleNames, role.RoleName().String())
		}
	}
	return roleNames
}
