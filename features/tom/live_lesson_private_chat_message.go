package tom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

func (s *suite) sendMsgToPrivateConversation(ctx context.Context, conversationID string, msgContent string, senderToken string) (context.Context, error) {
	req := &legacytpb.SendMessageRequest{
		ConversationId: conversationID,
		Message:        msgContent,
		Type:           legacytpb.MESSAGE_TYPE_TEXT,
		LocalMessageId: idutil.ULIDNow(),
	}

	_, err := legacytpb.NewChatServiceClient(s.Conn).SendMessage(contextWithToken(ctx, senderToken), req)
	return ctx, err
}

func (s *suite) sendsMessageToThePrivateConversationWithContent(ctx context.Context, user string, numOfMessage int, content string) (context.Context, error) {
	token := s.senderToken
	conversationID := s.conversationID

	for i := 0; i < numOfMessage; i++ {
		ctx, err := s.sendMsgToPrivateConversation(ctx, conversationID, content, token)
		if err != nil {
			return ctx, err
		}
		if i == numOfMessage-1 {
			s.lastMessage = content
		}
	}
	return ctx, nil
}

func (s *suite) userVerifyMessageWithContentInPrivateConversation(ctx context.Context, conversationID string, numMsg int, msgContent string, token string) (context.Context, error) {
	req := &tpb.LiveLessonPrivateConversationMessagesRequest{
		ConversationId: conversationID,
		Paging: &cpb.Paging{
			Limit: 100,
		},
	}

	rsp, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonPrivateConversationMessages(contextWithToken(ctx, token), req)
	if err != nil {
		return ctx, err
	}

	messages := rsp.GetMessages()

	if len(messages) != numMsg {
		return ctx, fmt.Errorf("LiveLessonPrivateConversationMessages: Missing messages")
	}
	for _, m := range messages {
		content := m.Content
		if content != msgContent {
			return ctx, fmt.Errorf("LiveLessonPrivateConversationMessages: Wrong content")
		}
	}

	return ctx, nil
}

func (s *suite) userSeeMessageWithContentInPrivateConversation(ctx context.Context, user string, numMsg int, msgContent string) (context.Context, error) {
	conversationID := s.conversationID
	token := s.receiverToken

	return s.userVerifyMessageWithContentInPrivateConversation(ctx, conversationID, numMsg, msgContent, token)
}

func (s *suite) userRefreshLiveLessonSessionForPrivateConversation(ctx context.Context, user string) (context.Context, error) {
	token := s.receiverToken

	_, err := tpb.NewLessonChatReaderServiceClient(s.Conn).RefreshLiveLessonSession(contextWithToken(ctx, token), &tpb.RefreshLiveLessonSessionRequest{LessonId: s.LessonChatState.lessonID})
	return ctx, err
}

func (s *suite) multipleTeacherCreatePrivateConversationsWithAStudent(ctx context.Context) (context.Context, error) {
	cliContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lessonID := s.lessonID
	studentID := s.LessonChatState.studentsInLesson[0]
	studentToken, err := s.genStudentToken(studentID)
	if err != nil {
		return ctx, err
	}
	s.receiverToken = studentToken

	teacherIDs := s.LessonChatState.TeachersInLesson
	senderTokens := make([]string, 0)
	privateConversationIDs := make([]string, 0)
	for _, teacherID := range teacherIDs {
		token, err := s.genStudentToken(teacherID)
		if err != nil {
			return ctx, err
		}
		senderTokens = append(senderTokens, token)

		lcm := tpb.NewLessonChatModifierServiceClient(s.Conn)
		rsp, err := lcm.CreateLiveLessonPrivateConversation(contextWithToken(cliContext, token), &tpb.CreateLiveLessonPrivateConversationRequest{
			UserIds:  []string{studentID},
			LessonId: lessonID,
		})
		if err != nil {
			return ctx, err
		}
		privateConversationIDs = append(privateConversationIDs, rsp.Conversation.ConversationId)
	}

	s.senderTokens = senderTokens
	s.privateConversationIDs = privateConversationIDs

	return ctx, nil
}

func (s *suite) multipleTeacherSendMessageToPrivateConversations(ctx context.Context, numOfMessage int, content string) (context.Context, error) {
	cliContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for index, senderToken := range s.senderTokens {
		conversationID := s.privateConversationIDs[index]

		for i := 0; i < numOfMessage; i++ {
			ctx, err := s.sendMsgToPrivateConversation(cliContext, conversationID, content, senderToken)
			if err != nil {
				return ctx, err
			}
		}
	}

	return ctx, nil
}

func (s *suite) userSeeMessageWithContentInAllPrivateConversation(ctx context.Context, user string, numMsg int, msgContent string) (context.Context, error) {
	token := s.receiverToken

	for _, conversationID := range s.privateConversationIDs {
		ctx, err := s.userVerifyMessageWithContentInPrivateConversation(ctx, conversationID, numMsg, msgContent, token)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func (s *suite) verifyAllConversationsHaveTheSameLatestStartTime(ctx context.Context) (context.Context, error) {
	lessonID := s.LessonChatState.lessonID
	lessonConversationID := s.LessonConversationMap[lessonID]
	privateConversationIDs := s.privateConversationIDs

	var latestStartTime pgtype.Timestamptz
	var privateLatestStartTime pgtype.Timestamptz

	query := `SELECT cl.latest_start_time FROM conversation_lesson cl WHERE cl.conversation_id = $1`
	if err := doRetry(func() (bool, error) {
		if err := s.DB.QueryRow(ctx, query, lessonConversationID).Scan(&latestStartTime); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return true, fmt.Errorf("lesson conversation is not created")
			}
			return false, err
		}

		return false, nil
	}); err != nil {
		return ctx, err
	}

	query = `SELECT cl.latest_start_time FROM private_conversation_lesson cl WHERE cl.conversation_id = $1`
	for _, conID := range privateConversationIDs {
		if err := doRetry(func() (bool, error) {
			if err := s.DB.QueryRow(ctx, query, conID).Scan(&privateLatestStartTime); err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return true, fmt.Errorf("lesson conversation is not created")
				}
				return false, err
			}

			return false, nil
		}); err != nil {
			return ctx, err
		}

		if privateLatestStartTime != latestStartTime {
			return ctx, errors.Errorf("latest start time is different")
		}
	}
	return ctx, nil
}
