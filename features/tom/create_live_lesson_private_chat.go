package tom

import (
	"context"
	"time"

	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/pkg/errors"
)

func (s *suite) createLiveLessonPrivateConversation(ctx context.Context, role1 string, role2 string) (context.Context, error) {
	cliContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var user2Id string

	var token string
	switch role1 {
	case "student":
		studentToken, _ := s.genStudentToken(s.LessonChatState.studentsInLesson[0])
		s.senderToken = studentToken
		token = studentToken
	case "teacher":
		teacherToken, _ := s.genTeacherToken(s.LessonChatState.TeachersInLesson[0])
		token = teacherToken
		s.senderToken = token
	default:
	}

	switch role2 {
	case "student":
		user2Id = s.LessonChatState.studentsInLesson[0]
		studentToken, _ := s.genStudentToken(user2Id)
		s.receiverToken = studentToken
	case "teacher":
		user2Id = s.LessonChatState.TeachersInLesson[1]
		teacherToken, _ := s.genTeacherToken(user2Id)
		s.receiverToken = teacherToken
	default:
	}

	lessonID := s.lessonID

	lcm := tpb.NewLessonChatModifierServiceClient(s.Conn)
	rsp, err := lcm.CreateLiveLessonPrivateConversation(contextWithToken(cliContext, token), &tpb.CreateLiveLessonPrivateConversationRequest{
		UserIds:  []string{user2Id},
		LessonId: lessonID,
	})

	s.conversationID = rsp.Conversation.ConversationId

	return ctx, err
}

func (s *suite) seeLiveLessonConversation(ctx context.Context, role1 string) (context.Context, error) {
	cliContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token := s.senderToken
	conversationID := s.conversationID

	rsp, err := tpb.NewChatReaderServiceClient(s.Conn).GetConversationV2(contextWithToken(cliContext, token), &tpb.GetConversationV2Request{
		ConversationId: conversationID,
	})
	if err != nil {
		return ctx, err
	}

	if rsp.Conversation.ConversationId != s.conversationID {
		return ctx, errors.New("Private conversation isn't created")
	}
	return ctx, nil
}
