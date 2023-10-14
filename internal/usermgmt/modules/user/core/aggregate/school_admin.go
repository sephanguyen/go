package aggregate

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
)

type DomainSchoolAdmin struct {
	entity.DomainSchoolAdmin
	LegacyUserGroups entity.LegacyUserGroups
}

type NullSchoolAdmin struct {
	entity.NullDomainSchoolAdmin
	LegacyUserGroups entity.LegacyUserGroups
}

func ValidSchoolAdmin(schoolAdmin DomainSchoolAdmin, isEnableUsername bool) error {
	if err := entity.ValidUser(isEnableUsername, schoolAdmin); err != nil {
		return err
	}
	if err := entity.ValidSchoolAdmin(schoolAdmin); err != nil {
		return err
	}
	return nil
}
