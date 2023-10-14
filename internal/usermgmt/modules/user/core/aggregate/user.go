package aggregate

import "github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"

type User interface {
	entity.User
	entity.DomainOrganization

	LegacyUserGroups() entity.LegacyUserGroups
}

type NullUser struct {
	entity.EmptyUser
	entity.NullOrganization
}

func (nullUser NullUser) LegacyUserGroups() entity.LegacyUserGroups {
	return nil
}
