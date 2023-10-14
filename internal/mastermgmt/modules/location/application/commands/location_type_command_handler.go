package commands

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/sliceutils"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/domain"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure"
	"github.com/manabie-com/backend/internal/mastermgmt/modules/location/infrastructure/repo"
	"github.com/manabie-com/backend/internal/mastermgmt/shared/utils"

	"github.com/jackc/pgx/v4"
)

type LocationTypeCommandHandler struct {
	DB database.Ext

	// ports
	LocationTypeRepo infrastructure.LocationTypeRepo
}

func (l *LocationTypeCommandHandler) ImportLocationTypes(ctx context.Context, payload ImportLocationTypeV2Payload) *utils.BusinessError {
	dbLocationTypes, err := l.LocationTypeRepo.GetAllLocationTypes(ctx, l.DB)
	if err != nil {
		return utils.NewSystemError(err)
	}

	existingLocTypes := sliceutils.Map(dbLocationTypes, func(l *repo.LocationType) *domain.LocationType {
		return &domain.LocationType{
			LocationTypeID: l.LocationTypeID.String,
			Name:           l.Name.String,
			DisplayName:    l.DisplayName.String,
			Level:          int(l.Level.Int),
			CreatedAt:      l.CreatedAt.Time,
			UpdatedAt:      l.UpdatedAt.Time,
		}
	})
	existingLocTypeMap := make(map[string]*domain.LocationType, len(existingLocTypes))
	if len(dbLocationTypes) > 1 {
		namesMap := make(map[string]bool)
		for _, v := range payload.LocationTypes {
			namesMap[v.Name] = true
		}
		for _, v := range dbLocationTypes {
			if _, exists := namesMap[v.Name.String]; !exists && v.Name.String != "org" {
				return utils.NewError("mustImportAllExistData",
					fmt.Errorf("invalid data. please make sure to import all existing data"))
			}
		}
		if existingLocTypes[len(existingLocTypes)-1].Name != payload.LocationTypes[len(payload.LocationTypes)-1].Name {
			return utils.NewError("canNotUpdateLowestType",
				fmt.Errorf("cant add lowest level: %d", payload.LocationTypes[len(payload.LocationTypes)-1].Level))
		}
	}
	// if exists then get the existing id, if not, generate an id
	for _, v := range existingLocTypes {
		existingLocTypeMap[v.LocationTypeID] = v
	}
	for _, l := range payload.LocationTypes {
		lt, ok := existingLocTypeMap[l.LocationTypeID]
		if ok {
			l.LocationTypeID = lt.LocationTypeID
		} else if l.LocationTypeID == "" {
			l.LocationTypeID = idutil.ULIDNow()
		}
	}

	err = database.ExecInTx(ctx, l.DB, func(ctx context.Context, tx pgx.Tx) error {
		err = l.LocationTypeRepo.Import(ctx, l.DB, payload.LocationTypes)
		return err
	})

	if err != nil {
		return utils.NewSystemError(fmt.Errorf("LocationTypeRepo.Import: %w", err))
	}
	return nil
}
