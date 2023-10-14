package queries

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/course/infrastructure"
)

type LocationQueryHandler struct {
	DB database.Ext

	// ports
	LocationModulePort infrastructure.LocationModulePort
}

func (l *LocationQueryHandler) GetLocationsByLocationIDs(ctx context.Context, payload GetLocationsByMultipleIDQuery) ([]string, error) {
	return l.LocationModulePort.GetLocationsByLocationIDs(ctx, l.DB, payload.LocationIDs)
}
