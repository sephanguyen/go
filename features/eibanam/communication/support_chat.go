package communication

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/manabie-com/backend/features/eibanam/communication/entity"
	"github.com/manabie-com/backend/features/eibanam/communication/helper"
	"github.com/manabie-com/backend/features/eibanam/communication/util"
	"github.com/manabie-com/backend/internal/golibs/try"
	legacytpb "github.com/manabie-com/backend/pkg/genproto/tom"
)

type SupportChatSuite struct {
	util *helper.CommunicationHelper
}

func (s *SupportChatSuite) SchoolAdminAddsTheRelationshipOfStudentParent(ctx context.Context, parentName, studentName string) (context.Context, error) {
	st := s.FromCtx(ctx)
	utilSt := util.StateFromContext(ctx)
	// support one student for now
	student := st.getTargetState(studentName)
	par := st.getTargetUserState(parentName)

	student.Student.Parents = append(student.Student.Parents, par)

	err := s.util.UpdateStudentWithParent(utilSt.School.Admins[0], student.Student)
	if err != nil {
		return ctx, err
	}

	return s.ToCtx(ctx, st), err
}

func (s *SupportChatSuite) SchoolAdminRemovesTheRelationshipOfStudentParent(ctx context.Context, parentName, studentName string) (context.Context, error) {
	st := s.FromCtx(ctx)
	utilSt := util.StateFromContext(ctx)
	// support one student for now
	student := st.getTargetState(studentName)
	par := st.getTargetUserState(parentName)

	newpar := []*entity.User{}
	for _, parent := range student.Student.Parents {
		if parent.ID == par.ID {
			continue
		}
		newpar = append(newpar, parent)
	}
	student.Student.Parents = newpar
	err := s.util.RemoveParentFromStudent(utilSt.School.Admins[0], student.Student, par)
	return s.ToCtx(ctx, st), err
}

func (s *SupportChatSuite) LoginCms(ctx context.Context, accountName string) (context.Context, error) {
	state := util.StateFromContext(ctx)
	switch accountName {
	case "school admin":
		sysAdmin, school, err := s.util.CreateSchoolAdminAndLoginToCMS(ctx, helper.AccountTypeSchoolAdmin)
		if err != nil {
			return ctx, err
		}
		state.SystemAdmin = sysAdmin
		state.School = school
	default:
		return ctx, fmt.Errorf("only support school admin for now")
	}
	return util.StateToContext(ctx, state), nil
}

type SupportChatState struct {
	Student1 *UserChatState
	Student2 *UserChatState
	Parent1  *UserChatState
	Parent2  *UserChatState
	Teacher1 *UserChatState
	Teacher2 *UserChatState
}
type UserChatState struct {
	// oneof
	Student *entity.Student
	Parent  *entity.User
	Teacher *entity.Teacher

	LearnerAppChats *legacytpb.ConversationListResponse
	CurrentChatID   string
	Stream          legacytpb.ChatService_SubscribeV2Client
}

type supportChatStateKey struct{}

// For less typing only
func (s *SupportChatSuite) FromCtx(ctx context.Context) *SupportChatState {
	return SupportChatStateFromCtx(ctx)
}

// For less typing only
func (s *SupportChatSuite) ToCtx(ctx context.Context, st *SupportChatState) context.Context {
	return SupportChatStateToCtx(ctx, st)
}

func SupportChatStateFromCtx(ctx context.Context) *SupportChatState {
	st := ctx.Value(supportChatStateKey{})
	if st == nil {
		return &SupportChatState{}
	}
	return st.(*SupportChatState)
}
func SupportChatStateToCtx(ctx context.Context, st *SupportChatState) context.Context {
	return context.WithValue(ctx, supportChatStateKey{}, st)
}

