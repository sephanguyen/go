package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/manabie-com/backend/internal/golibs/configs"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/stretchr/testify/require"
)

// TestE2ENewPool checks a basic connection against
// an actual Cloud SQL instance
func TestE2ENewPool(t *testing.T) {
	if !enabled("TEST_CONNECT_DATABASE") {
		t.Skipf(`test is disabled (to activate it, run "TEST_CONNECT_DATABASE=true go test -v -run ^TestE2ENewConnectionPoolDemo$")`)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	l := logger.NewZapLogger("debug", false)
	c, err := defaultTestPostgresConfig()
	require.NoError(t, err)
	pool, dbcancel, err := NewPool(ctx, l, *c)
	require.NoError(t, err)
	defer dbcancel()

	// do something with the pool
	require := require.New(t)
	t.Logf("trying to list all databases in this postgresql instance")
	rows, err := pool.Query(ctx, "SELECT datname FROM pg_catalog.pg_database")
	require.NoError(err)
	defer rows.Close()
	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		require.NoError(err)
		t.Logf("row: %s", s)
	}
	require.NoError(rows.Err())
}

func defaultTestPostgresConfig() (*configs.PostgresDatabaseConfig, error) {
	c := &configs.PostgresDatabaseConfig{}
	c.CloudSQLInstance = os.Getenv("CLOUDSQL_INSTANCE")
	if c.CloudSQLInstance == "" {
		return nil, &missingEnvErr{attr: "Cloud SQL instance", env: "CLOUDSQL_INSTANCE"}
	}
	c.CloudSQLUsePublicIP = true // most likely be true, since we are outside of the VPC
	c.CloudSQLAutoIAMAuthN = true
	c.User = os.Getenv("DBUSER")
	if c.User == "" {
		return nil, &missingEnvErr{attr: "database user", env: "DBUSER"}
	}
	c.Password = os.Getenv("DBPASSWORD") // password is optional
	c.Host = os.Getenv("DBHOST")
	c.Port = os.Getenv("DBPORT")
	c.DBName = os.Getenv("DBNAME")
	if c.DBName == "" {
		return nil, &missingEnvErr{attr: "database name", env: "DBNAME"}
	}

	c.MaxConns = 2
	return c, nil
}

type missingEnvErr struct {
	attr string
	env  string
}

func (e *missingEnvErr) Error() string {
	return fmt.Sprintf("%s, %q env var, must be specified", e.attr, e.env)
}

func enabled(flag string) bool {
	flagval := os.Getenv(flag)
	return flagval == "true" || flagval == "t" || flagval == "1"
}
