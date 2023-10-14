package commands

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/infrastructure"
)

type CreateExternalConfigurationHandler struct {
	DB         database.Ext
	ConfigRepo infrastructure.ExternalConfigRepo
}

func (g *CreateExternalConfigurationHandler) CreateMultiConfigurations(ctx context.Context, payload []*domain.ExternalConfiguration) error {
	return g.ConfigRepo.CreateMultipleConfigs(ctx, g.DB, payload)
}
