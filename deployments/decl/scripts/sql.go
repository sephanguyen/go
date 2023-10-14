package main

import (
	"path/filepath"

	"github.com/manabie-com/backend/deployments/decl/scripts/automation"
	"github.com/manabie-com/backend/internal/golibs/execwrapper"
)

func main() {
	source := filepath.Join(execwrapper.RootDirectory(), "deployments/decl/stag-defs.yaml")
	dest := filepath.Join(execwrapper.RootDirectory(), "deployments/helm/emulators/infras/migrations")
	err := automation.NewSQL().
		From(source).
		To(dest).
		Customize("0.init.sql",
			`CREATE DATABASE "alloydb";`,
			`CREATE DATABASE "mlflow";`,
			`CREATE DATABASE "unleashv2";`,
			`CREATE USER "kafka_connector" REPLICATION PASSWORD 'example';`,
			`GRANT CREATE, CONNECT, TEMPORARY, TEMP ON DATABASE "unleash" TO unleash;`,
			`GRANT CREATE, CONNECT, TEMPORARY, TEMP ON DATABASE "unleashv2" TO unleashv2;`,
		).
		Customize("bob.sql",
			"ALTER DEFAULT PRIVILEGES FOR ROLE postgres IN SCHEMA public GRANT SELECT ON TABLES TO shamir;",
			"ALTER ROLE shamir BYPASSRLS;",
			"ALTER ROLE hasura BYPASSRLS;",
		).
		Run()
	if err != nil {
		panic(err)
	}
}
