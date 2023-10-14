package communication

import (
	"context"
	"fmt"
	"time"

	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
	bpb "github.com/manabie-com/backend/pkg/manabuf/bob/v1"
	cpb "github.com/manabie-com/backend/pkg/manabuf/common/v1"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"
	"google.golang.org/protobuf/types/known/durationpb"
)

var (
	sendMsgTypeImage = "image"
	sendMsgTypePdf   = "pdf"
	sendMsgTypeText  = "text"

	sendMsgTypeHyperlink = "hyperlink"
)

// can't simulate
func (s *suite) redirectsToWebBrowser(person string) error {
	return nil
}

func (s *suite) sendsToTheConversationOnLearnerApp(userAccount string, msgType string) error {
	userAccount = s.loadFromCacheIfIsOneOfSyntax("user", userAccount)
	msgType = s.loadFromCacheIfIsOneOfSyntax("msgType", msgType)
	chosenChat := ""
	switch userAccount {
	case student:
		chosenChat = s.studentChatState.studentChats.GetConversations()[0].ConversationId
	case parent:
		chosenChat = s.studentChatState.parentChats.GetConversations()[0].ConversationId
	default:
		return fmt.Errorf("unsupported user %s", userAccount)
	}
	if chosenChat == "" {
		return fmt.Errorf("not found chat for account %s", userAccount)
	}
	req := &legacytpb.SendMessageRequest{
		ConversationId: chosenChat,
	}

	switch msgType {
	case sendMsgTypeText:
		req.Message = "hello world"
		req.Type = legacytpb.MESSAGE_TYPE_TEXT
	case sendMsgTypeHyperlink:
		req.Message = "https://google.com"
		req.Type = legacytpb.MESSAGE_TYPE_TEXT
	case sendMsgTypeImage: // mobile is treating image == file
		req.Type = legacytpb.MESSAGE_TYPE_FILE
		url, err := s.generateURLforFile(sendMsgTypeImage)
		if err != nil {
			return err
		}
		req.Message = fmt.Sprintf("<a href=\"%s\">%s</a>", url, "testimage.jpg")
	case sendMsgTypePdf:
		req.Type = legacytpb.MESSAGE_TYPE_FILE
		url, err := s.generateURLforFile(sendMsgTypePdf)
		if err != nil {
			return err
		}
		req.Message = fmt.Sprintf("<a href=\"%s\">%s</a>", url, "testpdf.pdf")
	default:
		return fmt.Errorf("unsupported msg type %s", msgType)
	}

	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(userAccount))
	defer cancel()
	_, err := legacytpb.NewChatServiceClient(s.tomConn).SendMessage(ctx, req)
	if err != nil {
		return err
	}
	newMsg := message{
		msgType: req.Type.String(),
		content: req.Message,
	}
	newMsg.setSeen(userAccount)
	s.studentChatState.newMessageBuffers[chosenChat] = append(s.studentChatState.newMessageBuffers[chosenChat], newMsg)
	return nil
}

func (s *suite) teacherHasAccessedToTheConversationOfChatGroup(userAccount string) error {
	userAccount = s.loadFromCacheIfIsOneOfSyntax("user", userAccount)
	err := s.teacherIsAtTheConversationScreen()
	if err != nil {
		return err
	}

	chatID := ""
	var expectedType tpb.ConversationType
	switch userAccount {
	case student:
		expectedType = tpb.ConversationType_CONVERSATION_STUDENT
	case parent:
		expectedType = tpb.ConversationType_CONVERSATION_PARENT
	default:
		return fmt.Errorf("not expect teacher to access %s chat in teacher screen", userAccount)
	}
	for _, chat := range s.studentChatState.teacherChats.GetItems() {
		if chat.ConversationType == expectedType &&
			chat.StudentId == s.profile.defaultStudent.id {
			chatID = chat.ConversationId
			break
		}
	}
	if chatID == "" {
		return fmt.Errorf("not found chat for %s in teacher conversation screen", userAccount)
	}
	s.studentChatState.teacherChosenChat = chatID
	s.studentChatState.teacherChosenChatType = userAccount

	return nil
}

// TODO: replace this function and start using connectV2Stream supporting context chaining
func (s *suite) personSubscribeToChat(person string) error {
	chatSvc := legacytpb.NewChatServiceClient(s.tomConn)
	ctx := contextWithChatHashKey(contextWithToken(context.Background(), s.getToken(person)), s.getID(person))
	streamV2, err := chatSvc.SubscribeV2(ctx, &legacytpb.SubscribeV2Request{})
	if err != nil {
		return err
	}

	sessionID := ""
	resp, err := streamV2.Recv()
	if err != nil {
		return err
	}

	if resp.Event.GetEventPing() == nil {
		return fmt.Errorf("stream must receive pingEvent first")
	}

	sessionID = resp.Event.GetEventPing().SessionId
	token := s.getToken(person)
	switch person {
	case teacher:
		s.studentChatState.teacherChannel = streamV2
	case student:
		s.studentChatState.studentChannel = streamV2
	case parent:
		s.studentChatState.parentChannel = streamV2
	default:
		return fmt.Errorf("does not support person %s to subscribe stream v2", person)
	}

	go func() {
		for {
			ctx, cancel := contextWithTokenAndTimeOut(context.Background(), token)
			defer cancel()
			_, err := chatSvc.PingSubscribeV2(ctx, &legacytpb.PingSubscribeV2Request{SessionId: sessionID})
			if err != nil {
				zapLogger.Error(fmt.Sprintf("failed to ping using subscribe v2 api: %s\n", err))
				return
			}

			time.Sleep(2 * time.Second)
		}
	}()

	return nil
}

