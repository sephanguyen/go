package mock

import (
	"github.com/manabie-com/backend/internal/usermgmt/modules/user/core/entity"
	"github.com/manabie-com/backend/internal/usermgmt/pkg/field"
)

type Configuration struct {
	RandomConfiguration
}

type RandomConfiguration struct {
	entity.NullDomainConfiguration
	ConfigKey   field.String
	ConfigValue field.String
}

func (c Configuration) ConfigKey() field.String {
	return c.RandomConfiguration.ConfigKey
}
func (c Configuration) ConfigValue() field.String {
	return c.RandomConfiguration.ConfigValue
}
