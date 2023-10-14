package queries

import "github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"

type GetExternalConfigurationByKey struct {
	Key string
}

type GetExternalConfigurations struct {
	SearchOption domain.ExternalConfigSearchArgs
}
