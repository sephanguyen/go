package hephaestus

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/hephaestus/configurations"

	"go.uber.org/multierr"
)

var DWHResourcePath string

const (
	localSQLDriverName string = "pgx"
	cloudSQLDriverName string = "cloudsql-postgres"
)

func initDB(cfg configs.PostgresDatabaseConfig) (*sql.DB, error) {
	targetSQLDriverName := localSQLDriverName
	if cfg.IsCloudSQL() {
		targetSQLDriverName = cloudSQLDriverName
	}

	connectionString, err := cfg.ConnectionString()
	if err != nil {
		return nil, fmt.Errorf("get connection string error: %s", err)
	}

	dlConn, err := sql.Open(targetSQLDriverName, connectionString)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %s", err)
	}

	return dlConn, nil
}

func counterSchedulerTable(dlConn, dwhConn *sql.DB) error {
	queryDL := "select count(*) as total from bob.scheduler where resource_path=$1"
	queryDWH := "select count(*) as total from bob.scheduler_public_info"
	var totalDWHRecords string
	if err := dlConn.QueryRow(queryDL, DWHResourcePath).Scan(&totalDWHRecords); err != nil {
		return err
	}
	var totalDLRecords string
	if err := dwhConn.QueryRow(queryDWH).Scan(&totalDLRecords); err != nil {
		return err
	}

	if totalDWHRecords != totalDLRecords {
		return fmt.Errorf("data is not correct on table scheduler and scheduler_public_info  %s/%s", totalDWHRecords, totalDLRecords)
	}
	zapLogger.Info(fmt.Sprintf("Verify table scheduler and scheduler_public_info %s/%s", totalDLRecords, totalDWHRecords))

	return nil
}

func RunAccuracyDWH(_ context.Context, c configurations.MigrateConfig, _ *bootstrap.Resources) error {
	zapLogger = logger.NewZapLogger("debug", c.Common.Environment == LocalEnv)
	cfgDL := c.DataLake.Databases["alloydb"]
	cfgDWH := c.DataWarehouses.Databases["kec"]

	dlConn, err := initDB(cfgDL)
	if err != nil {
		return fmt.Errorf("initDB: %s", err)
	}
	defer dlConn.Close()
	zapLogger.Info("connected DL")

	dwhConn, err := initDB(cfgDWH)
	if err != nil {
		return fmt.Errorf("initDB: %s", err)
	}
	defer dwhConn.Close()
	zapLogger.Info("connected DWH")

	return multierr.Combine(
		counterSchedulerTable(dlConn, dwhConn),
	)
}
