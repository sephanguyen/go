package zeus

import (
	"regexp"
	"sync"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/helper"
	"github.com/manabie-com/backend/internal/golibs/database"

	"github.com/cucumber/godog"
	"go.uber.org/zap"
)

var (
	buildRegexpMapOnce sync.Once
	regexpMap          map[string]*regexp.Regexp
)

func init() {
	common.RegisterTestWithCommonConnection("zeus", func(c *common.Config, dep *common.Connections) func(*godog.ScenarioContext) {
		return func(ctx *godog.ScenarioContext) {
			s := suite{dep: dep}
			s.database = dep.ZeusDB
			s.ZapLogger = dep.Logger
			s.configs = c
			initSteps(ctx, &s)
		}
	})
}

type suite struct {
	dep       *common.Connections
	configs   *common.Config
	database  database.Ext
	ZapLogger *zap.Logger
}

func initSteps(ctx *godog.ScenarioContext, s *suite) {
	steps := map[string]interface{}{}

	buildRegexpMapOnce.Do(func() {
		regexpMap = helper.BuildRegexpMapV2(steps)
	})

	for k, v := range steps {
		ctx.Step(regexpMap[k], v)
	}
}
