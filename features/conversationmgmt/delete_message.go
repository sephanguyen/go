package conversationmgmt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/conversationmgmt/common"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/agora_usermgmt/utils"
	"github.com/manabie-com/backend/internal/conversationmgmt/modules/conversation/core/domain"
	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	cpb "github.com/manabie-com/backend/pkg/manabuf/conversationmgmt/v1"

	"github.com/cucumber/godog"
)

type DeleteMessageSuite struct {
	*common.ConversationMgmtSuite
	latestMessage *domain.Message
}

func (c *SuiteConstructor) InitDeleteMessage(dep *Dependency, godogCtx *godog.ScenarioContext) {
	s := &DeleteMessageSuite{
		ConversationMgmtSuite: dep.convCommonSuite,
	}

	stepsMapping := map[string]interface{}{
		`^a new "([^"]*)" and granted organization location logged in Back Office of a new organization with some exist locations$`: s.StaffGrantedRoleAndOrgLocationLoggedInBackOfficeOfNewOrg,
		`^school admin creates a students with first name is "([^"]*)" and last name is "([^"]*)"$`:                                 s.CreatesAStudentWithFirstNameAndLastName,
		`^a new staff with role teacher is created$`:                                                                                s.StaffWithRoleTeacher,
		`^waiting for Agora User has been created$`:                                                                                 s.WaitingForAgoraUserHasBeenCreated,
		`^current staff create "([^"]*)" conversations for students$`:                                                               s.CurrentStaffCreateCreateNumberOfConversationsForStudents,
		`^current staff add latest message for student\'s conversation$`:                                                            s.addLatestMessageForConversations,
		`^current staff delete this latest message$`:                                                                                s.currentStaffDeleteLatestMessage,
		`^latest message of student\'s conversation should be deleted$`:                                                             s.latestMessageIsDeleted,
	}

	c.InitScenarioStepMapping(godogCtx, stepsMapping)
}

func (s *DeleteMessageSuite) addLatestMessageForConversations(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	curentConvoID := commonState.Conversations[0].ID
	s.latestMessage = &domain.Message{
		ConversationID:  curentConvoID,
		VendorMessageID: idutil.ULIDNow(),
		Message:         "test message",
		VendorUserID:    utils.GetAgoraUserID(commonState.CurrentStaff.ID),
		UserID:          commonState.CurrentStaff.ID,
		Type:            "text",
		SentTime:        time.Now(),
		IsDeleted:       false,
		Media:           []domain.MessageMedia{},
	}
	stmt := `
		UPDATE conversation 
		SET latest_message = $2, latest_message_sent_time = $3
		WHERE conversation_id = $1
	`

	latestMessageBytes, _ := s.latestMessage.ToBytes()
	if _, err := s.TomDBConn.Exec(ctx, stmt, curentConvoID, database.JSONB(latestMessageBytes), database.Timestamptz(s.latestMessage.SentTime)); err != nil {
		return ctx, err
	}

	return ctx, nil
}

func (s *DeleteMessageSuite) currentStaffDeleteLatestMessage(ctx context.Context) (context.Context, error) {
	commonState := common.StepStateFromContext(ctx)
	ctx = common.ContextWithToken(ctx, commonState.CurrentStaff.Token)
	_, err := cpb.NewConversationModifierServiceClient(s.ConversationMgmtGRPCConn).DeleteMessage(ctx, &cpb.DeleteMessageRequest{
		ConversationId:  s.latestMessage.ConversationID,
		VendorMessageId: s.latestMessage.VendorMessageID,
	})
	if err != nil {
		return ctx, fmt.Errorf("failed to delete message: [%+v]", err)
	}

	return ctx, nil
}

func (s *DeleteMessageSuite) latestMessageIsDeleted(ctx context.Context) (context.Context, error) {
	resp, err := cpb.NewConversationReaderServiceClient(s.ConversationMgmtGRPCConn).
		GetConversationsDetail(ctx, &cpb.GetConversationsDetailRequest{ConversationIds: []string{s.latestMessage.ConversationID}})
	if err != nil {
		return ctx, fmt.Errorf("s.GetConversationsDetail: %v", err)
	}
	if len(resp.Conversations) == 0 {
		return ctx, fmt.Errorf("not found conversation")
	}

	updatedLatestMessage := &domain.Message{}
	err = json.Unmarshal(resp.Conversations[0].LatestMessage, updatedLatestMessage)
	if err != nil {
		return ctx, fmt.Errorf("cannot get latest message: [%+v]", err)
	}

	if !updatedLatestMessage.IsDeleted {
		return ctx, fmt.Errorf("latest message is deleted, but haven't been updated yet")
	}

	return ctx, nil
}
