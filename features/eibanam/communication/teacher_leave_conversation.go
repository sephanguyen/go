package communication

import (
	"context"
	"fmt"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/internal/golibs/try"
	tpb "github.com/manabie-com/backend/pkg/manabuf/tom/v1"

	"github.com/cucumber/godog"
)

type TeacherLeaveConversationSuite struct {
	*SupportChatSuite
}

func NewTeacherLeaveConversationSuite(util *helper.CommunicationHelper) *TeacherLeaveConversationSuite {
	support := &SupportChatSuite{
		util: util,
	}
	return &TeacherLeaveConversationSuite{
		support,
	}
}

func (s *TeacherLeaveConversationSuite) InitScenario(ctx *godog.ScenarioContext) {
	stepsMapping := map[string]interface{}{
		`^"([^"]*)" logins CMS$`:                                                             s.LoginCms,
		`^"([^"]*)" logins Learner App$`:                                                     s.CreateAndLoginLeanerApp,
		`^"([^"]*)" login Teacher App$`:                                                      s.CreateMultipleLoginedTeacher,
		`^"([^"]*)" of "([^"]*)" logins Learner App$`:                                        s.ParentOfStudentLoginsLearnerApp,
		`^"([^"]*)" is at the conversation screen on Teacher App$`:                           s.IsAtTheConversationScreenOnTeacherApp,
		`^"([^"]*)" is at the conversation screen on Learner App$`:                           s.IsAtTheConversationScreenOnLearnerApp,
		`^"([^"]*)" joined "([^"]*)" group chat and "([^"]*)" group chat successfully$`:      s.JoinedMultipleChatGroup,
		`^"([^"]*)" has accessed to the conversation of "([^"]*)" chat group$`:               s.HasAccessedToTheConversationOf,
		`^"([^"]*)" sends "([^"]*)" to the conversation on Teacher App$`:                     s.SendsToTheConversationOnTeacherApp,
		`^"([^"]*)" leaves current chat group$`:                                              s.TeacherLeavesCurrentChatGroup,
		`^"([^"]*)" sees "([^"]*)" leave current chat group successfully$`:                   s.SeesTeacherLeaveCurrentChatGroupSuccessfully,
		`^"([^"]*)" refreshes and sees "([^"]*)" message of "([^"]*)" with name and avatar$`: s.RefreshesSeesMessageFromWithNameAndAvatar,
		`^"([^"]*)" sees "([^"]*)" message of "([^"]*)" with name and avatar$`:               s.SeesMessageFromWithNameAndAvatar,
		`^"([^"]*)" leaves "([^"]*)" chat group$`:                                            s.TeacherLeavesChatGroup,
		`^"([^"]*)" joins "([^"]*)" chat group$`:                                             s.JoinedChatGroup,
	}
	for pattern, function := range stepsMapping {
		ctx.Step(pattern, function)
	}
}

