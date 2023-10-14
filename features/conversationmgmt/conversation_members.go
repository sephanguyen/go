package conversationmgmt

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/features/conversationmgmt/common"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/try"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/cucumber/godog"
	"k8s.io/utils/strings/slices"
)

type ConversationMembersSuite struct {
	*common.ConversationMgmtSuite
	createdConversationID string
	conversationMemberIDs []string
}

func (c *SuiteConstructor) InitConversationMembers(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &ConversationMembersSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates a students with first name is "([^"]*)" and last name is "([^"]*)"$`:                                 s.CreatesAStudentWithFirstNameAndLastName,
		`^a new staff with role teacher is created$`:                                                                                s.StaffWithRoleTeacher,
		`^student create their conversation$`:                                                                                       s.studentCreateTheirConversation,
		`^waiting for Agora User has been created$`:                                                                                 s.WaitingForAgoraUserHasBeenCreated,
		`^teacher is added to conversation$`:                                                                                        s.teacherIsAddedToConversation,
		`^"([^"]*)" is removed from conversation$`:                                                                                  s.isRemovedFromConversation,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *ConversationMembersSuite) studentCreateTheirConversation(ctx context.Context) (context.Context, error) {
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

	resp, err := cpb.NewConversationModifierServiceClient(s.ConversationMgmtGRPCConn).CreateConversation(common.ContextWithToken(ctx, commonState.Students[0].Token), req)
	if err != nil {
		return ctx, err
	}

	s.conversationMemberIDs = append(s.conversationMemberIDs, memberIDs...)
	s.createdConversationID = resp.ConversationId
	return common.StepStateToContext(ctx, commonState), nil
}

// TODO: move this func to common bdd files
func (s *ConversationMembersSuite) teacherIsAddedToConversation(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	ctxWithRp := common.ContextWithResourcePath(ctx, commonState.CurrentResourcePath)

	req := &cpb.AddConversationMembersRequest{
		ConversationId: s.createdConversationID,
		MemberIds:      []string{commonState.CurrentStaff.ID},
	}
	_, err := cpb.NewConversationModifierServiceClient(s.ConversationMgmtGRPCConn).AddConversationMembers(
		ctxWithRp,
		req,
	)
	if err != nil {
		return ctx, fmt.Errorf("failed AddConversationMembers: %+v", err)
	}
	s.conversationMemberIDs = append(s.conversationMemberIDs, commonState.CurrentStaff.ID)

	// check if conversation member data is correct
	query := `
		SELECT cm.user_id
		FROM conversation c
		JOIN conversation_member cm ON cm.conversation_id = c.conversation_id
		WHERE c.conversation_id = $1
		AND cm.status = 'CONVERSATION_MEMBER_STATUS_ACTIVE'
		AND cm.deleted_at IS NULL
		AND c.deleted_at IS NULL;
	`
	err = try.Do(func(attempt int) (bool, error) {
		actualMemberIDs := []string{}
		rows, err := s.TomDBConn.Query(ctx, query, database.Text(s.createdConversationID))
		if err != nil {
			return false, fmt.Errorf("failed Query: %+v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var userID string
			if err = rows.Scan(&userID); err != nil {
				return false, fmt.Errorf("failed Scan: %+v", err)
			}
			actualMemberIDs = append(actualMemberIDs, userID)
		}

		if len(actualMemberIDs) != len(s.conversationMemberIDs) {
			if attempt < 10 {
				return true, nil
			}
			return false, fmt.Errorf("unexpected number of conversation members, want %d got %d", len(s.conversationMemberIDs), len(actualMemberIDs))
		}

		return false, nil
	})
	if err != nil {
		return ctx, err
	}
	return common.StepStateToContext(ctx, commonState), nil
}

func (s *ConversationMembersSuite) isRemovedFromConversation(ctx context.Context, member string) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)

	ctxWithRp := common.ContextWithResourcePath(ctx, commonState.CurrentResourcePath)
	currentStaffID := commonState.CurrentStaff.ID

	if len(commonState.Students) == 0 {
		return ctx, fmt.Errorf("no student created")
	}

	studentID := commonState.Students[0].ID

	var removeMembers []string

	switch member {
	case "teacher":
		removeMembers = []string{currentStaffID}
	case "teacher and student":
		removeMembers = []string{currentStaffID, studentID}
	default:
		return ctx, fmt.Errorf("no argument match")
	}
	req := &cpb.RemoveConversationMembersRequest{
		ConversationId: s.createdConversationID,
		MemberIds:      removeMembers,
	}
	_, err := cpb.NewConversationModifierServiceClient(s.ConversationMgmtGRPCConn).RemoveConversationMembers(
		ctxWithRp,
		req,
	)
	if err != nil {
		return ctx, fmt.Errorf("failed RemoveConversationMembers: %+v", err)
	}

	// check if conversation member data is correct
	query := `
		SELECT cm.user_id
		FROM conversation c
		JOIN conversation_member cm ON cm.conversation_id = c.conversation_id
		WHERE c.conversation_id = $1
		AND cm.status = 'CONVERSATION_MEMBER_STATUS_ACTIVE'
		AND cm.deleted_at IS NULL
		AND c.deleted_at IS NULL;
	`
	err = try.Do(func(attempt int) (bool, error) {
		actualMemberIDs := []string{}
		rows, err := s.TomDBConn.Query(ctx, query, database.Text(s.createdConversationID))
		if err != nil {
			return false, fmt.Errorf("failed Query: %+v", err)
		}
		defer rows.Close()
		for rows.Next() {
			var userID string
			if err = rows.Scan(&userID); err != nil {
				return false, fmt.Errorf("failed Scan: %+v", err)
			}
			actualMemberIDs = append(actualMemberIDs, userID)
		}

		if len(actualMemberIDs) == 0 {
			if attempt < 10 {
				return true, nil
			}
			return false, fmt.Errorf("expected conversation %s to have some members", s.createdConversationID)
		}

		// removed members should not exist in actual member IDs
		foundMembers := slices.Filter(nil, actualMemberIDs, func(s string) bool {
			// if actual member is in the list removed member, return it as found
			return slices.Contains(removeMembers, s)
		})

		if len(foundMembers) > 0 {
			return false, fmt.Errorf("member [%+v] must not belong to conversation %s", foundMembers, s.createdConversationID)
		}

		return false, nil
	})
	if err != nil {
		return ctx, err
	}
	return common.StepStateToContext(ctx, commonState), nil
}
