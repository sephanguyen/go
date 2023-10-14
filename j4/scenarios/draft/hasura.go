package draft

import (
	"context"
	"math/rand"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"
	j4 "github.com/manabie-com/j4/pkg/runner"
)

func ScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	return serviceutil.GenHasuraScenarios(ctx, c, dep, "draft", hasuraQueries)
}

var (
	hasuraQueries = []serviceutil.HasuraQuery{
		{
			Name:  "E2EInstanceSquadTagsList",
			Query: E2EInstanceSquadTagsList,
			VariablesCreator: func() map[string]interface{} {
				useCachedQuery := rand.Intn(10) < 5
				// reuse this query across requests to simulate caching
				if useCachedQuery {
					return map[string]interface{}{
						"limit":  10,
						"offset": 0,
					}
				}
				randLimit := rand.Int()
				return map[string]interface{}{
					"limit":  randLimit,
					"offset": 0,
				}
			},
		},
	}
	E2EInstanceSquadTagsList = `
          query E2EInstanceSquadTagsList ($offset: Int = 0, $limit: Int = 50, $squad_tag: String = "%%") {
            e2e_instances_squad_tags(offset: $offset, limit: $limit, where: {squad_tag:{_ilike:$squad_tag}}) {
              squad_tag
            }
          }`
)
