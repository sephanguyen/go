package queries

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/infrastructure"
)

type GetExternalConfigurationQueryHandler struct {
	DB         database.Ext
	ConfigRepo infrastructure.ExternalConfigRepo
}

func (g *GetExternalConfigurationQueryHandler) SearchWithKey(ctx context.Context, payload GetExternalConfigurations) ([]*domain.ExternalConfiguration, error) {
	cfs, err := g.ConfigRepo.SearchWithKey(ctx, g.DB, payload.SearchOption)
	if err != nil {
		return nil, err
	}
	return cfs, nil
}

func (g *GetExternalConfigurationQueryHandler) GetByKey(ctx context.Context, payload GetExternalConfigurationByKey) (*domain.ExternalConfiguration, error) {
	c, err := g.ConfigRepo.GetByKey(ctx, g.DB, payload.Key)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (g *GetExternalConfigurationQueryHandler) GetLocationConfigByKeysAndLocations(ctx context.Context, keys, locationIDs []string) ([]*domain.LocationConfiguration, error) {
	cf, err := g.ConfigRepo.GetByKeysAndLocations(ctx, g.DB, keys, locationIDs)
	if err != nil {
		return nil, fmt.Errorf("ConfigRepo.GetByKeysAndLocations: %w", err)
	}
	return cf, nil
}

func (g *GetExternalConfigurationQueryHandler) GetLocationConfigByKeys(ctx context.Context, keys, locationIDs []string) ([]*domain.LocationConfigurationV2, error) {
	cf, err := g.ConfigRepo.GetByKeysAndLocationsV2(ctx, g.DB, keys, locationIDs)
	if err != nil {
		return nil, fmt.Errorf("ConfigRepo.GetByKeysAndLocationsV2: %w", err)
	}
	return cf, nil
}
