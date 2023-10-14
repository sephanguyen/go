package syllabus

import (
	"context"

	"github.com/manabie-com/backend/j4/infras"
	j4 "github.com/manabie-com/j4/pkg/runner"
	"google.golang.org/grpc/metadata"
)

func ScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	scenarios := []*j4.Scenario{}

	sc, err := GenRetrieveCourseStatisticScenario(ctx, c, dep)
	if err != nil {
		return nil, err
	}
	sc2, err := GenRetrieveCourseStatisticV2Scenario(ctx, c, dep)
	if err != nil {
		return nil, err
	}
	scenarios = append(scenarios, sc)
	scenarios = append(scenarios, sc2)
	return scenarios, nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}
