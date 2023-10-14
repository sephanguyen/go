package serviceutil

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/j4/infras"
	j4 "github.com/manabie-com/j4/pkg/runner"
)

type HasuraQuery struct {
	Name  string
	Query string
	// Let you assign variable for querying variable dynamically, for example from-date value depends on current time
	VariablesCreator func() map[string]interface{}
}

func GenHasuraScenarios(ctx context.Context,
	cfg *infras.ManabieJ4Config,
	dep *infras.Dep,
	hasuraName string,
	qs []HasuraQuery) ([]*j4.Scenario, error) {
	tokenGenerator := NewTokenGenerator(cfg, dep.Connections)

	scenarios := []*j4.Scenario{}
	for _, item := range qs {
		scenarioConf, err := cfg.GetScenarioConfig(item.Name)
		if err != nil {
			return nil, err
		}
		scenarioOpt := infras.MustOptionFromConfig(&scenarioConf)
		name := item.Name
		query := item.Query
		itemVar := item.VariablesCreator

		scenarioOpt.TestClosure = func(parCtx context.Context) j4.TestFunc {
			// the testclojure is called at any time, need to ensure token is up to date
			tok, err := tokenGenerator.GetTokenFromShamir(parCtx, cfg.AdminID, cfg.SchoolID)
			if err != nil {
				panic(fmt.Errorf("failed to create test func %s", err))
			}
			return func(childCtx context.Context) error {
				_, err := dep.GetHasura(hasuraName).QueryRawHasuraV1(childCtx, tok, name, query, itemVar())
				if err != nil {
					return err
				}
				return nil
			}
		}
		sc, err := j4.NewScenario(name, *scenarioOpt)
		if err != nil {
			return nil, err
		}
		scenarios = append(scenarios, sc)
	}
	return scenarios, nil
}