func (s *SupportChatSuite) CreateLoginedTeacher(ctx context.Context, person string) (context.Context, error) {
	utilSt := util.StateFromContext(ctx)
	teacher, err := s.util.CreateNewTeacher(utilSt.School.Admins[0], int64(utilSt.School.ID))
	if err != nil {
		return ctx, err
	}
	token, err := s.util.LoginLeanerApp(teacher.User.Email, teacher.User.Password)
	if err != nil {
		return ctx, err
	}
	st := s.FromCtx(ctx)

	teacher.User.Token = token
	st.setTargetState(person, UserChatState{
		Teacher: teacher,
	})

	return s.ToCtx(ctx, st), nil
}

func (s *SupportChatState) getTargetUserState(target string) *entity.User {
	switch target {
	case "student":
		return &s.Student1.Student.User
	case "student S1":
		return &s.Student1.Student.User
	case "student S2":
		return &s.Student2.Student.User
	case "parent P1":
		return s.Parent1.Parent
	case "parent P2":
		return s.Parent2.Parent
	case "teacher":
		return s.Teacher1.Teacher.User
	case "teacher T1":
		return s.Teacher1.Teacher.User
	case "teacher T2":
		return s.Teacher2.Teacher.User
	default:
		panic(fmt.Sprintf("unknown target %s", target))
	}
}
func (s *SupportChatState) setTargetState(target string, st UserChatState) {
	switch target {
	case "student":
		s.Student1 = &st
	case "student S1":
		s.Student1 = &st
	case "student S2":
		s.Student2 = &st
	case "parent P1":
		s.Parent1 = &st
	case "parent P2":
		s.Parent2 = &st
	case "teacher":
		s.Teacher1 = &st
	case "teacher T1":
		s.Teacher1 = &st
	case "teacher T2":
		s.Teacher2 = &st
	default:
		panic(fmt.Sprintf("unknown target %s", target))
	}
}

func (s *SupportChatState) getTargetState(target string) *UserChatState {
	switch target {
	case "student":
		return s.Student1
	case "student S1":
		return s.Student1
	case "student S2":
		return s.Student2
	case "parent P1":
		return s.Parent1
	case "parent P2":
		return s.Parent2
	case "teacher":
		return s.Teacher1
	case "teacher T1":
		return s.Teacher1
	case "teacher T2":
		return s.Teacher2
	default:
		panic(fmt.Sprintf("unknown target %s", target))
	}
}

type oneofmsgtypekey struct{}

func (s *SupportChatSuite) selectOneOf(ctx context.Context, oneOf string) (context.Context, string) {
	if val := ctx.Value(oneofmsgtypekey{}); val != nil {
		return ctx, val.(string)
	}
	msgtype := selectOneOf(oneOf)
	ctx = context.WithValue(ctx, oneofmsgtypekey{}, msgtype)
	return ctx, msgtype
}

func (s *SupportChatSuite) SendMessageToParentChatGroup(ctx context.Context, person string, msgtype string) (context.Context, error) {
	if !strings.HasPrefix(person, "parent") {
		return ctx, fmt.Errorf("only support parent send msg for now")
	}
	st := s.FromCtx(ctx)

	ctx, msgtype = s.selectOneOf(ctx, msgtype)

	chatState := st.getTargetState(person)
	userState := st.getTargetUserState(person)

	convID := chatState.LearnerAppChats.Conversations[0].GetConversationId()
	err := try.Do(func(attempt int) (bool, error) {
		err := s.util.SendMsgToConversation(ctx, userState.Token, msgtype, convID)
		if err != nil {
			time.Sleep(2 * time.Second)
			return attempt < 5, err
		}
		return false, nil
	})
	return s.ToCtx(ctx, st), err
}

