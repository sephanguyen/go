package platform

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/logger"

	"github.com/cucumber/godog"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func init() {
	common.RegisterTest("platform", &common.SuiteBuilder[common.Config]{
		SuiteInitFunc:    testSuiteInitializer,
		ScenarioInitFunc: scenarioInitializer,
	})
}

type state struct {
	org string

	// network_policy.feature
	networkPolicyEnabled bool
}

type stateCtxKey struct{}

func stepStateFromContext(ctx context.Context) *state {
	return ctx.Value(stateCtxKey{}).(*state)
}

func stepStateToContext(ctx context.Context, state *state) context.Context {
	return context.WithValue(ctx, stateCtxKey{}, state)
}

func testSuiteInitializer(_ *common.Config, _ common.RunTimeFlag) func(ctx *godog.TestSuiteContext) {
	return func(ctx *godog.TestSuiteContext) {}
}

func scenarioInitializer(c *common.Config, _ common.RunTimeFlag) func(ctx *godog.ScenarioContext) {
	return func(ctx *godog.ScenarioContext) {
		ctx.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
			l := logger.NewZapLogger(c.Common.Log.ApplicationLevel, true, zap.AddCallerSkip(1))
			ctx = ctxzap.ToContext(ctx, l)

			state := &state{
				org: c.Common.Organization,
			}
			return stepStateToContext(ctx, state), nil
		})
		initSteps(ctx)
	}
}

var initSteps = func() func(ctx *godog.ScenarioContext) {
	// add your step defintions here
	steps := map[string]interface{}{
		`^network policy is enabled in current kubernetes cluster$`:                networkPolicyIsEnabledInCurrentKubernetesCluster,
		`^pod in "([^"]*)" namespace "([^"]*)" access pod in "([^"]*)" namespace$`: podInNamespaceAccessPodInNamespace,
	}

	regexpMap := helper.BuildRegexpMapV2(steps)
	return func(ctx *godog.ScenarioContext) {
		for k, v := range steps {
			ctx.Step(regexpMap[k], v)
		}
	}
}()

// k8sClient is used to return a global kubernetes.Clientset.
var k8sClient = func() func() *kubernetes.Clientset {
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(fmt.Errorf("failed to create in-cluster config: %s", err))
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(fmt.Errorf("failed to create k8s client: %s", err))
	}
	return func() *kubernetes.Clientset {
		return clientset
	}
}()
