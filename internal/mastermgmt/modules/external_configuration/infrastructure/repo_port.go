package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
)

type ExternalConfigRepo interface {
	GetByKey(ctx context.Context, db database.QueryExecer, cKey string) (c *domain.ExternalConfiguration, err error)
	GetByMultipleKeys(ctx context.Context, db database.QueryExecer, cKey []string) (c []*domain.ExternalConfiguration, err error)
	SearchWithKey(ctx context.Context, db database.QueryExecer, payload domain.ExternalConfigSearchArgs) (c []*domain.ExternalConfiguration, err error)
	CreateMultipleConfigs(ctx context.Context, db database.QueryExecer, configs []*domain.ExternalConfiguration) error
	GetByKeysAndLocations(ctx context.Context, db database.QueryExecer, configKeys, locationIDS []string) ([]*domain.LocationConfiguration, error)
	GetByKeysAndLocationsV2(ctx context.Context, db database.QueryExecer, configKeys, locationIDS []string) ([]*domain.LocationConfigurationV2, error)
}