func (s *SupportChatSuite) LoginLeanerApp(ctx context.Context, people string) (context.Context, error) {
	st := s.FromCtx(ctx)
	users := parseTargetUsers(people)
	for _, u := range users {
		userState := st.getTargetUserState(u)
		chatState := st.getTargetState(u)
		token, err := s.util.LoginLeanerApp(userState.Email, userState.Password)
		if err != nil {
			return ctx, err
		}
		userState.Token = token
		chats, err := s.util.GetLearnerAppChat(ctx, token)
		if err != nil {
			return ctx, fmt.Errorf("cannot list chat in learner app: %w", err)
		}
		chatState.LearnerAppChats = chats
		stream, err := s.util.ConnectChatStream(ctx, token)
		if err != nil {
			return ctx, err
		}
		chatState.Stream = stream
	}
	return s.ToCtx(ctx, st), nil
}

// func (s *SupportChatSuite) TeacherSeesMessageOfWithNameAndAvatar()

func (s *SupportChatSuite) JoinedChatGroupAndSubscribeStream(ctx context.Context, person string, chatGroupOf string) (context.Context, error) {
	ctx, err := s.JoinedChatGroup(ctx, person, chatGroupOf)
	if err != nil {
		return ctx, err
	}
	st := s.FromCtx(ctx)
	user := st.getTargetUserState("teacher")
	stream, err := s.util.ConnectChatStream(ctx, user.Token)
	if err != nil {
		return ctx, err
	}
	userChatSt := st.getTargetState("teacher")
	userChatSt.Stream = stream
	return s.ToCtx(ctx, st), nil
}

func (s *SupportChatSuite) JoinedChatGroup(ctx context.Context, person string, chatGroupOf string) (context.Context, error) {
	st := s.FromCtx(ctx)
	chatState := st.getTargetState(person)

	if chatState == nil {
		ctx, err := s.CreateLoginedTeacher(ctx, person)
		if err != nil {
			return ctx, err
		}
	}
	teacherUser := st.getTargetUserState(person)

	chatGroupOwner := st.getTargetUserState(chatGroupOf)

	var convID string
	err := try.Do(func(attempt int) (bool, error) {
		list, err := s.util.ListSupportUnjoinedChat(ctx, teacherUser.Token)
		if err != nil {
			return false, err
		}
		for _, item := range list.GetItems() {
			for _, u := range item.GetUsers() {
				if u.GetId() == chatGroupOwner.ID {
					convID = item.GetConversationId()
				}
			}
		}
		if convID == "" {
			time.Sleep(2 * time.Second)
			return attempt < 5, fmt.Errorf("cannot find chat of %s in unjoined list", chatGroupOf)
		}
		return false, nil
	})
	if err != nil {
		return ctx, err
	}

	err = s.util.JoinSupportChatGroup(ctx, teacherUser.Token, convID)
	if err != nil {
		return ctx, err
	}
	return s.ToCtx(ctx, st), nil
}

func parseTargetUsers(str string) []string {
	temp := strings.Split(str, ",")
	var ret = make([]string, 0, len(temp))
	for _, item := range temp {
		ret = append(ret, strings.TrimSpace(item))
	}
	return ret
}

func (s *SupportChatSuite) RefreshesSeesMessageFromWithNameAndAvatar(ctx context.Context, msgReceiver, msgType, msgSender string) (context.Context, error) {
	st := s.FromCtx(ctx)
	state := st.getTargetState(msgReceiver)
	userState := st.getTargetUserState(msgReceiver)
	if state.Stream == nil {
		return ctx, fmt.Errorf("%s is not yet connected to stream", msgReceiver)
	}
	ctx, msgType = s.selectOneOf(ctx, msgType)
	msges, err := s.util.ListConversationMessages(ctx, userState.Token, state.CurrentChatID)
	if err != nil {
		return ctx, err
	}
	var (
		msg   helper.Message
		found bool
	)
	for idx := range msges {
		item := msges[idx]
		err = s.util.CheckMsgType(msgType, item.Type)
		if err != nil {
			continue
		}
		found = true
		msg = item
		break
	}

	if !found {
		return ctx, fmt.Errorf("not found msg type %s", msgType)
	}

	senderUserState := st.getTargetUserState(msgSender)
	if msg.Sender != senderUserState.ID {
		return ctx, fmt.Errorf("want sender id is %s, has %s", userState.ID, msg.Sender)
	}

	return ctx, nil
}

