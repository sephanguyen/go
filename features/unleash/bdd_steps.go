package unleash

import (
	"regexp"
	"sync"

	"github.com/manabie-com/backend/features/helper"

	"github.com/cucumber/godog"
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{
		`^send request get unleash config$`:               s.sendRequest,
		`^the request to check the unleash health$`:       s.theRequestToCheckUnleashHealth,
		`^send request to check health$`:                  s.sendRequest,
		`^unleash must return healthy status$`:            s.unleashMustReturnHealthStatus,
		`^the request to check the unleash-proxy health$`: s.theRequestToCheckUnleashProxyHealth,
		`^unleash-proxy must return health status$`:       s.unleashProxyMustReturnHealthStatus,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
