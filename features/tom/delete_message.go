package tom

import (
	"context"
	"time"

	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/pkg/errors"
)

func (s *suite) aDeleteMessage(ctx context.Context, user string, msgOwner string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var token string
	switch user {
	case "student":
		token = s.studentToken
	case "teacher":
		token = s.TeacherToken
	default:
	}

	var messageID string
	switch msgOwner {
	case "own":
		messageID = s.messageID
	case "student":
		messageID = s.studentMessageID
	default:
	}

	s.RequestAt = time.Now()
	req := &tpb.DeleteMessageRequest{
		MessageId: messageID,
	}

	s.Request = req
	s.Response, s.ResponseErr = tpb.NewChatModifierServiceClient(s.Conn).DeleteMessage(contextWithToken(ctx2, token), req)

	return ctx, nil
}

func (s *suite) aSeeDeletedMessageInConversation(ctx context.Context, user string) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	tokens := make([]string, 0)
	switch user {
	case "student":
		tokens = append(tokens, s.studentToken)
	case "teacher":
		if s.TeacherToken != "" {
			tokens = append(tokens, s.TeacherToken)
		} else {
			for _, token := range s.teacherTokens {
				tokens = append(tokens, token)
			}
		}

	default:
	}

	convID := s.conversationID
	for _, token := range tokens {
		s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).ConversationDetail(contextWithToken(ctx2, token), &pb.ConversationDetailRequest{
			ConversationId: convID,
			Limit:          10,
		})

		deletedMessages := make([]*pb.MessageResponse, 0)

		messagesResp := s.Response.(*pb.ConversationDetailResponse).Messages
		for _, message := range messagesResp {
			if message.IsDeleted {
				deletedMessages = append(deletedMessages, message)
			}
		}

		if len(deletedMessages) == 0 {
			return ctx, errors.New("message is not deleted")
		}

		for _, message := range deletedMessages {
			if message.Content != "" || message.UrlMedia != "" {
				return ctx, errors.New("Content or media url is not hidden")
			}
		}
	}

	return ctx, nil
}

func (s *suite) aStudentSendsItemWithContent(ctx context.Context, msgType, msgContent string) (context.Context, error) {
	userID := s.studentID
	return s.userSendsItemWithContent(ctx, msgType, msgContent, userID, cpb.UserGroup_USER_GROUP_STUDENT)
}

func (s *suite) teacherSeeDeletedChatInLessonChat(ctx context.Context) (context.Context, error) {
	context, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	token := s.TeacherToken
	conversationID := s.conversationID

	rsp, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonConversationMessages(contextWithToken(context, token), &tpb.LiveLessonConversationMessagesRequest{
		ConversationId: conversationID,
	})
	if err != nil {
		return ctx, err
	}

	if len(rsp.Messages) == 0 {
		return ctx, errors.New("Teacher can't see deleted message in lesson chat")
	}

	if !rsp.Messages[0].IsDeleted {
		return ctx, errors.New("Message is not deleted")
	}

	if rsp.Messages[0].Content != "" {
		return ctx, errors.New("Content is not hidden")
	}
	return ctx, nil
}