func (s *SupportChatSuite) SeesMessageFromWithNameAndAvatar(ctx context.Context, msgReceiver, msgType, msgSender string) (context.Context, error) {
	st := s.FromCtx(ctx)
	state := st.getTargetState(msgReceiver)
	if state.Stream == nil {
		return ctx, fmt.Errorf("%s is not yet connected to stream", msgReceiver)
	}
	ctx, msgType = s.selectOneOf(ctx, msgType)
	var (
		msg   helper.Message
		found bool
	)
	for try := 0; try < 5; try++ {
		msg2, err := s.util.DrainMsgFromStream(state.Stream)
		if err != nil {
			return ctx, err
		}
		err = s.util.CheckMsgType(msgType, msg2.Type)
		if err != nil {
			continue
		}
		found = true
		msg = msg2
		break
	}
	if !found {
		return ctx, fmt.Errorf("not found msg type %s", msgType)
	}

	userState := st.getTargetUserState(msgSender)
	if msg.Sender != userState.ID {
		return ctx, fmt.Errorf("want sender id is %s, has %s", userState.ID, msg.Sender)
	}

	return ctx, nil
}

func (s *SupportChatSuite) TeacherSeesMessageFromWithNameAndAvatar(ctx context.Context, msgType, msgSender string) (context.Context, error) {
	st := s.FromCtx(ctx)
	ctx, msgType = s.selectOneOf(ctx, msgType)
	state := st.getTargetState("teacher")
	var found bool
	var msg helper.Message
	for try := 0; try < 5; try++ {
		msg2, err := s.util.DrainMsgFromStream(state.Stream)
		if err != nil {
			return ctx, err
		}
		err = s.util.CheckMsgType(msgType, msg2.Type)
		if err != nil {
			continue
		}
		found = true
		msg = msg2
		break
	}
	if !found {
		return ctx, fmt.Errorf("not found msg type %s", msgType)
	}

	userState := st.getTargetUserState(msgSender)
	if msg.Sender != userState.ID {
		return ctx, fmt.Errorf("want sender id is %s, has %s", userState.ID, msg.Sender)
	}

	return s.ToCtx(ctx, st), nil
}

func (s *SupportChatSuite) CreateStudentWithParentInfo(ctx context.Context, adminName string, studentName string, parentNamesStr string) (context.Context, error) {
	mainSt := s.FromCtx(ctx)

	state := util.StateFromContext(ctx)
	if state.School == nil || len(state.School.Admins) == 0 {
		return ctx, errors.New("missing create school and admin step")
	}
	parents := parseTargetUsers(parentNamesStr)

	switch adminName {
	case "school admin":
		// create student
		newStudent, err := s.util.CreateStudent(state.School.Admins[0], 4, []string{state.School.DefaultLocation}, true, len(parents))
		if err != nil {
			return ctx, err
		}

		courses, err := s.util.CreateCourses(state.School.Admins[0], state.School.ID, newStudent.Grade.ID, 1)
		if err != nil {
			return ctx, err
		}

		if err = s.util.AddCourseToStudent(state.School.Admins[0], newStudent, courses); err != nil {
			return ctx, err
		}

		state.Students = append(state.Students, newStudent)
		mainSt.setTargetState(studentName, UserChatState{
			Student: newStudent,
		})
		// mainSt.CurrentStudent = &UserChatState{
		// 	Student: newStudent,
		// }
		switch len(parents) {
		case 2:
			mainSt.Parent2 = &UserChatState{
				Parent: newStudent.Parents[1],
			}
			fallthrough
		case 1:
			mainSt.Parent1 = &UserChatState{
				Parent: newStudent.Parents[0],
			}
		}
		ctx = s.ToCtx(ctx, mainSt)
	default:
		return ctx, fmt.Errorf("%s doesn't have to create new student", adminName)

	}

	return util.StateToContext(ctx, state), nil
}
