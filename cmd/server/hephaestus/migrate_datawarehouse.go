package hephaestus

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/hephaestus/configurations"
)

var dataWarehouseMigratePath string
var dataWarehouseName string

func MigrateDataWarehouse(ctx context.Context, cfg configurations.MigrateConfig, rsc *bootstrap.Resources) error {
	c := database.MigrationConfig{
		PostgresMigrate: database.PostgresMigrateConfig{
			Source:   dataWarehouseMigratePath,
			Database: cfg.DataWarehouses.Databases[dataWarehouseName],
		},
	}

	return database.MigrateDatabase(ctx, c, rsc.Logger())
}