func (s *suite) generateURLforFile(fileType string) (string, error) {
	resumableUpload := &bpb.ResumableUploadURLRequest{
		AllowOrigin:   "http://localhost:3001",
		ContentType:   "",
		Expiry:        durationpb.New(time.Second * 1800),
		FileExtension: "",
		PrefixName:    "",
	}
	switch fileType {
	case sendMsgTypeImage:
		resumableUpload.FileExtension = "jpg"
		resumableUpload.PrefixName = "testimage"
	case sendMsgTypePdf:
		resumableUpload.FileExtension = sendMsgTypePdf
		resumableUpload.PrefixName = "testpdf"
	default:
		return "", fmt.Errorf("unsupported file type %s", fileType)
	}

	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
	defer cancel()
	res, err := bpb.NewUploadServiceClient(s.bobConn).GenerateResumableUploadURL(ctx, resumableUpload)
	if err != nil {
		return "", err
	}
	return res.ResumableUploadUrl, nil
}

func (s *suite) teacherSendsToTheConversationOnTeacherApp(msgType string) error {
	msgType = s.loadFromCacheIfIsOneOfSyntax("msgType", msgType)
	req := &legacytpb.SendMessageRequest{
		ConversationId: s.studentChatState.teacherChosenChat,
	}

	switch msgType {
	case sendMsgTypeText:
		req.Message = "hello world"
		req.Type = legacytpb.MESSAGE_TYPE_TEXT
	case sendMsgTypeHyperlink:
		req.Message = "https://google.com"
		req.Type = legacytpb.MESSAGE_TYPE_TEXT
	case sendMsgTypeImage: // mobile is treating image == file
		req.Type = legacytpb.MESSAGE_TYPE_FILE
		url, err := s.generateURLforFile(sendMsgTypeImage)
		if err != nil {
			return err
		}
		req.Message = fmt.Sprintf("<a href=\"%s\">%s</a>", url, "testimage.jpg")
	case sendMsgTypePdf:
		req.Type = legacytpb.MESSAGE_TYPE_FILE
		url, err := s.generateURLforFile(sendMsgTypePdf)
		if err != nil {
			return err
		}
		req.Message = fmt.Sprintf("<a href=\"%s\">%s</a>", url, "testpdf.pdf")
	default:
		return fmt.Errorf("unsupported msg type %s", msgType)
	}

	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
	defer cancel()
	_, err := legacytpb.NewChatServiceClient(s.tomConn).SendMessage(ctx, req)
	if err != nil {
		return err
	}
	chatID := s.studentChatState.teacherChosenChat
	s.studentChatState.newMessageBuffers[chatID] = append(s.studentChatState.newMessageBuffers[chatID], message{
		msgType:       req.Type.String(),
		content:       req.Message,
		seenByTeacher: false,
	})
	return nil
}

func (s *suite) isAtTheConversationScreen(person string) error {
	if person == teacher {
		return s.teacherIsAtTheConversationScreen()
	}

	req := &legacytpb.ConversationListRequest{
		Limit: 100,
	}

	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(person))
	defer cancel()
	res, err := legacytpb.NewChatServiceClient(s.tomConn).ConversationList(ctx, req)
	if err != nil {
		return err
	}
	switch person {
	case parent:
		s.studentChatState.parentChats = res
	case student:
		s.studentChatState.studentChats = res
	default:
		return fmt.Errorf("not expect person %s to access conversation screen", person)
	}
	err = s.personSubscribeToChat(person)
	if err != nil {
		return err
	}
	return nil
}

func (s *suite) teacherIsAtTheConversationScreen() error {
	req := &tpb.ListConversationsInSchoolRequest{
		Paging: &cpb.Paging{
			Limit: 100,
		},
	}

	ctx, cancel := contextWithTokenAndTimeOut(context.Background(), s.getToken(teacher))
	defer cancel()
	res, err := tpb.NewChatReaderServiceClient(s.tomConn).ListConversationsInSchool(ctx, req)
	if err != nil {
		return err
	}
	s.studentChatState.teacherChats = res
	err = s.personSubscribeToChat(teacher)
	if err != nil {
		return err
	}
	return nil
}
