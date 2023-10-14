package conversationmgmt

import (
	"github.com/manabie-com/backend/features/conversationmgmt/common"

	"github.com/cucumber/godog"
)

type TestSuite struct {
	*common.ConversationMgmtSuite
}

func (c *SuiteConstructor) InitTest(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &TestSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates "([^"]*)" students with "([^"]*)" parents info for each student$`:                                    s.CreatesNumberOfStudentsWithParentsInfo,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}
