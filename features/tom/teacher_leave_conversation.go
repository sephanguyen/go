package tom

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/manabie-com/backend/internal/golibs/database"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	"github.com/manabie-com/backend/internal/golibs/try"
	domain "github.com/manabie-com/backend/internal/tom/domain/core"
	"github.com/manabie-com/backend/internal/tom/repositories"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *suite) aTeacherWhoJoinedAllConversations(ctx context.Context) (context.Context, error) {
	return godogutil.MultiErrChain(ctx,
		s.aSignedAsATeacher,
		s.teacherJoinAllConversations,
		s.returnsStatusCode, "OK",
		s.teacherMustBeMemberOfAllConversationsWithSpecifySchools,
	)
}
func (s *suite) teacherLeaveSomeConversations(ctx context.Context) (context.Context, error) {
	joined := len(s.JoinedConversationIDs)
	// nolint:gosec
	left := rand.Intn(joined)
	if left == 0 {
		left = 1
	}
	s.LeftConversationIDs = s.JoinedConversationIDs[:left]

	_, err := tpb.NewChatModifierServiceClient(s.Conn).LeaveConversations(contextWithToken(ctx, s.TeacherToken),
		&tpb.LeaveConversationsRequest{
			ConversationIds: s.LeftConversationIDs,
		})
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
func (s *suite) teacherMustNotBeMemberOfConversationsRecentlyLeft(ctx context.Context) (context.Context, error) {
	cmrepo := &repositories.ConversationMemberRepo{}
	memberships, err := cmrepo.FindByConversationIDs(ctx, s.DB, database.TextArray(s.LeftConversationIDs))
	if err != nil {
		return ctx, err
	}
	if len(memberships) != len(s.LeftConversationIDs) {
		return ctx, fmt.Errorf("some conversation members were removed instead of deactivate when leave conversation")
	}
	for _, members := range memberships {
		var found bool
		for _, member := range members {
			if member.UserID.String == s.teacherID {
				found = true
				if member.Status.String != domain.ConversationStatusInActive {
					return ctx, fmt.Errorf("expect status after leaving conversation to be %s, have %s", domain.ConversationStatusInActive, domain.ConversationStatusActive)
				}
			}
		}
		if !found {
			return ctx, fmt.Errorf("membership of teacher has been hard removed instead of deactivate")
		}
	}
	messageRepo := &repositories.MessageRepo{}
	for _, convID := range s.LeftConversationIDs {
		messages, err := messageRepo.FindMessages(ctx, s.DB, &domain.FindMessagesArgs{
			ConversationID:       database.Text(convID),
			EndAt:                database.Timestamptz(time.Now()),
			IncludeSystemMsg:     true,
			IncludeMessageTypes:  pgtype.TextArray{Status: pgtype.Null},
			ExcludeMessagesTypes: pgtype.TextArray{Status: pgtype.Null},
			Limit:                1,
		})
		if err != nil {
			return ctx, err
		}
		if messages[0].TargetUser.String != s.teacherID {
			return ctx, fmt.Errorf("message type system does not have target user = teacher id")
		}
	}

	return ctx, nil
}
func (s *suite) teacherNLeavesChatIDs(ctx context.Context, teacherID string, chatIDs []string) (context.Context, error) {
	token := s.teacherTokens[teacherID]

	_, err := tpb.NewChatModifierServiceClient(s.Conn).LeaveConversations(contextWithToken(context.Background(), token),
		&tpb.LeaveConversationsRequest{
			ConversationIds: chatIDs,
		})
	if err != nil {
		return ctx, err
	}
	repo := &repositories.ConversationMemberRepo{}
	memberShip, err := repo.FindByCIDsAndUserID(ctx, s.DB, database.TextArray(chatIDs), database.Text(teacherID))
	if err != nil {
		return ctx, err
	}
	for _, member := range memberShip {
		if member.Status.String != domain.ConversationStatusInActive {
			return ctx, fmt.Errorf("status of teacher is still %s after leaving conversation", member.Status.String)
		}
	}
	return ctx, nil
}
func (s *suite) teacherNumberLeavesStudentChat(ctx context.Context, teacherNum int) (context.Context, error) {
	teacherID := s.teachersInConversation[teacherNum-1]
	s.teacherWhoLeftChat = teacherID
	return s.teacherNLeavesChatIDs(ctx, teacherID, []string{s.conversationID})
}
func (s *suite) teacherLeavesChatIDs(ctx context.Context, chatIDs []string) (context.Context, error) {
	teacherID := s.teachersInConversation[0]
	s.teacherWhoLeftChat = teacherID
	return s.teacherNLeavesChatIDs(ctx, teacherID, chatIDs)
}
func (s *suite) teacherLeavesStudentChat(ctx context.Context) (context.Context, error) {
	return s.teacherLeavesChatIDs(ctx, []string{s.conversationID})
}
func (s *suite) teacherRejoinsStudentChat(ctx context.Context) (context.Context, error) {
	_, err := tpb.NewChatModifierServiceClient(s.Conn).JoinConversations(contextWithToken(context.Background(), s.TeacherToken),
		&tpb.JoinConversationsRequest{
			ConversationIds: []string{s.conversationID},
		},
	)
	return ctx, err
}
func (s *suite) teacherWhoLeftChatDoesNotReceiveSentMessage(ctx context.Context) (context.Context, error) {
	stream, exist := s.SubV2Clients[s.teacherWhoLeftChat]
	if !exist {
		return ctx, fmt.Errorf("stream for teacher does not exist")
	}
	defer func() {
		err := stream.CloseSend()
		if err != nil {
			s.ZapLogger.Error(fmt.Sprintf("error closing stream %s", err))
		}
	}()
	var newMsg *pb.MessageResponse

	for try := 0; try < 5; try++ {
		resp, err := stream.Recv()
		if err == io.EOF {
			return ctx, fmt.Errorf("received eof before try runs out")
		}
		newMsg = resp.GetEvent().GetEventNewMessage()
		if newMsg != nil {
			return ctx, fmt.Errorf("teacher who left conversation still received sent message: %v", newMsg)
		}
	}

	return ctx, nil
}
func (s *suite) teacherWhoLeftChatSendsAMessage(ctx context.Context) (context.Context, error) {
	return s.userSendsItemWithContent(ctx, "text", "Hello world", s.teacherWhoLeftChat, cpb.UserGroup_USER_GROUP_TEACHER)
}
func (s *suite) teacherWhoLeftChatCannotSendMessage(ctx context.Context) (context.Context, error) {
	ctx, err := s.userSendsItemWithContent(ctx, "text", "Hello world", s.teacherWhoLeftChat, cpb.UserGroup_USER_GROUP_TEACHER)
	sts, isGrpcError := status.FromError(err)
	if !isGrpcError {
		return ctx, fmt.Errorf("expect error is grpc error, but have %v", err)
	}
	if sts.Code() != codes.NotFound {
		return ctx, fmt.Errorf("expect error status not found, got %v", sts)
	}
	if sts.Message() != "not found conversation" {
		return ctx, fmt.Errorf("expect error message %s, got %s", "not found conversation", sts.Message())
	}
	return ctx, nil
}
func (s *suite) otherTeachersReceiveLeaveConversationSystemMessage(ctx context.Context) (context.Context, error) {
	otherTeachers := []string{}
	for _, teacherID := range s.teachersInConversation {
		if teacherID != s.teacherWhoLeftChat {
			otherTeachers = append(otherTeachers, teacherID)
		}
	}
	if len(otherTeachers) == 0 {
		return ctx, errors.New("there is no other teachers in conversation, re-input your test case")
	}
	return s.usersReceiveLeaveConversationSystemMessage(ctx, otherTeachers)
}
func (s *suite) usersReceiveLeaveConversationSystemMessage(ctx context.Context, users []string) (context.Context, error) {
	return s.usersReceiveMatchedNewMessage(ctx, users, func(newMsg *pb.MessageResponse) error {
		if newMsg.GetContent() != tpb.CodesMessageType_CODES_MESSAGE_TYPE_LEAVE_CONVERSATION.String() {
			return fmt.Errorf("want new message to be leave conversation system message, has %v", newMsg)
		}
		if newMsg.GetTargetUser() != s.teacherWhoLeftChat {
			return fmt.Errorf("message leave conversation does not have target user = teacher id")
		}
		return nil
	})
}
func (s *suite) studentReceiveLeaveConversationSystemMessage(ctx context.Context) (context.Context, error) {
	return s.usersReceiveLeaveConversationSystemMessage(ctx, []string{s.studentID})
}
func (s *suite) teacherWhoLeftConversationReceivesLeaveConversationSystemMessage(ctx context.Context) (context.Context, error) {
	return s.usersReceiveLeaveConversationSystemMessage(ctx, []string{s.teacherWhoLeftChat})
}
func (s *suite) teacherLeavesStudentChatAndChatHeDoesNotJoin(ctx context.Context, invalidChatType string) (context.Context, error) {
	chatIDs := []string{s.conversationID}
	switch invalidChatType {
	case "existing chat":
		s.invalidLeavingChat = s.ConversationIDs[0] // those chats were generated randomly in the test case
	case "non existing chat":
		s.invalidLeavingChat = idutil.ULIDNow()
	default:
		panic("invalid argument given in test case")
	}
	chatIDs = append(chatIDs, s.invalidLeavingChat)
	return s.teacherLeavesChatIDs(ctx, chatIDs)
}
func (s *suite) theInvalidChatDoesNotRecordTeacherMembership(ctx context.Context) (context.Context, error) {
	cmemrepo := &repositories.ConversationMemberRepo{}
	cmember, err := cmemrepo.FindByCIDAndUserID(ctx, s.DB, database.Text(s.invalidLeavingChat), database.Text(s.teacherWhoLeftChat))
	if err == nil && cmember.ConversationID.String == s.invalidLeavingChat {
		return ctx, fmt.Errorf("tom must not store membership of teacher for invalid chat")
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ctx, fmt.Errorf("expect err no rows, got %s", err)
	}
	return ctx, nil
}

func (s *suite) theConversationThatTeacherLeftIsInConversationList(ctx context.Context, displayed string) (context.Context, error) {
	req := &tpb.ListConversationsInSchoolRequest{
		Paging:     &cpb.Paging{Limit: 100},
		JoinStatus: tpb.ConversationJoinStatus_CONVERSATION_JOIN_STATUS_JOINED,
	}
	err := try.Do(func(attempt int) (bool, error) {
		res, err := tpb.NewChatReaderServiceClient(s.Conn).ListConversationsInSchoolV2(contextWithToken(context.Background(), s.TeacherToken), req)
		if err != nil {
			return false, err
		}
		checkList := map[string]struct{}{}
		for _, conv := range res.GetItems() {
			checkList[conv.GetConversationId()] = struct{}{}
		}
		switch displayed {
		case "displayed":
			if _, ok := checkList[s.conversationID]; !ok {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("want left conversation to be displayed back to chat list, but has none")
			}
		case "not displayed":
			if _, ok := checkList[s.conversationID]; ok {
				time.Sleep(2 * time.Second)
				return attempt < 5, fmt.Errorf("want left conversation to disappear from chat list, but still displayed")
			}
		default:
			panic("invalid argument in test case")
		}
		return false, nil
	})
	if err != nil {
		fmt.Printf("[DEBUG] school id: %s,convid %s, teacher id %s\n", s.schoolID, s.conversationID, s.teacherWhoLeftChat)
	}
	return ctx, err
}
