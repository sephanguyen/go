package entity

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/valueobj"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Configuration interface {
	ConfigID() field.String
	ConfigKey() field.String
	ConfigValue() field.String
	ConfigValueType() field.String
}

type DomainConfiguration interface {
	Configuration
	valueobj.HasOrganizationID
}

type NullDomainConfiguration struct{}

func (config NullDomainConfiguration) ConfigID() field.String {
	return field.NewNullString()
}

func (config NullDomainConfiguration) ConfigKey() field.String {
	return field.NewNullString()
}

func (config NullDomainConfiguration) ConfigValue() field.String {
	return field.NewNullString()
}

func (config NullDomainConfiguration) ConfigValueType() field.String {
	return field.NewNullString()
}

func (config NullDomainConfiguration) OrganizationID() field.String {
	return field.NewNullString()
}