func (s *SupportChatSuite) JoinedGroupChatGroupChatSuccessfully(ctx context.Context, teacherName string, person1 string) (context.Context, error) {
	st := s.FromCtx(ctx)
	teacherSt := st.getTargetUserState(teacherName)
	chatList, err := s.util.ListSupportUnjoinedChat(ctx, teacherSt.Token)
	if err != nil {
		return ctx, err
	}
	p1id := st.getTargetUserState(person1).ID
	tojoined := []string{}

	for _, item := range chatList.GetItems() {
		for _, u := range item.GetUsers() {
			if u.Id == p1id {
				tojoined = append(tojoined, item.GetConversationId())
			}
		}
	}
	for _, conv := range tojoined {
		err = s.util.JoinSupportChatGroup(ctx, teacherSt.Token, conv)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}

func (s *TeacherLeaveConversationSuite) TeacherLeavesChatGroup(ctx context.Context, teacher1 string, chatGroupOf string) (context.Context, error) {
	ctx, err := s.HasAccessedToTheConversationOf(ctx, teacher1, chatGroupOf)
	if err != nil {
		return ctx, err
	}
	st := s.SupportChatSuite.FromCtx(ctx)
	chatSt := st.getTargetState(teacher1)
	userSt := st.getTargetUserState(teacher1)

	targetChat := chatSt.CurrentChatID
	err = s.util.LeaveSupportChatGroup(ctx, userSt.Token, targetChat)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *TeacherLeaveConversationSuite) SeesTeacherLeaveCurrentChatGroupSuccessfully(ctx context.Context, teacher1 string, teacher2 string) (context.Context, error) {
	st := s.SupportChatSuite.FromCtx(ctx)
	chatSt := st.getTargetState(teacher1)
	var (
		msg   helper.Message
		found bool
	)
	for try := 0; try < 5; try++ {
		msg2, err := s.util.DrainMsgFromStream(chatSt.Stream)
		if err != nil {
			return ctx, err
		}
		err = s.util.CheckMsgType("system", msg2.Type)
		if err != nil {
			continue
		}
		if msg2.Content != tpb.CodesMessageType_CODES_MESSAGE_TYPE_LEAVE_CONVERSATION.String() {
			continue
		}
		msg = msg2
		found = true
		break
	}
	if !found {
		return ctx, fmt.Errorf("cannot receive leaving system msg")
	}
	teacher2id := st.getTargetUserState(teacher2).ID
	if msg.TargetUser != teacher2id {
		return ctx, fmt.Errorf("want target user %s, has %s", teacher2id, msg.TargetUser)
	}
	return ctx, nil
}

func (s *TeacherLeaveConversationSuite) TeacherLeavesCurrentChatGroup(ctx context.Context, teacherName string) (context.Context, error) {
	st := s.SupportChatSuite.FromCtx(ctx)
	chatSt := st.getTargetState(teacherName)
	userSt := st.getTargetUserState(teacherName)
	targetChat := chatSt.CurrentChatID
	err := s.util.LeaveSupportChatGroup(ctx, userSt.Token, targetChat)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *TeacherLeaveConversationSuite) SendsToTheConversationOnTeacherApp(ctx context.Context, teacherName string, msgtype string) (context.Context, error) {
	st := s.SupportChatSuite.FromCtx(ctx)
	chatSt := st.getTargetState(teacherName)
	userSt := st.getTargetUserState(teacherName)
	targetChat := chatSt.CurrentChatID
	ctx, msgtype = s.selectOneOf(ctx, msgtype)
	err := s.util.SendMsgToConversation(ctx, userSt.Token, msgtype, targetChat)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (s *TeacherLeaveConversationSuite) HasAccessedToTheConversationOf(ctx context.Context, teacherName string, targetperson string) (context.Context, error) {
	st := s.SupportChatSuite.FromCtx(ctx)
	teacherSt := st.getTargetUserState(teacherName)
	var targetChat string
	err := try.Do(func(attempt int) (bool, error) {
		chatList, err := s.util.ListSupportJoinedChat(ctx, teacherSt.Token)
		if err != nil {
			return false, err
		}

		target := st.getTargetUserState(targetperson).ID
		var (
			found bool
		)
		for _, item := range chatList.GetItems() {
			for _, u := range item.GetUsers() {
				if u.Id == target {
					targetChat = item.GetConversationId()
					found = true
					break
				}
			}
		}
		if !found {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("cannot find %s chat", target)
		}
		return false, nil
	})
	if err != nil {
		return ctx, err
	}

	st.getTargetState(teacherName).CurrentChatID = targetChat
	return s.SupportChatSuite.ToCtx(ctx, st), nil
}

func (s *TeacherLeaveConversationSuite) JoinedMultipleChatGroup(ctx context.Context, teacherName string, person1, person2 string) (context.Context, error) {
	ctx, err := s.SupportChatSuite.JoinedChatGroup(ctx, teacherName, person1)
	if err != nil {
		return ctx, err
	}
	ctx, err = s.SupportChatSuite.JoinedChatGroup(ctx, teacherName, person2)
	return ctx, err
}

func (s *TeacherLeaveConversationSuite) IsAtTheConversationScreenOnLearnerApp(ctx context.Context, targetName string) (context.Context, error) {
	st := s.SupportChatSuite.FromCtx(ctx)
	targets := parseTargetUsers(targetName)
	for _, target := range targets {
		chatSt := st.getTargetState(target)
		userst := st.getTargetUserState(target)
		stream, err := s.util.ConnectChatStream(ctx, userst.Token)
		if err != nil {
			return ctx, err
		}
		chatSt.Stream = stream
	}
	return s.SupportChatSuite.ToCtx(ctx, st), nil
}

func (s *TeacherLeaveConversationSuite) IsAtTheConversationScreenOnTeacherApp(ctx context.Context, teacherNames string) (context.Context, error) {
	st := s.SupportChatSuite.FromCtx(ctx)
	targetTeachers := parseTargetUsers(teacherNames)
	for _, target := range targetTeachers {
		chatSt := st.getTargetState(target)
		userst := st.getTargetUserState(target)
		stream, err := s.util.ConnectChatStream(ctx, userst.Token)
		if err != nil {
			return ctx, err
		}
		chatSt.Stream = stream
	}
	return s.SupportChatSuite.ToCtx(ctx, st), nil
}

func (s *TeacherLeaveConversationSuite) ParentOfStudentLoginsLearnerApp(ctx context.Context, parent string, studentName string) (context.Context, error) {
	return s.LoginLeanerApp(ctx, parent)
}
func (s *TeacherLeaveConversationSuite) CreateAndLoginLeanerApp(ctx context.Context, studentName string) (context.Context, error) {
	ctx, err := s.CreateStudentWithParentInfo(ctx, "school admin", studentName, "parent S1")
	if err != nil {
		return ctx, err
	}
	ctx, err = s.LoginLeanerApp(ctx, studentName)
	if err != nil {
		return ctx, err
	}
	return ctx, nil
}
func (s *TeacherLeaveConversationSuite) CreateMultipleLoginedTeacher(ctx context.Context, teachers string) (context.Context, error) {
	targetTeachers := parseTargetUsers(teachers)
	for _, t := range targetTeachers {
		ctx2, err := s.CreateLoginedTeacher(ctx, t)
		if err != nil {
			return ctx, err
		}
		ctx = ctx2
	}
	return ctx, nil
}
