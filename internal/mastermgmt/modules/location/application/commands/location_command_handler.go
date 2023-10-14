package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"

	"github.com/jackc/pgx/v4"
)

type LocationCommandHandler struct {
	DB database.Ext

	// ports
	LocationRepo     infrastructure.LocationRepo
	LocationTypeRepo infrastructure.LocationTypeRepo
}

func (l *LocationCommandHandler) ImportLocationV2(ctx context.Context, payload UpsertLocation) (err error) {
	err = database.ExecInTx(ctx, l.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = l.LocationRepo.UpsertLocations(ctx, l.DB, payload.Locations)
		return err
	})
	if err != nil {
		return fmt.Errorf("LocationRepo.UpsertLocations: %w", err)
	}

	locIDs := sliceutils.Map(payload.Locations, func(l *domain.Location) string {
		return l.LocationID
	})
	err = l.LocationRepo.UpdateAccessPath(ctx, l.DB, locIDs)
	if err != nil {
		return fmt.Errorf("LocationRepo.UpdateAccessPath: %w", err)
	}

	return nil
}
