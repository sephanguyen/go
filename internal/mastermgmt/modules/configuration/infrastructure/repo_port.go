package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/configuration/domain"
	domain_external "github.com/manabie-com/backend/internal/mastermgmt/modules/external_configuration/domain"
)

type ConfigRepo interface {
	GetByKey(ctx context.Context, db database.QueryExecer, cKey string) (c *domain.InternalConfiguration, err error)
	GetByMultipleKeys(ctx context.Context, db database.QueryExecer, cKey []string) (c []*domain.InternalConfiguration, err error)
	SearchWithKey(ctx context.Context, db database.QueryExecer, payload domain.ConfigSearchArgs) (c []*domain.InternalConfiguration, err error)
}

type ExternalConfigRepo interface {
	SearchWithKey(ctx context.Context, db database.QueryExecer, payload domain_external.ExternalConfigSearchArgs) (c []*domain_external.ExternalConfiguration, err error)
}
