package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/manabie-com/backend/features/communication/common"
	"github.com/manabie-com/backend/features/conversationmgmt/common/entities"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/pkg/errors"
)

func (s *ConversationMgmtSuite) CurrentStaffCreateCreateNumberOfConversationsForStudents(ctx context.Context, nums string) (context.Context, error) {
	stepState := StepStateFromContext(ctx)
	if stepState.Organization == nil || len(stepState.Organization.Staffs) == 0 {
		return ctx, errors.New("missing created organization and staff with granted role step")
	}

	numConversation, err := strconv.Atoi(nums)
	if err != nil {
		return ctx, fmt.Errorf("s.StudentCreateNumberOfConversations: %v", err)
	}

	memberIDs := make([]string, 0)
	for _, student := range stepState.Students {
		memberIDs = append(memberIDs, student.ID)
	}

	memberIDs = append(memberIDs, stepState.CurrentStaff.ID)

	ctx = common.ContextWithToken(ctx, stepState.CurrentStaff.Token)
	conversationModifierService := cpb.NewConversationModifierServiceClient(s.ConversationMgmtGRPCConn)

	for i := 0; i < numConversation; i++ {
		conversation := &entities.Conversation{
			Name:      idutil.ULIDNow(),
			MemberIDs: memberIDs,
			OptionalConfig: []byte(`{
				"test_field": "test_value"
			}`),
		}

		req := &cpb.CreateConversationRequest{
			Name:           conversation.Name,
			MemberIds:      conversation.MemberIDs,
			OptionalConfig: conversation.OptionalConfig,
		}
		resp, err := conversationModifierService.CreateConversation(ctx, req)
		if err != nil {
			return ctx, err
		}

		conversation.ID = resp.ConversationId
		stepState.Conversations = append(stepState.Conversations, conversation)
	}
	return StepStateToContext(ctx, stepState), nil
}
