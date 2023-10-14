package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type DomainOrganization interface {
	valueobj.HasOrganizationID
	valueobj.HasSchoolID
}

/*func ValidOrganization(organization DomainOrganization) error {
	return nil
}*/

type NullOrganization struct{}

func (ent NullOrganization) OrganizationID() field.String {
	return field.NewNullString()
}
