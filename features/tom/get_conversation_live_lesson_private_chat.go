package tom

import (
	"context"
	"fmt"
	"time"

	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

func (s *suite) getTheLiveLessonPrivateConversationDetail(ctx context.Context, user string) (context.Context, error) {
	cliContext, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	token := s.receiverToken

	cr := tpb.NewChatReaderServiceClient(s.Conn)
	rsp, err := cr.GetConversationV2(contextWithToken(cliContext, token), &tpb.GetConversationV2Request{
		ConversationId: s.conversationID,
	})
	s.Response = rsp
	return ctx, err
}

func (s *suite) seeLiveLessonConversationDetail(ctx context.Context, user string) (context.Context, error) {
	resp := s.Response.(*tpb.GetConversationV2Response)
	conversationType := tpb.ConversationType_CONVERSATION_LESSON_PRIVATE.String()
	conversation := resp.Conversation
	if conversationType != conversation.ConversationType.String() {
		return ctx, fmt.Errorf("want conversation type %s, has %s", conversationType, resp.Conversation.ConversationType.String())
	}
	if s.conversationID != conversation.ConversationId {
		return ctx, fmt.Errorf("want conversation id %s,has %s", s.conversationID, resp.Conversation.ConversationId)
	}
	if s.lastMessage != conversation.LastMessage.Content {
		return ctx, fmt.Errorf("want last message %s,has %s", s.lastMessage, resp.Conversation.LastMessage.Content)
	}
	if len(conversation.Users) != 2 {
		return ctx, fmt.Errorf("want num of users %d,has %d", 2, len(conversation.Users))
	}
	return ctx, nil
}
