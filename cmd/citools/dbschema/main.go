package main

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/spf13/cobra"
)

var allServicesWithDB string

func main() {
	if err := newGenDBSchemaCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func genDBSchema(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*len(allServicesWithDB))*time.Second)
	defer cancel()

	sevices := strings.Split(allServicesWithDB, ",")
	for _, srv := range sevices {
		sr := database.NewSchemaRecorder(srv)
		if err := sr.Record(ctx); err != nil {
			return err
		}
	}

	return nil
}

// newGenDBSchemaCmd returns a command that migrates up a database from the migration files
// and saves all the schema as JSON.
func newGenDBSchemaCmd() *cobra.Command {
	command := &cobra.Command{
		Use:   "mock_gendbschema",
		Short: "generate database schema for all services",
		Args:  cobra.NoArgs,
		RunE:  genDBSchema,
	}

	command.PersistentFlags().StringVar(&allServicesWithDB, "services", "", "")

	return command
}
