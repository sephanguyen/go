package conversationmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/conversationmgmt/common"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/cucumber/godog"
)

type UpdateConversationDetailSuite struct {
	*common.ConversationMgmtSuite
}

func (c *SuiteConstructor) InitUpdateConversationInfo(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &UpdateConversationDetailSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates a students with first name is "([^"]*)" and last name is "([^"]*)"$`:                                 s.CreatesAStudentWithFirstNameAndLastName,
		`^a new staff with role teacher is created$`:                                                                                s.StaffWithRoleTeacher,
		`^waiting for Agora User has been created$`:                                                                                 s.WaitingForAgoraUserHasBeenCreated,
		`^current staff create "([^"]*)" conversations for students$`:                                                               s.CurrentStaffCreateCreateNumberOfConversationsForStudents,
		`^student update conversation info$`:                                                                                        s.studentUpdateConversationInfo,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *UpdateConversationDetailSuite) studentUpdateConversationInfo(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	req := &cpb.UpdateConversationInfoRequest{
		ConversationId: commonState.Conversations[0].ID,
		Name:           idutil.ULIDNow(),
		OptionalConfig: []byte(`{
			"test_field": "updated_test_value"
		}`),
	}

	_, err := cpb.NewConversationModifierServiceClient(s.ConversationMgmtGRPCConn).
		UpdateConversationInfo(common.ContextWithToken(ctx, commonState.Students[0].Token), req)
	if err != nil {
		return ctx, fmt.Errorf("s.UpdateConversationInfo: %v", err)
	}

	query := `
		SELECT name FROM conversation c 
		WHERE conversation_id = $1
	`
	var conversationName string
	row := s.TomDBConn.QueryRow(ctx, query, req.ConversationId)
	if err := row.Scan(&conversationName); err != nil {
		return ctx, fmt.Errorf("error finding conversation name with conversationID: %s: %w", req.ConversationId, err)
	}

	if conversationName != req.Name {
		return ctx, fmt.Errorf("error update conversation with conversation ID: %+v", req.ConversationId)
	}

	return ctx, nil
}
