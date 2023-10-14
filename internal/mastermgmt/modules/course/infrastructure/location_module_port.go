package infrastructure

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
)

type LocationModulePort interface {
	GetLocationsByLocationIDs(ctx context.Context, db database.Ext, locationIDs []string) ([]string, error)
}
