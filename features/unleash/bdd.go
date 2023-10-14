package unleash

import (
	"context"

	"github.com/cucumber/godog"
	"github.com/manabie-com/backend/features/common"

	"go.uber.org/zap"
)

func init() {
	common.RegisterTestWithCommonConnection("unleash", ScenarioInitializer)
}

func StepStateFromContext(ctx context.Context) *common.StepState {
	return ctx.Value(common.StepStateKey{}).(*common.StepState)
}

func StepStateToContext(ctx context.Context, state *common.StepState) context.Context {
	return context.WithValue(ctx, common.StepStateKey{}, state)
}

func ScenarioInitializer(c *common.Config, deps *common.Connections) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		s := newSuite(c, deps)
		initSteps(ctx, s)

		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			return StepStateToContext(ctx, s.StepState), nil
		})

		ctx.After(func(ctx context.Context, sc *godog.Scenario, err error) (context.Context, error) {
			stepState := StepStateFromContext(ctx)
			for _, v := range stepState.Subs {
				if v.IsValid() {
					err := v.Drain()
					if err != nil {
						return nil, err
					}
				}
			}
			return ctx, nil
		})
	}
}

type suite struct {
	*common.Connections
	*common.StepState
	ZapLogger               *zap.Logger
	Cfg                     *common.Config
	CommonSuite             *common.Suite
	ApplicantID             string
	UnleashSrvAddr          string
	UnleashAPIKey           string
	UnleashLocalAdminAPIKey string
}

func newSuite(c *common.Config, conns *common.Connections) *suite {
	s := &suite{
		Connections: conns,
		Cfg:         c,
		ZapLogger:   conns.Logger,
		ApplicantID: conns.ApplicantID,
		CommonSuite: &common.Suite{},
	}

	// s.CommonSuite.Connections = s.Connections
	s.CommonSuite.StepState = &common.StepState{}
	s.StepState = s.CommonSuite.StepState

	s.CommonSuite.StepState.FirebaseAddress = conns.FirebaseAddr
	s.CommonSuite.StepState.ApplicantID = conns.ApplicantID

	// Unleash cfg
	s.UnleashSrvAddr = c.UnleashSrvAddr
	s.UnleashAPIKey = c.UnleashAPIKey
	s.UnleashLocalAdminAPIKey = c.UnleashLocalAdminAPIKey

	return s
}

type Suite struct {
	suite
}
