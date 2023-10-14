package main

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
)

func init() {
	bootstrap.RegisterJob("sql_migrate", migrateDatabase)
}

func migrateDatabase(ctx context.Context, c database.MigrationConfig, rsc *bootstrap.Resources) error {
	return database.MigrateDatabase(ctx, c, rsc.Logger())
}
