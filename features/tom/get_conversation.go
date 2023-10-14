package tom

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/manabie-com/backend/internal/golibs/idutil"
	pb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
)

func (s *suite) aGetConversationRequest(ctx context.Context) (context.Context, error) {
	s.Request = &pb.GetConversationRequest{}
	return ctx, nil
}
func (s *suite) AGetConversationRequest(ctx context.Context) (context.Context, error) {
	return s.aGetConversationRequest(ctx)
}
func (s *suite) aUserMakesGetConversationRequestWithAnInvalidToken(ctx context.Context) (context.Context, error) {
	ctx2, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).GetConversation(contextWithToken(ctx2, "invalid-token"),
		s.Request.(*pb.GetConversationRequest))
	return ctx, nil
}
func (s *suite) aInGetConversationRequest(ctx context.Context, arg1 string) (context.Context, error) {
	switch arg1 {
	case "current studentId":
		s.Request.(*pb.GetConversationRequest).UserId = s.studentID
	case "current classId":
		s.Request.(*pb.GetConversationRequest).ClassId = uint32(s.classID)
	case "another studentId":
		s.Request.(*pb.GetConversationRequest).UserId = idutil.ULIDNow()
	case "current conversationID":
		s.Request.(*pb.GetConversationRequest).ConversationId = s.conversationID
	}

	return ctx, nil
}
func (s *suite) AInGetConversationRequest(ctx context.Context, arg1 string) (context.Context, error) {
	return s.aInGetConversationRequest(ctx, arg1)
}
func (s *suite) getConversationResponseHasUserWithRoleStatus(ctx context.Context, numUser int, role string, status string) (context.Context, error) {
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
	for _, u := range s.Response.(*pb.GetConversationResponse).Conversation.Users {
		if u.Group != expectRole.String() {
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
func (s *suite) getConversationResponseHasLatestMessageWithContent(ctx context.Context, msgContent string) (context.Context, error) {
	res := s.Response.(*pb.GetConversationResponse)
	if res.Conversation.LastMessage == nil {
		return ctx, fmt.Errorf("last message is nil")
	}
	if res.Conversation.LastMessage.Content != msgContent {
		return ctx, fmt.Errorf("want last message to has content %s, actual content is %s", msgContent, res.Conversation.LastMessage.Content)
	}
	return ctx, nil
}

func (s *suite) aTeacherMakesGetConversationRequestWith(ctx context.Context, convIDtype string) (context.Context, error) {
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
	req := &pb.GetConversationRequest{
		ConversationId: convid,
	}
	s.Request = req
	s.Response, s.ResponseErr = pb.NewChatServiceClient(s.Conn).GetConversation(contextWithToken(ctx, teacherToken), req)
	return ctx, nil
}

func (s *suite) tomMustReturnConversationWithTypeInGetConversationResponse(ctx context.Context, convtype string) (context.Context, error) {
	resp := s.Response.(*pb.GetConversationResponse)
	reqConvID := s.Request.(*pb.GetConversationRequest).ConversationId
	if resp.Conversation.ConversationType.String() != convtype {
		return ctx, fmt.Errorf("want conversation type %s, has %s", convtype, resp.Conversation.ConversationType.String())
	}
	if reqConvID != resp.Conversation.ConversationId {
		return ctx, fmt.Errorf("want conversation id %s,has %s", reqConvID, resp.Conversation.ConversationId)
	}
	return ctx, nil
}
