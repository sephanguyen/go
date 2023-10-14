package tom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

func (s *suite) aGetConversationV2Request(ctx context.Context) (context.Context, error) {
	s.Request = &tpb.GetConversationV2Request{}
	return ctx, nil
}
func (s *suite) AGetConversationV2Request(ctx context.Context) (context.Context, error) {
	return s.aGetConversationV2Request(ctx)
}
func (s *suite) aUserMakesGetConversationV2RequestWithAnInvalidToken(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.Response, s.ResponseErr = tpb.NewChatReaderServiceClient(s.Conn).GetConversationV2(contextWithToken(ctx2, "invalid-token"),
		s.Request.(*tpb.GetConversationV2Request))
	return ctx, nil
}

func (s *suite) getConversationV2ResponseHasUserWithRoleStatus(ctx context.Context, numUser int, role string, status string) (context.Context, error) {
	var (
		expectRole      cpb.UserGroup
		expectIsPresent bool
	)
	switch status {
	case "active":
		expectIsPresent = true
	case "inactive":
		expectIsPresent = false
	}
	switch role {
	case "student":
		expectRole = cpb.UserGroup_USER_GROUP_STUDENT
	case "teacher":
		expectRole = cpb.UserGroup_USER_GROUP_TEACHER
	}
	actualusr := 0
	for _, u := range s.Response.(*tpb.GetConversationV2Response).Conversation.Users {
		if u.Group.String() != expectRole.String() {
			continue
		}
		if u.IsPresent != expectIsPresent {
			continue
		}
		actualusr++
	}
	if actualusr != numUser {
		return ctx, fmt.Errorf("want %d user to have role %s status %s, but actually have %d", numUser, role, status, actualusr)
	}
	return ctx, nil
}
func (s *suite) getConversationV2ResponseHasLatestMessageWithContent(ctx context.Context, msgContent string) (context.Context, error) {
	res := s.Response.(*tpb.GetConversationV2Response)
	if res.Conversation.LastMessage == nil {
		return ctx, fmt.Errorf("last message is nil")
	}
	if res.Conversation.LastMessage.Content != msgContent {
		return ctx, fmt.Errorf("want last message to has content %s, actual content is %s", msgContent, res.Conversation.LastMessage.Content)
	}
	return ctx, nil
}

func (s *suite) aTeacherMakesGetConversationV2RequestWith(ctx context.Context, convIDtype string) (context.Context, error) {
	var convid, aTeacherID string

	switch convIDtype {
	case "student conversation id":
		convid = s.conversationID
		aTeacherID = s.teachersInConversation[0]
	case "lesson conversation id":
		convid = s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
		aTeacherID = s.LessonChatState.TeachersInLesson[0]
	default:
		panic(fmt.Errorf("invalid test input: %s", convIDtype))
	}
	schoolID, _ := strconv.Atoi(s.schoolID)

	teacherToken, err := s.generateExchangeToken(aTeacherID, cpb.UserGroup_USER_GROUP_TEACHER.String(), applicantID, int64(schoolID), s.ShamirConn)
	if err != nil {
		return ctx, err
	}
	req := &tpb.GetConversationV2Request{
		ConversationId: convid,
	}
	s.Request = req
	s.Response, s.ResponseErr = tpb.NewChatReaderServiceClient(s.Conn).GetConversationV2(contextWithToken(ctx, teacherToken), req)
	return ctx, nil
}

func (s *suite) tomMustReturnConversationWithTypeInGetConversationV2Response(ctx context.Context, convtype string) (context.Context, error) {
	resp := s.Response.(*tpb.GetConversationV2Response)
	reqConvID := s.Request.(*tpb.GetConversationV2Request).ConversationId
	if resp.Conversation.ConversationType.String() != convtype {
		return ctx, fmt.Errorf("want conversation type %s, has %s", convtype, resp.Conversation.ConversationType.String())
	}
	if reqConvID != resp.Conversation.ConversationId {
		return ctx, fmt.Errorf("want conversation id %s,has %s", reqConvID, resp.Conversation.ConversationId)
	}
	return ctx, nil
}
