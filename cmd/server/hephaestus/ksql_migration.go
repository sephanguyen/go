package hephaestus

import (
	"context"

	"github.com/manabie-com/backend/cmd/server/hephaestus/cmd/server/hephaestus/ksql/ksqlCmd"
	"github.com/manabie-com/backend/internal/golibs/bootstrap"
	"github.com/manabie-com/backend/internal/golibs/logger"
	"github.com/manabie-com/backend/internal/hephaestus/configurations"
)

func MigrateKsql(ctx context.Context, c configurations.Config, rsc *bootstrap.Resources) error {
	zapLogger = logger.NewZapLogger("debug", c.Common.Environment == LocalEnv)
	return ksqlCmd.UpKsql("", "", c.Kafka.KsqlAddr)
}
