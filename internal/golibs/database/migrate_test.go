package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/stretchr/testify/require"
)

// TestE2EMigrateDatabaseLocal tests that migrateDatabase() can successfully perform migration
// for a database. Therefore, it requires an actual PostgreSQL database.
// Environment variables DBUSER, DBPASSWORD, DBHOST, DBPORT must be set properly to run this test.
//
// Currently, file:///migrations/<dbname> is chosen as the source.
func TestE2EMigrateDatabaseLocal(t *testing.T) {
	if !enabled("TEST_MIGRATE") {
		t.Skipf(`test is disabled (to activate it, run "TEST_MIGRATE=true go test -run ^TestE2EMigrateDatabaseLocal$")`)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	dbConfig := defaultLocalDBConfig()
	c := MigrationConfig{
		PostgresMigrate: PostgresMigrateConfig{
			Source:   "file://" + execwrapper.Abs("migrations", dbConfig.DBName),
			Database: dbConfig,
		},
	}

	l := logger.NewZapLogger("debug", false)
	err := MigrateDatabase(ctx, c, l)
	require.NoError(t, err)
}

func defaultLocalDBConfig() configs.PostgresDatabaseConfig {
	c := configs.PostgresDatabaseConfig{}
	c.User = os.Getenv("DBUSER")
	if c.User == "" {
		c.User = "postgres"
	}
	c.Password = os.Getenv("DBPASSWORD") // don't automatically set password, as a safeguard
	c.Host = os.Getenv("DBHOST")
	if c.Host == "" {
		c.Host = "localhost"
	}
	c.Port = os.Getenv("DBPORT")
	if c.Port == "" {
		c.Port = "5432"
	}
	c.DBName = os.Getenv("DBNAME")
	if c.DBName == "" {
		c.DBName = "bob"
	}
	return c
}

// TestE2EMigrateDatabaseCloudSQL is similar to TestE2EMigrateDatabaseCloudSQL, but for
// a Cloud SQL database instead.
//
// Currently, file:///migrations/zeus is chosen as the source, since it's the least harmful.
func TestE2EMigrateDatabaseCloudSQL(t *testing.T) {
	if !enabled("TEST_MIGRATE") {
		t.Skipf(`test is disabled (to activate it, run "TEST_MIGRATE=true go test -run ^TestE2EMigrateDatabaseCloudSQL$")`)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	dbConfig, err := defaultCloudSQLConfig()
	require.NoError(t, err)
	c := MigrationConfig{
		PostgresMigrate: PostgresMigrateConfig{
			Source:   "file://" + execwrapper.Abs("migrations/zeus"),
			Database: *dbConfig,
		},
	}

	l := logger.NewZapLogger("debug", false)

	err = MigrateDatabase(ctx, c, l)
	require.NoError(t, err)
}

func defaultCloudSQLConfig() (*configs.PostgresDatabaseConfig, error) {
	c := &configs.PostgresDatabaseConfig{}
	c.CloudSQLInstance = os.Getenv("CLOUDSQL_INSTANCE")
	if c.CloudSQLInstance == "" {
		return nil, fmt.Errorf(`Cloud SQL instance name, "CLOUDSQL_INSTANCE" env var, must be specified`)
	}
	// we are likely not running inside staging's VPC, thus must use public IP
	c.CloudSQLUsePublicIP = true
	c.CloudSQLAutoIAMAuthN = true
	c.CloudSQLImpersonateServiceAccountEmail = "" // add email if you want to enable impersonation

	c.User = os.Getenv("DBUSER")
	if c.User == "" {
		return nil, fmt.Errorf(`database user, "DBUSER" env var, must be specified`)
	}
	c.Password = os.Getenv("DBPASSWORD") // even with Cloud SQL, we must support password auth
	c.DBName = os.Getenv("DBNAME")
	if c.DBName == "" {
		return nil, fmt.Errorf(`database name, "DBNAME" env var, must be specified`)
	}
	return c, nil
}
