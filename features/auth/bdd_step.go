package auth

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
		// health check
		`^everything is OK$`:                                           s.everythingIsOK,
		`^health check endpoint called$`:                               s.healthCheckEndpointCalled,
		`^auth service should return "([^"]*)" with status "([^"]*)"$`: s.authServiceShouldReturnWithStatus,

		// exchange salesforce token
		`^a user signed in as "([^"]*)" in "([^"]*)" organization$`:            s.userSignedInAsInOrganization,
		`^user exchanges salesforce token$`:                                    s.userExchangesSalesforceToken,
		`^user exchanges salesforce token successfully$`:                       s.userExchangesSalesforceTokenSuccessfully,
		`^user can not exchanges salesforce token with status code "([^"]*)"$`: s.userCanNotExchangesSalesforceTokenWithStatusCode,
	}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})
	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
