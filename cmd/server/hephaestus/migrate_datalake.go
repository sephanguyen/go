package hephaestus

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/hephaestus/configurations"
)

var dataLakeMigratePath string
var dataLakeName string

func MigrateDataLake(ctx context.Context, cfg configurations.MigrateConfig, rsc *bootstrap.Resources) error {
	c := database.MigrationConfig{
		PostgresMigrate: database.PostgresMigrateConfig{
			Source:   dataLakeMigratePath,
			Database: cfg.DataLake.Databases[dataLakeName],
		},
	}

	return database.MigrateDatabase(ctx, c, rsc.Logger())
}
