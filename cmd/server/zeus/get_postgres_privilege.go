package zeus

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/zeus/configurations"
	"github.com/manabie-com/backend/internal/zeus/repositories"

	"go.uber.org/zap"
)

var serviceName string

func init() {
	bootstrap.RegisterJob("get_postgres_privileges", getPostgresPrivileges).
		Desc("get Postgres privilege info for specific service").
		StringVar(&serviceName, "serviceName", "", "service to get postgres privileges for")
}

func getPostgresPrivileges(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger := rsc.Logger()
	db := rsc.DB()
	repo := &repositories.PostgresNamespaceRepo{}
	postgresNamespaces, err := repo.Get(ctx, db)
	if err != nil {
		return fmt.Errorf("PostgresNamespaceRepo.Get() failed for service %s: %s", serviceName, err)
	}

	if len(postgresNamespaces) == 0 {
		return fmt.Errorf("length postgresNamespaces for service %s equals to 0", serviceName)
	}

	for _, namespace := range postgresNamespaces {
		accessPrivileges := database.FromTextArray(namespace.AccessPrivileges)
		key := fmt.Sprintf("%s_access_privileges", serviceName)
		zapLogger.Info("postgres privilege permissions", zap.Any(key, accessPrivileges))
	}
	return nil
}
