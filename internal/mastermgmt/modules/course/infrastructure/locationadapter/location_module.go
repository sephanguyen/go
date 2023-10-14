package locationadapter

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/interceptors"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
)

type LocationAdapter struct {
	LocationRepo repo.LocationRepo
}

func (l *LocationAdapter) GetLocationsByLocationIDs(ctx context.Context, db database.Ext, locationIDs []string) (ids []string, err error) {
	ctx, span := interceptors.StartSpan(ctx, "LocationAdapter.GetLocationsByLocationIDs")
	defer span.End()
	locations, err := l.LocationRepo.GetLocationsByLocationIDs(ctx, db, database.TextArray(locationIDs), false)
	if err != nil {
		return ids, fmt.Errorf("locationRepo GetLocationsByLocationIDs err: %w", err)
	}

	for _, location := range locations {
		ids = append(ids, location.LocationID)
	}

	return ids, nil
}
