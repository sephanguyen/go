package main

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/j4/infras"
	"github.com/manabie-com/backend/j4/scenarios/syllabus"

	"github.com/manabie-com/j4/pkg/instrument"
	j4 "github.com/manabie-com/j4/pkg/runner"
)

type ScenarioIntializer func(context.Context, *infras.ManabieJ4Config, *infras.Dep) ([]*j4.Scenario, error)

var scenarioRegistry = map[string]ScenarioIntializer{
	// "communication": communication.ScenarioIntializer,
	// "communication": communication.SystemNotificationScenarioIntializer,
	// "draft": draft.ScenarioIntializer, // remember to enable deploying draft in local first
	// "usermgmt":      usermgmt.ScenarioIntializer,
	// "syllabus":  syllabus.ScenarioIntializer,
	// "timesheet": timesheet.ScenarioIntializer,
	// "lesson": lesson.ScenarioIntializer,
	// "invoice":       invoice.ScenarioIntializer,
	// "payment":       payment.ScenarioIntializer,
	// "virtualclassroom": virtualclassroom.ScenarioIntializer,
	"syllabus": syllabus.ScenarioIntializer,
}

func runRegisteredScenarios(ctx context.Context, b *j4.Runner, cfg *infras.ManabieJ4Config, dep *infras.Dep) error {
	for name, initializer := range scenarioRegistry {
		scenarios, err := initializer(ctx, cfg, dep)
		if err != nil {
			return fmt.Errorf("init scenario for %s failed: %s", name, err)
		}
		for _, sc := range scenarios {
			err = b.RegisterScenario(sc)
			if err != nil {
				return fmt.Errorf("j4.Runner.AddScenario: %s", err)
			}
		}
	}
	return b.Run(ctx)
}

func Run(ctx context.Context, cfg *j4.Config, manabieCfg *infras.ManabieJ4Config) error {
	b, err := j4.NewRunner(*cfg)
	if err != nil {
		return fmt.Errorf("j4.NewRunner: %s", err)
	}
	fmt.Println("RQLite startup completed")
	infraDependencies := setupInfras(ctx, manabieCfg)

	return runRegisteredScenarios(ctx, b, manabieCfg, infraDependencies)
}

func setupInfras(ctx context.Context, c *infras.ManabieJ4Config) *infras.Dep {
	instrument.InitOC()
	conns := &infras.Connections{}
	// TODO: init all this stuff inside connections pkg, automatically map config with map item
	if err := conns.ConnectDB(ctx, c.PostgresV2); err != nil {
		panic(err)
	}

	if err := conns.ConnectGrpcPool(ctx, c.ClusterGrpcAddr, true, infras.TlsSkipVerifyDialingOpts); err != nil {
		panic(err)
	}
	if err := conns.ConnecGrpc(ctx, c.ShamirAddr, infras.InsecureDialingOpts); err != nil {
		panic(err)
	}
	if err := conns.ConnectHasuras(ctx, c); err != nil {
		panic(err)
	}
	if err := conns.ConnectKafka(c); err != nil {
		panic(err)
	}
	return &infras.Dep{Connections: conns}
}
