package communication

import (
	"context"

	"github.com/manabie-com/backend/features/common"
	"github.com/manabie-com/backend/features/communication/stress"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"

	"github.com/cucumber/godog"
)

type StressTestSc struct {
	defaultloc string
	*common.Suite
	*common.Connections
	*helper.CommunicationHelper
}

func (s *StressTestSc) newSchoolAdminToken(ctx context.Context) (context.Context, error) {
	school, loc, _, err := s.Suite.NewOrgWithOrgLocation(ctx)
	if err != nil {
		return ctx, err
	}
	s.defaultloc = loc
	ctx = contextWithResourcePath(ctx, i32ToStr(school))
	ctx, err = s.Suite.ASignedInWithSchool(ctx, "school admin", school)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
func (s *StressTestSc) runWithConversationsWithMemberEach(ctx context.Context, totalConv int, memPerConv int) (context.Context, error) {
	stress := stress.NewStressChat(s.Suite, s.Connections, s.CommunicationHelper, s.defaultloc, totalConv, memPerConv)
	return stress.Run(ctx)
}

func (c *SuiteConstructor) InitStressTestChat(dep *Dependency, godogCtx *godog.ScenarioContext) {
	suite := StressTestSc{
		Suite:               dep.commonSuite,
		Connections:         dep.conns,
		CommunicationHelper: dep.helper,
	}
	stepsMapping := map[string]interface{}{
		`^new school admin token$`:                                       suite.newSchoolAdminToken,
		`^run with "([^"]*)" conversations with "([^"]*)" members each$`: suite.runWithConversationsWithMemberEach,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}
