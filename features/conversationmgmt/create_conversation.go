package conversationmgmt

import (
	"context"

	"github.com/manabie-com/backend/features/conversationmgmt/common"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/cucumber/godog"
)

type CreateConversationSuite struct {
	*common.ConversationMgmtSuite
}

func (c *SuiteConstructor) InitCreateConversation(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &CreateConversationSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates a students with first name is "([^"]*)" and last name is "([^"]*)"$`:                                 s.CreatesAStudentWithFirstNameAndLastName,
		`^waiting for Agora User has been created$`:                                                                                 s.WaitingForAgoraUserHasBeenCreated,
		`^student create their conversation$`:                                                                                       s.studentCreateTheirConversation,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *CreateConversationSuite) studentCreateTheirConversation(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	memberIDs := make([]string, 0)
	for _, student := range commonState.Students {
		memberIDs = append(memberIDs, student.ID)
	}
	req := &cpb.CreateConversationRequest{
		Name:      "Test Conversation",
		MemberIds: memberIDs,
		OptionalConfig: []byte(`{
			"test_field": "test_value"
		}`),
	}

	ctx, cancel := common.ContextWithTokenAndTimeOut(ctx, commonState.Students[0].Token)
	defer cancel()
	_, err := cpb.NewConversationModifierServiceClient(s.ConversationMgmtGRPCConn).CreateConversation(ctx, req)
	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
