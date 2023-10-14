package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/manabie-com/backend/internal/golibs/configs"

	"cloud.google.com/go/cloudsqlconn/postgres/pgxv4"
	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx"
	"go.uber.org/zap"
)

type PostgresMigrateConfig struct {
	Source   string                         `yaml:"source"`
	Database configs.PostgresDatabaseConfig `yaml:"database"`
}

// MigrationConfig is used to configure database migration jobs.
type MigrationConfig struct {
	Common          configs.CommonConfig  `yaml:"common"`
	PostgresMigrate PostgresMigrateConfig `yaml:"postgres_migrate"`
}

const (
	localSQLDriverName string = "pgx"
	cloudSQLDriverName string = "cloudsql-postgres"
)

func MigrateDatabase(ctx context.Context, c MigrationConfig, l *zap.Logger) error {
	// Prepare the necessary connection, dialer, driver
	targetSQLDriverName := localSQLDriverName
	if c.PostgresMigrate.Database.IsCloudSQL() {
		l.Info("target database is a Cloud SQL instance")
		// Register a custom driver for cloudsqlconn
		// See https://github.com/GoogleCloudPlatform/cloud-sql-go-connector#using-the-dialer-with-databasesql
		connopts, err := c.PostgresMigrate.Database.DefaultCloudSQLConnOpts(ctx)
		if err != nil {
			return fmt.Errorf("c.PostgresMigrate.Database.DefaultCloudSQLConnOpts: %s", err)
		}
		cleanup, err := pgxv4.RegisterDriver(cloudSQLDriverName, connopts...)
		if err != nil {
			return err
		}
		defer func() {
			l.Debug("running cleanup for cloudsqlconn")
			if err := cleanup(); err != nil {
				l.Warn("failed to clean up for cloudsqlconn", zap.Error(err))
			}
		}()
		targetSQLDriverName = cloudSQLDriverName
	} else {
		l.Info("target database is a normal PostgreSQL instance")
	}

	connstring, err := c.PostgresMigrate.Database.ConnectionString()
	if err != nil {
		return err
	}
	db, err := sql.Open(targetSQLDriverName, connstring)
	if err != nil {
		return fmt.Errorf("sql.Open: %s", err)
	}
	defer func() {
		l.Debug("invoking sql.DB.Close()")
		if err := db.Close(); err != nil {
			l.Warn("sql.DB.Close() failed", zap.Error(err))
		}
	}()
	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		return fmt.Errorf("postgres.WithInstance: %s", err)
	}

	// Create a migration instance
	m, err := migrate.NewWithDatabaseInstance(c.PostgresMigrate.Source, c.PostgresMigrate.Database.DBName, driver)
	if err != nil {
		return fmt.Errorf("migrate.NewWithDatabaseInstance: %s", err)
	}
	m.Log = newMigrateLogger(l)

	// Handle cancel signal
	signalCtx, signalCancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer signalCancel()
	go func() {
		<-signalCtx.Done()

		// send signal to stop migration
		// Note that it does not stop when an SQL statement is running.
		// In that case we must wait until that SQL statement is completed.
		m.GracefulStop <- true
	}()

	// Finally, run the migration
	err = m.Up()
	if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, migrate.ErrNilVersion) {
		m.Log.Printf("no migration")
		return nil
	}
	return err
}

// migrateLoggerAdapter implements migrate.Logger.
type migrateLoggerAdapter struct {
	zapl *zap.Logger
}

func newMigrateLogger(zapl *zap.Logger) *migrateLoggerAdapter {
	return &migrateLoggerAdapter{zapl: zapl}
}

func (l *migrateLoggerAdapter) Printf(format string, v ...interface{}) {
	l.zapl.Info(fmt.Sprintf(format, v...))
}

func (l *migrateLoggerAdapter) Verbose() bool {
	// // verbose == true when app log level < INFO
	// // verbose == false when app log level >= INFO
	// return l.l.Core().Enabled(zapcore.InfoLevel)
	return true // always return true for now
}
