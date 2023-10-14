package configurations

import (
	"strings"

	"github.com/Unleash/unleash-client-go/v3/context"
)

type EnvStrategy struct{}
type OrgStrategy struct{}

// Discussion: https://manabie.slack.com/archives/C024PDK329F/p1652773011998579
// Ticket: https://manabie.atlassian.net/browse/LT-15462
type VariantStrategy struct{}

func (s EnvStrategy) Name() string {
	return "strategy_environment"
}

func (s EnvStrategy) IsEnabled(params map[string]interface{}, ctx *context.Context) bool {
	if ctx == nil {
		return false
	}
	value, found := params["environments"]
	if !found {
		return false
	}

	environments, ok := value.(string)
	if !ok {
		return false
	}

	for _, e := range strings.Split(environments, ",") {
		if e == ctx.Properties["env"] {
			return true
		}
	}

	return false
}

func (s OrgStrategy) Name() string {
	return "strategy_organization"
}

func (s OrgStrategy) IsEnabled(params map[string]interface{}, ctx *context.Context) bool {
	if ctx == nil {
		return false
	}
	value, found := params["organizations"]
	if !found {
		return false
	}

	organizations, ok := value.(string)
	if !ok {
		return false
	}

	for _, e := range strings.Split(organizations, ",") {
		if e == ctx.Properties["org"] {
			return true
		}
	}

	return false
}

func (s VariantStrategy) Name() string {
	return "strategy_variant"
}

func (s VariantStrategy) IsEnabled(params map[string]interface{}, ctx *context.Context) bool {
	if ctx == nil {
		return false
	}
	value, found := params["variants"]
	if !found {
		return false
	}

	variants, ok := value.(string)
	if !ok {
		return false
	}

	for _, e := range strings.Split(variants, ",") {
		if e == ctx.Properties["variants"] {
			return true
		}
	}

	return false
}
