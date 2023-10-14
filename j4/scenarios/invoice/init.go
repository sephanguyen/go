package invoice

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	j4 "github.com/manabie-com/j4/pkg/runner"
)

func ScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	return serviceutil.GenHasuraScenarios(ctx, c, dep, "invoicemgmt", hasuraQueries)
}
