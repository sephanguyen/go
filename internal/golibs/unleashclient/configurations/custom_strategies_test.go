package configurations

import (
	"testing"

	"github.com/Unleash/unleash-client-go/v3/context"
	"github.com/stretchr/testify/assert"
)

func TestOrgStrategyGetName(t *testing.T) {
	t.Parallel()
	orgStrategy := OrgStrategy{}
	strategyName := orgStrategy.Name()
	assert.Equal(t, strategyName, "strategy_organization")
}

func TestOrgStrategyIsEnabled(t *testing.T) {
	t.Parallel()
	orgStrategy := OrgStrategy{}
	var unleashCtx context.Context
	params := make(map[string]interface{})
	params["organization"] = "test_invalid_org"

	result := orgStrategy.IsEnabled(params, &unleashCtx)
	assert.Equal(t, result, false)
}

func TestEnvStrategyGetName(t *testing.T) {
	t.Parallel()
	envStrategy := EnvStrategy{}
	strategyName := envStrategy.Name()
	assert.Equal(t, strategyName, "strategy_environment")
}

func TestEnvStrategyIsEnabled(t *testing.T) {
	t.Parallel()
	envStrategy := EnvStrategy{}
	var unleashCtx context.Context
	params := make(map[string]interface{})
	params["environments"] = "test_invalid_env"

	assert.Equal(t, envStrategy.IsEnabled(params, &unleashCtx), false)
}

func TestVariantStrategyGetName(t *testing.T) {
	t.Parallel()
	variantStrategy := VariantStrategy{}
	strategyName := variantStrategy.Name()
	assert.Equal(t, strategyName, "strategy_variant")
}

func TestVariantStrategyIsEnabled(t *testing.T) {
	t.Parallel()
	variantStrategy := VariantStrategy{}
	var unleashCtx context.Context
	params := make(map[string]interface{})
	params["variants"] = "test_invalid_variant"

	assert.Equal(t, variantStrategy.IsEnabled(params, &unleashCtx), false)
}
