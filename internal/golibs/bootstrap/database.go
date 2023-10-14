package bootstrap

import (
	"context"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

func initDatabase(ctx context.Context, c interface{}, rsc *Resources) error {
	pgconf, err := extract[configs.PostgresConfigV2](c, postgresV2FieldName)
	if err != nil {
		return ignoreErrFieldNotFound(err)
	}

	_ = rsc.WithDatabaseC(ctx, pgconf.Databases)
	return nil
}

// Databaser handles initializing connections to databases.
type Databaser interface {
	// ConnectV2 is similar to Connect, but uses the new postgres configuration
	// that allows https://github.com/GoogleCloudPlatform/cloud-sql-go-connector.
	//
	// It also returns a clean up function that must be invoked when the program finishes
	// to clean up underlying resources.
	ConnectV2(ctx context.Context, l *zap.Logger, cfg configs.PostgresDatabaseConfig) (*pgxpool.Pool, func() error, error)
}

// database implements databaser interface using the canonical function database.NewConnectionPoolV2.
type databaseImpl struct{}

var _ Databaser = &databaseImpl{}

// newDatabaseImpl returns a default databaseImpl object.
// This function doesn't really do anything for now.
func newDatabaseImpl() *databaseImpl {
	return &databaseImpl{}
}

func (d *databaseImpl) ConnectV2(ctx context.Context, l *zap.Logger, cfg configs.PostgresDatabaseConfig) (*pgxpool.Pool, func() error, error) {
	return database.NewPool(ctx, l, cfg)
}
