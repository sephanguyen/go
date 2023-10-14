package tom

import (
	"context"
	"time"

	"github.com/manabie-com/backend/internal/golibs/godogutil"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
)

func (s *suite) studentSeenConversation(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	s.RequestAt = time.Now()
	req := &pb.SeenMessageRequest{
		ConversationId: s.conversationID,
	}
	token, err := s.genStudentToken(s.studentID)
	if err != nil {
		return ctx, err
	}
	s.studentToken = token

	s.Request = req
	s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).SeenMessage(contextWithToken(ctx2, token), req)
	return ctx, nil
}
func (s *suite) teacherSendMessageToConversation(ctx context.Context) (context.Context, error) {
	ctx, err := s.aSendMessageRequest(ctx)
	if err != nil {
		return ctx, err
	}
	return s.aSendAChatMessageToConversation(ctx, "teacher")
}
func (s *suite) tomMustMarkMessagesInConversationAsReadForStudent(ctx context.Context) (context.Context, error) {
	return godogutil.MultiErrChain(ctx,
		s.getTotalUnreadMessage, "student",
		s.tomMustReturnsTotalUnreadMessage, 0,
	)
}
