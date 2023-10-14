package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type OrganizationAttribute interface {
	OrganizationName() field.String
	TenantID() field.String
	SalesforceClientID() field.String
}

type DomainOrganization interface {
	OrganizationAttribute

	valueobj.HasOrganizationID
}

/*func ValidOrganization(organization DomainOrganization) error {
	return nil
}*/

type NullOrganization struct{}

func (ent NullOrganization) OrganizationID() field.String {
	return field.NewNullString()
}
func (ent NullOrganization) OrganizationName() field.String {
	return field.NewNullString()
}
func (ent NullOrganization) TenantID() field.String {
	return field.NewNullString()
}
func (ent NullOrganization) SalesforceClientID() field.String {
	return field.NewNullString()
}
