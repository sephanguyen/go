package queries

import "github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"

type GetConfigurationByKey struct {
	Key string
}

type GetConfigurations struct {
	SearchOption domain.ConfigSearchArgs
}
