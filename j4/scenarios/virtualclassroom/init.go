package virtualclassroom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/serviceutil"

	j4 "github.com/manabie-com/j4/pkg/runner"

	"google.golang.org/grpc/metadata"
)

func ScenarioIntializer(ctx context.Context, c *infras.ManabieJ4Config, dep *infras.Dep) ([]*j4.Scenario, error) {
	tokenGen := serviceutil.NewTokenGenerator(c, dep.Connections)

	getLiveLessonScenario := &GetLiveLessonStateScenario{
		tokenGenerator: tokenGen,
		j4cfg:          c,
		conns:          dep.Connections,
	}
	oneLessonGetScenario, err := getLiveLessonScenario.getOneLessonTestScenario(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get test scenario 1 for get live lesson state: %w", err)
	}
	multiLessonGetScenario, err := getLiveLessonScenario.getMultipleLessonTestScenario(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get test scenario 2 for get live lesson state: %w", err)
	}

	modifyLiveLessonScenario := &ModifyLiveLessonStateScenario{
		tokenGenerator: tokenGen,
		j4cfg:          c,
		conns:          dep.Connections,
	}
	oneLessonModifyScenario, err := modifyLiveLessonScenario.getOneLessonTestScenario(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get test scenario 1 for modify live lesson state: %w", err)
	}
	multiLessonModifyScenario, err := modifyLiveLessonScenario.getMultipleLessonTestScenario(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get test scenario 1 for modify live lesson state: %w", err)
	}

	return []*j4.Scenario{
		oneLessonGetScenario,
		oneLessonModifyScenario,
		multiLessonGetScenario,
		multiLessonModifyScenario,
	}, nil
}

func contextWithValidVersion(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "pkg", "com.manabie.liz", "version", "1.0.0")
}

func contextWithToken(ctx context.Context, token string) context.Context {
	ctx = contextWithValidVersion(ctx)
	return metadata.AppendToOutgoingContext(contextWithValidVersion(ctx), "token", token)
}
