package tom

import (
	"context"
	"fmt"

	"github.com/manabie-com/backend/internal/bob/constants"
	"github.com/manabie-com/backend/internal/golibs/godogutil"
	"github.com/manabie-com/backend/internal/golibs/idutil"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
)

func (s *suite) studentsInLessonSeenTheConversation(ctx context.Context) (context.Context, error) {
	convID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	toks, err := s.genStudentTokens(s.LessonChatState.studentsInLesson)
	if err != nil {
		return ctx, err
	}
	for idx := range s.LessonChatState.studentsInLesson {
		_, err := legacytpb.NewChatServiceClient(s.Conn).SeenMessage(contextWithToken(context.Background(), toks[idx]), &legacytpb.SeenMessageRequest{ConversationId: convID})
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) theSecondTeacherInLessonSeenTheConversation(ctx context.Context) (context.Context, error) {
	convID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	tok, err := s.genTeacherToken(s.LessonChatState.secondTeacher)
	if err != nil {
		return ctx, err
	}
	_, err = legacytpb.NewChatServiceClient(s.Conn).SeenMessage(contextWithToken(context.Background(), tok), &legacytpb.SeenMessageRequest{ConversationId: convID})
	return ctx, err
}
func (s *suite) userWithTokenSeesStatusCallingLiveLessonConversationDetail(ctx context.Context, seenStatus, token string) (context.Context, error) {
	var seen bool
	switch seenStatus {
	case "seen":
		seen = true
	case "unseen":
	default:
		panic(fmt.Sprintf("invalid seen status: %s", seenStatus))
	}

	req := tpb.LiveLessonConversationDetailRequest{
		LessonId: s.LessonChatState.lessonID,
	}
	res, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonConversationDetail(
		contextWithToken(context.Background(), token),
		&req,
	)
	if err != nil {
		return ctx, err
	}
	if seen != res.Conversation.Seen {
		return ctx, fmt.Errorf("want seen status to be %v, has %v", seen, res.Conversation.Seen)
	}
	return ctx, nil
}
func (s *suite) studentsSeesStatusCallingLiveLessonConversationDetail(ctx context.Context, seenStatus string) (context.Context, error) {
	toks, err := s.genStudentTokens(s.LessonChatState.studentsInLesson)
	if err != nil {
		return ctx, err
	}
	for idx := range s.LessonChatState.studentsInLesson {
		ctx, err := s.userWithTokenSeesStatusCallingLiveLessonConversationDetail(ctx, seenStatus, toks[idx])
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) theSecondTeacherSeesStatusCallingLiveLessonConversationDetail(ctx context.Context, seenStatus string) (context.Context, error) {
	tok, err := s.genTeacherToken(s.LessonChatState.secondTeacher)
	if err != nil {
		return ctx, err
	}
	return s.userWithTokenSeesStatusCallingLiveLessonConversationDetail(ctx, seenStatus, tok)
}
func (s *suite) theSecondTeacherSeesMessagesCallingLiveLessonConversationMessages(ctx context.Context, msgCount int) (context.Context, error) {
	convID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	req := tpb.LiveLessonConversationMessagesRequest{ConversationId: convID, Paging: &cpb.Paging{Limit: 100}}
	tok, err := s.genTeacherToken(s.LessonChatState.secondTeacher)
	if err != nil {
		return ctx, err
	}
	res, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonConversationMessages(contextWithToken(ctx, tok), &req)
	if err != nil {
		return ctx, err
	}
	if len(res.Messages) != msgCount {
		return ctx, fmt.Errorf("expect messages list has %d items, actually have %d items", msgCount, len(res.Messages))
	}
	return ctx, nil
}
func (s *suite) receiveMsgWithTypeContentFromStream(ctx context.Context, stream legacytpb.ChatService_SubscribeV2Client, msgType string, msgContent string) (context.Context, error) {
	expectType := legacytpb.MESSAGE_TYPE_TEXT
	switch msgType {
	case "text":
	case "system":
		expectType = legacytpb.MESSAGE_TYPE_SYSTEM
	default:
		panic(fmt.Errorf("unsupported msg type %s", msgType))
	}
	var found bool
	for try := 0; try < 10; try++ {
		resp, err := stream.Recv()
		if err != nil {
			return ctx, err
		}
		if msg := resp.Event.GetEventNewMessage(); msg != nil {
			if msg.Type != expectType {
				continue
			}
			if msg.Content != msgContent {
				return ctx, fmt.Errorf("want latest received message to be %s, has %s", msgContent, msg.Content)
			}
			found = true
			break
		}
	}
	if !found {
		return ctx, fmt.Errorf("not received any new message from stream")
	}
	return ctx, nil
}
func (s *suite) theInLessonReceivesMessageWithTypeWithContent(ctx context.Context, person string, numMsg int, msgtype, msgContent string) (context.Context, error) {
	ids := []string{}
	switch person {
	case "second teacher":
		ids = append(ids, s.LessonChatState.secondTeacher)
	case "students":
		ids = append(ids, s.LessonChatState.studentsInLesson...)
	default:
		panic(fmt.Sprintf("insupported person arg: %s", person))
	}
	for _, id := range ids {
		stream, ok := s.SubV2Clients[id]
		if !ok {
			return ctx, fmt.Errorf("stream for %s not found", person)
		}
		for i := 0; i < numMsg; i++ {
			ctx, err := s.receiveMsgWithTypeContentFromStream(ctx, stream, msgtype, msgContent)
			if err != nil {
				return ctx, err
			}
		}
	}
	return ctx, nil
}
func (s *suite) aStudentSendsMessageWithContentToLiveLessonChat(ctx context.Context, numMsg int, msgContent string) (context.Context, error) {
	astudent := s.LessonChatState.studentsInLesson[0]
	for i := 0; i < numMsg; i++ {
		ctx, err := s.userWithIDAndGroupSendMessageToLiveLesson(ctx, astudent, cpb.UserGroup_USER_GROUP_STUDENT, msgContent)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) aTeacherSendsMessageWithContentToLiveLessonChat(ctx context.Context, numMsg int, msgContent string) (context.Context, error) {
	ateacher := s.LessonChatState.TeachersInLesson[0]
	for i := 0; i < numMsg; i++ {
		ctx, err := s.userWithIDAndGroupSendMessageToLiveLesson(ctx, ateacher, cpb.UserGroup_USER_GROUP_TEACHER, msgContent)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) theFirstTeacherSendsMessageWithContentToLiveLessonChat(ctx context.Context, numMsg int, msgContent string) (context.Context, error) {
	for i := 0; i < numMsg; i++ {
		ctx, err := s.userWithIDAndGroupSendMessageToLiveLesson(ctx, s.LessonChatState.firstTeacher, cpb.UserGroup_USER_GROUP_TEACHER, msgContent)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
func (s *suite) studentsJoinLessonWithoutRefreshingLessonSession(ctx context.Context) (context.Context, error) {
	tok, err := s.genStudentTokens(s.LessonChatState.studentsInLesson)
	if err != nil {
		return ctx, err
	}
	return s.makeUsersSubscribeV2Ctx(ctx, s.LessonChatState.studentsInLesson, tok)
}
func (s *suite) aSecondTeacherJoinsLessonWithoutRefreshingLessonSession(ctx context.Context) (context.Context, error) {
	ctx, err := godogutil.MultiErrChain(ctx, s.aEvtLessonWithMessage, "JoinLesson", s.aValidIDInJoinLesson, cpb.UserGroup_USER_GROUP_TEACHER.String(), s.bobSendEventEvtLesson, s.tomAddAboveUserToThisLessonConversation, "must")
	if err != nil {
		return ctx, err
	}
	s.LessonChatState.secondTeacher = s.teacherID
	tok, err := s.genTeacherToken(s.secondTeacher)
	if err != nil {
		return ctx, err
	}
	return s.makeUsersSubscribeV2Ctx(ctx, []string{s.LessonChatState.secondTeacher}, []string{tok})
}
func (s *suite) aSecondTeacherJoinsLessonRefreshingLessonSession(ctx context.Context) (context.Context, error) {
	ctx, err := s.aSecondTeacherJoinsLessonWithoutRefreshingLessonSession(ctx)
	if err != nil {
		return ctx, err
	}
	tok, err := s.genTeacherToken(s.LessonChatState.secondTeacher)
	if err != nil {
		return ctx, err
	}
	_, err = tpb.NewLessonChatReaderServiceClient(s.Conn).RefreshLiveLessonSession(contextWithToken(ctx, tok), &tpb.RefreshLiveLessonSessionRequest{LessonId: s.LessonChatState.lessonID})
	return ctx, err
}
func (s *suite) aTeacherJoinsLessonCreatingNewLessonSession(ctx context.Context) (context.Context, error) {
	ctx, err := godogutil.MultiErrChain(ctx, s.aEvtLessonWithMessage, "JoinLesson", s.aValidIDInJoinLesson, cpb.UserGroup_USER_GROUP_TEACHER.String(), s.bobSendEventEvtLesson, s.tomAddAboveUserToThisLessonConversation, "must")
	if err != nil {
		return ctx, err
	}
	currentTeacher := s.teacherID
	s.LessonChatState.firstTeacher = s.LessonChatState.TeachersInLesson[0]
	tok, err := s.genTeacherToken(currentTeacher)
	s.TeacherToken = tok
	if err != nil {
		return ctx, err
	}
	ctx, err = s.makeUsersSubscribeV2Ctx(ctx, []string{currentTeacher}, []string{tok})
	if err != nil {
		return ctx, err
	}
	_, err = tpb.NewLessonChatReaderServiceClient(s.Conn).RefreshLiveLessonSession(contextWithToken(ctx, tok), &tpb.RefreshLiveLessonSessionRequest{LessonId: s.LessonChatState.lessonID})
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
func (s *suite) teacherSeesCorrectInfoCallingLiveLessonConversationDetail(ctx context.Context) (context.Context, error) {
	req := tpb.LiveLessonConversationDetailRequest{LessonId: s.LessonChatState.lessonID}
	tok, err := s.genTeacherToken(s.LessonChatState.firstTeacher)
	if err != nil {
		return ctx, err
	}

	res, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonConversationDetail(contextWithToken(context.Background(), tok), &req)
	if err != nil {
		return ctx, err
	}
	expectConversationID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	conv := res.GetConversation()
	if conv.ConversationId != expectConversationID {
		return ctx, fmt.Errorf("want conversation id %s, has %s", expectConversationID, conv.ConversationId)
	}
	if conv.ConversationName != s.LessonChatState.lessonName {
		return ctx, fmt.Errorf("want conversation name %s, has %s", s.LessonChatState.lessonName, conv.ConversationName)
	}
	var foundTeacher bool
	for idx := range conv.Users {
		if conv.Users[idx].GetId() == s.teacherID {
			foundTeacher = true
		}
	}
	if !foundTeacher {
		return ctx, fmt.Errorf("teacher is not a member of lesson conversation")
	}
	return ctx, nil
}
func (s *suite) teacherSeesCorrectLatestMessageCallingLiveLessonConversationDetail(ctx context.Context) (context.Context, error) {
	req := tpb.LiveLessonConversationDetailRequest{LessonId: s.LessonChatState.lessonID}
	tok, err := s.genTeacherToken(s.LessonChatState.firstTeacher)
	if err != nil {
		return ctx, err
	}

	res, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonConversationDetail(contextWithToken(context.Background(), tok), &req)
	if err != nil {
		return ctx, err
	}
	expectedLastMsg := s.sentMessages[len(s.sentMessages)-1]
	if expectedLastMsg.Message != res.Conversation.LastMessage.Content {
		return ctx, fmt.Errorf("want latest msg to have content %s, has %s", expectedLastMsg.Message, res.Conversation.LastMessage.Content)
	}
	return ctx, nil
}
func (s *suite) theSecondTeacherSeesMessagesWithContentCallingLiveLessonConversationMessages(ctx context.Context, numMsg int, msgContent string) (context.Context, error) {
	convID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	req := tpb.LiveLessonConversationMessagesRequest{ConversationId: convID, Paging: &cpb.Paging{Limit: 100}}
	tok, err := s.genTeacherToken(s.LessonChatState.firstTeacher)
	if err != nil {
		return ctx, err
	}
	res, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonConversationMessages(contextWithToken(ctx, tok), &req)
	if err != nil {
		return ctx, err
	}
	expectConversationID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	if len(res.Messages) != numMsg {
		return ctx, fmt.Errorf("messages returned in api are %d compared to actual %d wanted messages", len(res.Messages), numMsg)
	}
	for idx := range res.Messages {
		has := res.Messages[idx]
		if has.ConversationId != expectConversationID {
			return ctx, fmt.Errorf("wrong conversation id in message response, want %s, has %s", expectConversationID, has.ConversationId)
		}
		if has.Content != msgContent {
			return ctx, fmt.Errorf("msg %d-th has content %s, but want %s", idx, has.Content, msgContent)
		}
	}
	return ctx, nil
}
func (s *suite) userWithIDAndGroupSendMessageToLiveLesson(ctx context.Context, userID string, userGroup cpb.UserGroup, msgContent string) (context.Context, error) {
	conversationID := s.LessonChatState.LessonConversationMap[s.LessonChatState.lessonID]
	sendMsgReq := &legacytpb.SendMessageRequest{
		ConversationId: conversationID,
		LocalMessageId: idutil.ULIDNow(),
	}
	s.conversationID = conversationID

	sendMsgReq.Type = legacytpb.MESSAGE_TYPE_TEXT
	sendMsgReq.Message = msgContent
	s.sentMessages = append(s.sentMessages, sendMsgReq)

	token, err := s.generateExchangeToken(userID, userGroup.String(), applicantID, constants.ManabieSchool, s.ShamirConn)
	if err != nil {
		return ctx, err
	}

	resp, err := legacytpb.NewChatServiceClient(s.Conn).SendMessage(contextWithToken(context.Background(), token), sendMsgReq)
	if err != nil {
		return ctx, err
	}

	messageID := resp.MessageId
	if userGroup == cpb.UserGroup_USER_GROUP_STUDENT {
		s.studentMessageID = messageID
	} else {
		s.messageID = messageID
	}

	return ctx, err
}

func (s *suite) teacherSeesStatusCallingLiveLessonConversationDetail(ctx context.Context, seenStatus string) (context.Context, error) {
	tok, err := s.genTeacherToken(s.LessonChatState.firstTeacher)
	if err != nil {
		return ctx, err
	}
	return s.userWithTokenSeesStatusCallingLiveLessonConversationDetail(ctx, seenStatus, tok)
}
func (s *suite) teacherSeesEmptyLatestMessageCallingLiveLessonConversationDetail(ctx context.Context) (context.Context, error) {
	req := tpb.LiveLessonConversationDetailRequest{
		LessonId: s.LessonChatState.lessonID,
	}
	tok, err := s.genTeacherToken(s.firstTeacher)
	if err != nil {
		return ctx, err
	}

	res, err := tpb.NewLessonChatReaderServiceClient(s.Conn).LiveLessonConversationDetail(
		contextWithToken(context.Background(), tok),
		&req,
	)
	if err != nil {
		return ctx, err
	}
	if res.Conversation.LastMessage != nil {
		return ctx, fmt.Errorf("still receive non-nil latest message")
	}
	return ctx, nil
}

func (s *suite) theInLessonReceiveSilentNotificationWithContent(ctx context.Context, person string, msgContent string) (context.Context, error) {
	userIDs := []string{}
	role := ""
	switch person {
	case "students":
		userIDs = append(userIDs, s.LessonChatState.studentsInLesson...)
		role = cpb.UserGroup_USER_GROUP_STUDENT.String()
	case "second teacher":
		userIDs = append(userIDs, s.LessonChatState.secondTeacher)
		role = cpb.UserGroup_USER_GROUP_TEACHER.String()
	default:
		panic(fmt.Errorf("unsupported user receiving noti: %s", person))
	}
	for _, id := range userIDs {
		ctx, err := s.userWithIDAndRoleReceivesNotificationFromConversation(ctx, id, role, true, tpb.ConversationType_CONVERSATION_LESSON.String())
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *suite) aSecondTeacherJoinsLessonButNotSubscribeStream(ctx context.Context) (context.Context, error) {
	ctx, err := godogutil.MultiErrChain(
		ctx,
		s.aEvtLessonWithMessage, "JoinLesson",
		s.aValidIDInJoinLesson, cpb.UserGroup_USER_GROUP_TEACHER.String(),
		s.bobSendEventEvtLesson,
		s.tomAddAboveUserToThisLessonConversation, "must",
	)
	if err != nil {
		return ctx, err
	}

	s.LessonChatState.secondTeacher = s.teacherID
	return ctx, nil
}
