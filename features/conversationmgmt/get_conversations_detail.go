package conversationmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/conversationmgmt/common"
	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/cucumber/godog"
)

type GetConversationsDetailSuite struct {
	*common.ConversationMgmtSuite
}

func (c *SuiteConstructor) InitGetConversationsDetail(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &GetConversationsDetailSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates a students with first name is "([^"]*)" and last name is "([^"]*)"$`:                                 s.CreatesAStudentWithFirstNameAndLastName,
		`^waiting for Agora User has been created$`:                                                                                 s.WaitingForAgoraUserHasBeenCreated,
		`^a new staff with role teacher is created$`:                                                                                s.StaffWithRoleTeacher,
		`^current staff create "([^"]*)" conversations for students$`:                                                               s.CurrentStaffCreateCreateNumberOfConversationsForStudents,
		`^current student get conversations detail$`:                                                                                s.currentStudentGetConversations,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *GetConversationsDetailSuite) currentStudentGetConversations(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	expectedConversations := commonState.Conversations
	student := commonState.Students[0]
	mapExpectedConversations := make(map[string]*entities.Conversation, len(expectedConversations))
	for _, expectedConversation := range expectedConversations {
		mapExpectedConversations[expectedConversation.ID] = expectedConversation
	}

	conversationIDs := make([]string, 0)
	for _, conversation := range expectedConversations {
		conversationIDs = append(conversationIDs, conversation.ID)
	}

	ctx = common.ContextWithToken(ctx, commonState.Students[0].Token)

	resp, err := cpb.NewConversationReaderServiceClient(s.ConversationMgmtGRPCConn).
		GetConversationsDetail(ctx, &cpb.GetConversationsDetailRequest{ConversationIds: conversationIDs})
	if err != nil {
		return ctx, fmt.Errorf("s.GetConversationsDetail: %v", err)
	}

	var vendorUserID string
	query := `
		SELECT agora_user_id FROM agora_user au 
		WHERE user_id = $1
	`
	row := s.TomDBConn.QueryRow(ctx, query, student.ID)
	if err := row.Scan(&vendorUserID); err != nil {
		return ctx, fmt.Errorf("error finding agora_user_id in agora_user with user_id: %s: %w", student.ID, err)
	}

	actualConversations := resp.Conversations
	for i := 0; i < len(actualConversations); i++ {
		if expect, exist := mapExpectedConversations[actualConversations[i].ConversationId]; exist {
			if expect.Name != actualConversations[i].Name {
				return ctx, fmt.Errorf("error Conversation stored incorrect Name data. conversation ID: %+v", expect.ID)
			}
			for _, member := range actualConversations[i].Members {
				if member.ConversationId != actualConversations[i].ConversationId {
					return ctx, fmt.Errorf("error Conversation Member stored incorrect ConversationID data. conversation ID: %+v", expect.ID)
				}
				if member.User.UserId != student.ID {
					return ctx, fmt.Errorf("error Conversation Member stored incorrect MemberID data. conversation ID: %+v", expect.ID)
				}
				if member.User.VendorUserId != vendorUserID {
					return ctx, fmt.Errorf("error Conversation Member stored incorrect VendorUserId data. conversation ID: %+v", expect.ID)
				}
			}
		}
	}
	return ctx, nil
}
