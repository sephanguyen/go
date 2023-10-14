package tom

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/internal/golibs/try"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

func (s *suite) studentListConversation(ctx context.Context, studenttype string) (context.Context, error) {
	req := &pb.ConversationListRequest{
		Limit: 10,
	}
	studentID := s.studentID
	if studenttype == "in lesson" {
		studentID = s.studentsInLesson[0]
	}
	tok, err := s.genStudentToken(studentID)
	if err != nil {
		return ctx, err
	}
	err = try.Do(func(attempt int) (bool, error) {
		time.Sleep(1 * time.Second)

		s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).ConversationList(contextWithToken(ctx, tok), req)
		if s.ResponseErr != nil {
			return attempt < 5, err
		}
		return false, nil
	})

	if err != nil {
		return ctx, err
	}

	return ctx, nil
}
func (s *suite) teacherListConversation(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &pb.ConversationListRequest{
		Limit: 10,
	}

	s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).ConversationList(contextWithToken(ctx2, s.TeacherToken), req)

	return ctx, nil
}
func (s *suite) tomMustNotReturnLessonConversation(ctx context.Context) (context.Context, error) {
	rsp := s.Response.(*pb.ConversationListResponse)
	for _, item := range rsp.Conversations {
		if item.GetConversationType() == pb.ConversationType(tpb.ConversationType_CONVERSATION_LESSON) {
			return ctx, fmt.Errorf("tom returned lesson conversation %s", item.ConversationId)
		}
	}
	return ctx, nil
}
