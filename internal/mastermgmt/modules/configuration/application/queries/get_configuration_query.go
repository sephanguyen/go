package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/infrastructure"
	external_domain "github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
)

type GetConfigurationQueryHandler struct {
	DB                 database.Ext
	ConfigRepo         infrastructure.ConfigRepo
	ExternalConfigRepo infrastructure.ExternalConfigRepo
}

func (g *GetConfigurationQueryHandler) SearchWithKey(ctx context.Context, payload GetConfigurations) ([]*domain.InternalConfiguration, error) {
	cfs, err := g.ConfigRepo.SearchWithKey(ctx, g.DB, payload.SearchOption)
	if err != nil {
		return nil, err
	}
	return cfs, nil
}

func (g *GetConfigurationQueryHandler) SearchExternalConfigWithKey(ctx context.Context, payload GetConfigurations) ([]*external_domain.ExternalConfiguration, error) {
	externalPayload := external_domain.ExternalConfigSearchArgs{
		Keyword: payload.SearchOption.Keyword,
		Limit:   payload.SearchOption.Limit,
		Offset:  payload.SearchOption.Offset,
	}
	cfs, err := g.ExternalConfigRepo.SearchWithKey(ctx, g.DB, externalPayload)
	if err != nil {
		return nil, err
	}
	return cfs, nil
}

func (g *GetConfigurationQueryHandler) GetByKey(ctx context.Context, payload GetConfigurationByKey) (*domain.InternalConfiguration, error) {
	c, err := g.ConfigRepo.GetByKey(ctx, g.DB, payload.Key)
	if err != nil {
		return nil, err
	}
	return c, nil
}
