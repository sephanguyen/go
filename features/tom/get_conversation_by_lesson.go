package tom

import (
	"context"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/tom"

	"github.com/pkg/errors"
)

func (s *suite) aConversationByLessonRequest(ctx context.Context) (context.Context, error) {
	s.Request = &pb.ConversationByLessonRequest{
		Limit:    10,
		EndAt:    nil,
		LessonId: "",
	}

	return ctx, nil
}
func (s *suite) aInConversationByLessonRequest(ctx context.Context, arg1 string) (context.Context, error) {
	if arg1 == "current lessonID" {
		s.Request.(*pb.ConversationByLessonRequest).LessonId = s.lessonID
	}

	return ctx, nil
}
func (s *suite) aUserGetAllConversationOfLesson(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).ConversationByLesson(contextWithToken(ctx2, s.studentToken), s.Request.(*pb.ConversationByLessonRequest))
	return ctx, nil
}
func (s *suite) tomMustReturnConversationOfLesson(ctx context.Context, expectedConversation int) (context.Context, error) {
	resp := s.Response.(*pb.ConversationByLessonResponse)

	for _, conversation := range resp.Conversations {
		if conversation.ConversationType != pb.CONVERSATION_LESSON {
			return ctx, errors.Errorf("Expected conversation type: %s", pb.CONVERSATION_LESSON.String())
		}
	}

	if len(resp.Conversations) != expectedConversation {
		return ctx, errors.Errorf("total conversation does not match, expected: %d, got: %d", expectedConversation, len(resp.Conversations))
	}

	return ctx, nil
}
